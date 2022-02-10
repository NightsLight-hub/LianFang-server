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
	"github.com/sxy/lianfang/pkg/util"
	"io"
	"net/http"
)

func SetupWsRouter(engine *gin.RouterGroup) {
	engine.GET("/container/ssh/:cid", containerSSH)
}

func containerSSH(c *gin.Context) {
	cid := c.Param("cid")
	logrus.Debugf("in container ssh, cid is %s", cid)
	err := cri.ValidContainer(cid)
	if err != nil {
		if util.IsResourceNotExistError(err) {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrResponse{Msg: err.Error()})
		}
		logrus.Error("%+v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
	}
	logrus.Info("get /container/ssh ws connect")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Error(errors.WithStack(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
	}
	defer conn.Close()
	hr, err := exec(cid)
	if err != nil {
		logrus.Error("%+v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
	}
	// 关闭I/O流
	defer hr.Close()
	// 退出进程
	defer func() {
		hr.Conn.Write([]byte("exit\r"))
	}()

	// 转发输入/输出至websocket
	go func() {
		// 将伪终端的输出转发到前端
		wsWriterCopy(hr.Conn, conn)
	}()
	// 将前端输入拷贝到伪终端
	wsReaderCopy(conn, hr.Conn)
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
func wsWriterCopy(reader io.Reader, writer *websocket.Conn) {
	buf := make([]byte, 8192)
	for {
		nr, err := reader.Read(buf)
		if nr > 0 {
			err := writer.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// 将前端的输入转发到终端
func wsReaderCopy(reader *websocket.Conn, writer io.Writer) {
	for {
		messageType, p, err := reader.ReadMessage()
		if err != nil {
			return
		}
		if messageType == websocket.TextMessage {
			writer.Write(p)
		}
	}
}
