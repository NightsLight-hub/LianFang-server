/*
@Time : 2021/12/21 17:26
@Author : sunxy
@File : default
@description:
*/
package router

import "github.com/gin-gonic/gin"

func SetupDefaultRouter(engine *gin.RouterGroup) {
	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	engine.GET("/whoayou", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is LianFang--联坊",
		})
	})
}
