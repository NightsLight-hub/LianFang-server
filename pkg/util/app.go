package util

import (
	"os"
	"path/filepath"
)

// Pwd 获取程序所在目录
func Pwd() string {
	appDir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		panic(err)
	}
	return appDir
}

func ShutDown() {
	os.Exit(1)
}
