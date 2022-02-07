/*
Package common
@Time : 2022/1/28 15:28
@Author : sunxy
@File : setup
@description:
*/
package common

func Setup() {
	SetupConfig()
	SetupLogger()
	SetupCache()
	SetupDb()
	RegisterElegantStop()
}
