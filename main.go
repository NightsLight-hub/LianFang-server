/*
@Time : 2021/12/21 9:14
@Author : sunxy
@File : main
@description:
*/
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sxy/lianfang/pkg/router"
	"net/http"
)

func main() {
	msgChan := make(chan error)
	go startV1HttpRouter(msgChan)
	fmt.Println("This is LianFang--联坊")
	fmt.Println(<-msgChan)
}

func startV1HttpRouter(ch chan error) {
	r := gin.Default()
	v1 := r.Group("v1")
	v1.Use(Cors())
	router.SetupDefaultRouter(v1)
	router.SetupContainersRouter(v1)

	ch <- r.Run(":8081")
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, Origin, X-Requested-With, Accept, X-Registry-Auth")
		c.Header("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
