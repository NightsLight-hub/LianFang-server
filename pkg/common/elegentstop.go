/*
Package common
@Time : 2022/1/8 13:09
@Author : sunxy
@File : elegentstop
@description:
*/
package common

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func RegisterElegantStop() {
	c := make(chan os.Signal)
	// 监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				logrus.Warn("Program will Exit... %s", s)
				elegantStop()
				os.Exit(0)
			default:
				logrus.Warn("Ignore signal...%s", s)
			}
		}
	}()
}

func elegantStop() {
	logrus.Info("close nutsDB")
	NutsDB.Close()
}
