/*
Package common
@Time : 2022/1/21 16:53
@Author : sunxy
@File : logger
@description:
*/
package common

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sxy/lianfang/pkg/util"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

func SetupLogger() {
	pwd := util.Pwd()
	logDir := filepath.Join(pwd, "logs")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Printf("Panic, create %s failed!!!Err is %v \n", logDir, err)
		os.Exit(1)
	}
	rotateFileLogger := &lumberjack.Logger{
		Filename:   filepath.Join(pwd, "logs", "LianFang.log"),
		MaxSize:    30, // megabytes
		MaxBackups: 10,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
		LocalTime:  true,
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:          true,
		FullTimestamp:          true,
		PadLevelText:           true,
		DisableQuote:           true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
	})
	switch Cfg.LogLevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		fmt.Printf("logLevel %s not support", Cfg.LogLevel)
		os.Exit(1)
	}
	if os.Getenv(EnvLianFangDebug) != "" || Cfg.LianFangDebug {
		logrus.SetOutput(io.MultiWriter(os.Stdout, rotateFileLogger))
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetOutput(rotateFileLogger)
	}
}
