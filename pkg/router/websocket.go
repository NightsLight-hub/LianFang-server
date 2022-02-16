/*
Package router
@Time : 2022/2/9 14:13
@Author : sunxy
@File : websocket
@description:
*/
package router

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sxy/lianfang/pkg/common"
	"github.com/sxy/lianfang/pkg/cri"
	"io"
	"net/http"
)

const keyMsg string = "LianFangSshChannelPrepareFinished"

func SetupWsRouter(engine *gin.RouterGroup) {
	engine.GET("/container/ssh/:cid", containerSSH)
}

func containerSSH(c *gin.Context) {
	cid := c.Param("cid")
	logrus.Debugf("in container ssh, cid is %s", cid)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("%+v", errors.WithStack(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
	}
	defer conn.Close()
	err = cri.ValidContainer(cid)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	hr, err := exec(cid)
	if err != nil {
		logrus.Errorf("%+v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(keyMsg))
	if err != nil {
		logrus.Errorf("%+v", errors.Wrap(err, "Send key msg to frontend failed"))
		return
	}
	// 关闭I/O流
	defer hr.Close()
	// 退出进程
	defer func() {
		hr.Conn.Write([]byte("exit\r"))
	}()

	termStopCh := make(chan int)
	browserCloseCh := make(chan int)
	// 转发输入/输出至websocket
	go func() {
		// 将伪终端的输出转发到前端
		forwardExecToBrowser(hr.Conn, conn, termStopCh)
	}()

	go func() {
		// 将前端输入拷贝到伪终端
		forwardBrowserToExec(conn, hr.Conn, browserCloseCh)
	}()
	// 当浏览器或者term 中断， 关闭当前websocket
	select {
	case <-termStopCh:
		return
	case <-browserCloseCh:
		return
	}
}

// 开启container exec 并附加到/bin/sh上
func exec(container string) (hr types.HijackedResponse, err error) {
	// 执行/bin/bash命令
	ir, err := common.DockerClient().ContainerExecCreate(context.Background(), container, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		// todo alpine 没有/bin/bash吧
		Cmd: []string{"/bin/sh"},
		Tty: true,
	})
	if err != nil {
		return
	}

	// 附加到上面创建的/bin/bash进程中
	hr, err = common.DockerClient().ContainerExecAttach(context.Background(), ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		return
	}
	return
}

// 将终端的输出转发到前端
//  reader exec终端的连接reader
func forwardExecToBrowser(reader io.Reader, writer *websocket.Conn, flagCh chan int) {
	buf := make([]byte, 8192)
	for {
		nr, err := reader.Read(buf)
		if nr > 0 {
			err := writer.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if err != nil {
				logrus.Errorf("%+v", errors.Wrap(err, "write msg to browser failed"))
				flagCh <- 1
				return
			}
		}
		if err != nil {
			// 终端关闭
			if err != io.EOF {
				logrus.Errorf("%+v", errors.WithStack(err))
			}
			flagCh <- 1
			return
		}
	}
}

// 将前端的输入转发到终端
// 当前端输入
func forwardBrowserToExec(reader *websocket.Conn, writer io.Writer, flagCh chan int) {
	for {
		_, p, err := reader.ReadMessage()
		if err != nil {
			logrus.Errorf("%+v", errors.Wrap(err, "read browser msg failed"))
			flagCh <- 1
			return
		}
		_, err = writer.Write(p)
		if err != nil {
			logrus.Errorf("%+v", errors.Wrap(err, "write cmd to container exec failed"))
			flagCh <- 1
			return
		}
	}
}
