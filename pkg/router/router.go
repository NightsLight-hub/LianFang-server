/*
@Time : 2021/12/21 14:41
@Author : sunxy
@File : router
@description:
*/
package router

import (
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/system"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sxy/lianfang/pkg/cri"
	"github.com/sxy/lianfang/pkg/handler/containers"
	"github.com/sxy/lianfang/pkg/models"
	"github.com/sxy/lianfang/pkg/parser"
	"github.com/sxy/lianfang/pkg/util"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"
)

const (
	SKIP      string = "skip"
	OVERRIDE  string = "override"
	KEEP_BOTH string = "keep_both"
	// 最大两M可预览
	MaxFileReadSize = 2 << 20
)

func SetupContainersRouter(engine *gin.RouterGroup) {
	engine.GET("/containers", getContainers)
	engine.GET("/container/:cid/stats", getContainerStats)
	engine.GET("/container/show/:cid/*path", containerFsShow)
}

func getContainers(c *gin.Context) {
	cs, err := containers.GetService().List()
	if err != nil {

		errResp := new(ErrResponse)
		errResp.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server Error",
		})
	} else {
		c.JSON(http.StatusOK, cs)
	}
}

func getContainerStats(c *gin.Context) {
	cid := c.Param("cid")
	sts, err := containers.GetService().Stats(cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server Error",
		})
	} else {
		c.String(http.StatusOK, string(sts))
	}
}
func containerFsShow(c *gin.Context) {
	cid := c.Param("cid")
	path := c.Param("path")
	dirOnly, err := strconv.ParseBool(c.DefaultQuery("dirOnly", "false"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("dirOnly must be true or false!"))
	}
	path, err = url.PathUnescape(path)
	if err != nil {
		logrus.Warn("非法参数："+path, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrResponse{Msg: "不合法的参数"})
	}
	if err := antiPathInjection(path); err != nil {
		logrus.Warnf("%+v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrResponse{Msg: err.Error()})
	}
	logrus.Debug("in container show , cid : %s, path : %s", cid, path)
	pathStat, err := cri.GetContainerFileInfo(cid, path)
	if err != nil {
		logrus.Error("%+v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
	}
	// If the destination is a symbolic link, we should follow it.
	if pathStat.Mode&os.ModeSymlink != 0 {
		linkTarget := pathStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			srcParent, _ := archive.SplitPathDirEntry(path)
			linkTarget = filepath.Join(srcParent, linkTarget)
		}
		linkTarget, _ = archive.GetRebaseName(path, linkTarget)
		path = linkTarget
		// 为了能在windows下远程docker 容器调试，不得不来一次目录分隔符修正
		path = util.GetLinuxPath(path)
		pathStat, err = cri.GetContainerFileInfo(cid, path)
		if err != nil {
			logrus.Error("%+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
		}
	}
	if pathStat.Mode.IsDir() {
		var output string
		output, err = cri.StatCommand(cid, path, true)
		if err != nil {
			logrus.Error("container fs show failed err : %+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
		}
		resp, err := parser.ParseStatCmdOutput(output, path, dirOnly)
		if err != nil {
			logrus.Error("podFs show failed err : %+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
		}
		c.JSON(http.StatusOK, resp)
		return
	} else {
		var output string
		output, err = cri.StatCommand(cid, path, false)
		if err != nil {
			logrus.Error("podFs show failed err : %+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
		}
		resp, err := parser.ParseStatCmdOutput(output, path, dirOnly)
		if err != nil {
			logrus.Error("podFs show failed err : %+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: err.Error()})
		}
		if len(resp.Files) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrResponse{Msg: fmt.Sprintf("%s not exist", path)})
		}
		if resp.Files[0].Size > MaxFileReadSize {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrResponse{Msg: "文件过大，不支持预览，请下载查看"})
		}
		catOutPut, err := cri.CatCommand(cid, path)
		if err != nil {
			logrus.Error("%+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResponse{Msg: fmt.Sprintf("exec cat command failed, err: %v", err)})
		}
		response := &models.FileResponse{
			Content: string(catOutPut),
		}
		c.JSON(http.StatusOK, response)
		return
	}
}

// 判断路径参数合法性（防路径遍历）
func antiPathInjection(paths ...string) error {
	// 防止注入命令
	for _, p := range paths {
		if strings.Contains(p, "&&") || strings.Contains(p, ";") {
			return errors.New(`path can't have && or ";"`)
		}
		// for _, c := range p {
		// 	if unicode.IsSpace(c) {
		// 		return errors.New(`path can't have white space`)
		// 	}
		// }
	}
	// 防路径注入
	segs := strings.Split(filepath.Join(paths...), string(filepath.Separator))
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		if seg == ".." {
			return errors.New("bad path")
		}
	}
	return nil
}
