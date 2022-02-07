/*
Package store
@Time : 2022/1/28 15:30
@Author : sunxy
@File : setup
@description:
*/
package store

func Setup() {
	go GetService().Start()
}
