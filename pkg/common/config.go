/*
Package common
@Time : 2022/1/21 17:08
@Author : sunxy
@File : config
@description:
*/
package common

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/sxy/lianfang/pkg/util"
	"os"
	"path/filepath"
)

var Cfg *Config

func SetupConfig() {
	Cfg = DefaultConfig()
	viper.SetConfigName("LianFang") // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	confPath := filepath.Join(util.Pwd(), "conf")
	viper.AddConfigPath(confPath) // path to look for the config file in
	viper.AddConfigPath("./conf")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Printf("Can not find config file in %s", confPath)
			return
		} else {
			// Config file was found but another error was produced
			fmt.Printf("Read config failed, %s", err.Error())
			os.Exit(1)
		}
	}
	Cfg.LianFangDebug = viper.GetBool("LianFangDebug")
	Cfg.LogLevel = viper.GetString("LogLevel")
	Cfg.NutsDir = viper.GetString("NutsDir")
	Cfg.UpdateInterval = viper.GetInt("UpdateInterval")
	Cfg.ContainerStopTimeout = viper.GetInt64("ContainerStopTimeout")
	Cfg.GinLog = viper.GetBool("GinLog")
}

type Config struct {
	LianFangDebug        bool   `json:"LianFangDebug"`
	LogLevel             string `json:"LogLevel"`
	NutsDir              string `json:"NutsDir"`
	UpdateInterval       int    `json:"UpdateInterval"`
	ContainerStopTimeout int64  `json:"ContainerStopTimeout"`
	GinLog               bool   `json:"GinLog"`
}

func DefaultConfig() *Config {
	return &Config{
		LianFangDebug:  false,
		LogLevel:       "debug",
		NutsDir:        "./store",
		UpdateInterval: 3,
	}
}
