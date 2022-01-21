/*
@Time : 2021/12/21 13:39
@Author : sunxy
@File : stats
@description:
*/
package common

import (
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	dockerClient     *client.Client
	dockerClientOnce sync.Once
)

func DockerClient() *client.Client {
	dockerClientOnce.Do(func() {
		var err error
		dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			logrus.WithError(err).Fatal("Connect to docker failed!")
		}
	})
	return dockerClient
}
