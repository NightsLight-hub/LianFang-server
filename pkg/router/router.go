/*
@Time : 2021/12/21 14:41
@Author : sunxy
@File : router
@description:
*/
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sxy/lianfang/pkg/handler/containers"

	"net/http"
)

func SetupContainersRouter(engine *gin.RouterGroup) {
	engine.GET("/containers", func(c *gin.Context) {
		cs, err := containers.GetService().List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Server Error",
			})
		} else {
			c.JSON(http.StatusOK, cs)
		}
	})
	engine.GET("/container/:cid/stats", func(c *gin.Context) {
		cid := c.Param("cid")
		sts, err := containers.GetService().Stats(cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Server Error",
			})
		} else {
			c.String(http.StatusOK, string(sts))
		}
	})
}
