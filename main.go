/*
@Time : 2021/12/21 9:14
@Author : sunxy
@File : main
@description:
*/
package main

import (
	"fmt"
	"github.com/sxy/lianfang/pkg/common"
	"github.com/sxy/lianfang/pkg/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sxy/lianfang/pkg/router"
)

func main() {
	common.Setup()
	store.Setup()
	msgChan := make(chan error)
	go startV1HttpRouter(msgChan)
	fmt.Println("This is LianFang--联坊")
	fmt.Println(<-msgChan)
}

func startV1HttpRouter(ch chan error) {

	r := ginEngine()
	gin.DefaultWriter = logrus.StandardLogger().Writer()
	v1 := r.Group("/api/v1")
	v1.Use(Cors())
	ws := r.Group("/ws")
	ws.Use(Cors())
	router.SetupDefaultRouter(v1)
	router.SetupContainersRouter(v1)
	router.SetupWsRouter(ws)

	ch <- r.Run(":8081")
}

func ginEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if common.Cfg.GinLog {
		r.Use(gin.Logger())
	}
	return r
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, Origin, X-Requested-With, Accept, X-Registry-Auth")
		c.Header("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
