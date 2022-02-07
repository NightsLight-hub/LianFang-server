/*
Package common
@Time : 2022/1/28 12:28
@Author : sunxy
@File : nustdb
@description:
*/
package common

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xujiajun/nutsdb"
	"os"
)

var NutsDB *nutsdb.DB

func SetupDb() {
	// Open the database located in the /tmp/nutsdb directory.
	// It will be created if it doesn't exist.
	opt := nutsdb.DefaultOptions
	opt.Dir = Cfg.NutsDir
	var err error
	NutsDB, err = nutsdb.Open(opt)
	if err != nil {
		logrus.Errorf("%v", errors.Wrapf(err, "open nustDB failed!"))
		os.Exit(1)
	}
}
