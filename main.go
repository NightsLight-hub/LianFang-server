/*
@Time : 2021/12/21 9:14
@Author : sunxy
@File : main
@description:
*/
package LianFang_server

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/whoayou", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is LianFang--联坊",
		})
	})
	r.Run()

	fmt.Println("This is LianFang--联坊")
}
