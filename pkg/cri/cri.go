/*
@Time : 2021/12/28 17:33
@Author : sunxy
@File : exec
@description:
*/
package cri

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/common"
	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/logger"
	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/pkg/cri/docker"
	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/pkg/cri/k8s"
	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/util"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	K8SOperationTimeout = 60 * time.Second
	cacheExpireTime     = 5 * time.Second
)

const (
	CommandNotFound = "command not found"
)

/**
export LC_ALL=C
export LANG=en_US.UTF-8
*/
// 暂时废弃，等以后做kubernetes的时候复用
func LsCommand(namespace, pod, container, path string) (string, error) {
	cacheKey := common.LsCommandCacheKey(namespace, pod, container, path)
	cacheResult := common.Cache().Get(cacheKey)
	if cacheResult != nil {
		return cacheResult.(string), nil
	}
	// 注意 centos / alpine / busybox 的 ls 命令输出区别
	cmd := []string{"sh", "-c",
		fmt.Sprintf(`export LC_ALL=C && export LANG=en_US.UTF-8 && ls -lAL --full-time --color=never '%s'`, path)}
	output, errOut, err := k8s.Exec(cmd, namespace, pod, container)
	if len(errOut) != 0 || err != nil {
		logger.Error(fmt.Sprintf("K8s pod [%s:%s:%s] exec ls command failed, errout: %s, err: %+v",
			namespace, pod, container, errOut, err))
		return "", errors.New(fmt.Sprintf("K8s pod [%s:%s:%s] exec ls command failed, errout: %s, err: %v",
			namespace, pod, container, errOut, err))
	}
	cacheErr := common.Cache().Put(cacheKey, output, cacheExpireTime)
	if cacheErr != nil {
		logger.Warn("Update LsCommand cache[%s] failed", cacheKey)
	}
	return output, nil
}

func StatCommand(containerId, path string, isDir bool) (string, error) {
	cacheKey := common.StatCommandCacheKey(containerId, path)
	cacheResult := common.Cache().Get(cacheKey)
	if cacheResult != nil {
		logger.Debug("cache key %s hit", cacheKey)
		return cacheResult.(string), nil
	}
	var command string
	if isDir {
		command = fmt.Sprintf("export LC_ALL=C && export LANG=en_US.UTF-8 && cd '%s' && stat -L -c '%%A %%s %%Z %%n' .* *", path)
	} else {
		command = fmt.Sprintf("export LC_ALL=C && export LANG=en_US.UTF-8&& stat -L -c '%%A %%s %%Z %%n' '%s'", path)
	}
	cmd := []string{"sh", "-c", command}
	output, errOut, err := docker.Exec(containerId, cmd)
	// 没有文件返回空
	if strings.Contains(errOut, "stat: cannot stat '*': No such file or directory") || strings.Contains(errOut, "stat: can't stat '*': No such file or directory") {
		return "", err
	}
	output, err = must(cmd, output, errOut, err)
	if err != nil {
		return "", err
	}
	cacheErr := common.Cache().Put(cacheKey, output, cacheExpireTime)
	if cacheErr != nil {
		logger.Warn("Update LsCommand cache[%s] failed", cacheKey)
	}
	return output, nil
}

func CatCommand(containerId, path string) (string, error) {
	// cat 不常用，且文件内容可能很大，不适用缓存来节省内存
	cmd := []string{"sh", "-c", fmt.Sprintf("cat '%s'", path)}
	output, errOut, err := docker.Exec(containerId, cmd)
	return must(cmd, output, errOut, err)
}

// 检查 errOut 和 err 都是空，不为空说明命令执行错误
func must(cmd []string, output, errOut string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if len(errOut) != 0 {
		return "", errors.Errorf("exec cmd [ %v ] get error: %s", cmd, errOut)
	}
	return output, nil
}

//
//  CopyFromContainer
//  @Description:       well, copy file/dir from container to host path
//  @param containerId
//  @param srcFilePath
//  @param destFilePath
//  @return err
func CopyFromContainer(containerId, srcFilePath, destFilePath string) (err error) {
	return docker.CopyFrom(containerId, srcFilePath, destFilePath)
}

//
//  CopyToContainer
//  @Description:
//  @param containerId
//  @param srcFilePath
//  @param destFilePath
//  @return err
func CopyToContainer(containerId, srcFilePath, destFilePath string) error {
	go func() {
		cacheKey := common.StatCommandCacheKey(containerId, destFilePath)
		common.Cache().Delete(cacheKey)
		logger.Debug("cache key %s delete", cacheKey)
	}()
	return docker.CopyTo(containerId, srcFilePath, destFilePath)
}

//
//  Move
//  @Description: 容器内 移动目录
//  @param cid
//  @param srcPath
//  @param dstPath
//  @return error
//
func Move(cid, srcPath, dstPath string) error {
	cmd := []string{"sh", "-c", fmt.Sprintf("mv '%s' '%s'", srcPath, dstPath)}
	output, errOut, err := docker.Exec(cid, cmd)
	_, err = must(cmd, output, errOut, err)
	//  删除文件目录的 stat 命令缓存，避免查看的时候看不到
	go func() {
		p := filepath.Dir(srcPath)
		cacheKey := common.StatCommandCacheKey(cid, p)
		common.Cache().Delete(cacheKey)
	}()
	return err
}

//
//  Delete
//  @Description:  批量删除文件
//  @param cid
//  @param srcPaths
//  @return error
//
func Delete(cid string, srcPaths []string) error {
	pathsBuffer := new(bytes.Buffer)
	for _, p := range srcPaths {
		pathsBuffer.WriteString(fmt.Sprintf(`'%s'`, p))
		pathsBuffer.WriteString(" ")
	}
	cmd := []string{"sh", "-c", fmt.Sprintf("rm -rf %s", strings.TrimSpace(pathsBuffer.String()))}
	output, errOut, err := docker.Exec(cid, cmd)
	_, err = must(cmd, output, errOut, err)
	//  删除文件目录的 stat 命令缓存，避免查看的时候看不到
	go func(srcPath string) {
		p := filepath.Dir(srcPath)
		cacheKey := common.StatCommandCacheKey(cid, p)
		common.Cache().Delete(cacheKey)
		logger.Debug("cache key %s delete", cacheKey)
	}(srcPaths[0])
	return err
}

func UnCompress(cid, path, dest, format string) error {
	var c string
	switch format {
	case "zip":
		c = fmt.Sprintf(`mkdir -p '%s' && unzip '%s' -d '%s'`, dest, path, dest)
	case "tar":
		c = fmt.Sprintf(`mkdir -p '%s' && tar x -f '%s' -C '%s'`, dest, path, dest)
	case "tar.gz", "tgz":
		c = fmt.Sprintf(`mkdir -p '%s' && tar x -zf '%s' -C '%s'`, dest, path, dest)
	default:
		return errors.Errorf("file format %s not supported", format)
	}
	cmd := []string{"sh", "-c", c}
	_, errOut, err := docker.Exec(cid, cmd)
	if err != nil {
		return err
	}
	if len(errOut) != 0 && strings.Contains(errOut, "not found") {
		return errors.New(CommandNotFound)
	}
	//  删除文件目录的 stat 命令缓存，避免查看的时候看不到
	go func(p string, err error) {
		if err == nil {
			cacheKey := common.StatCommandCacheKey(cid, p)
			common.Cache().Delete(cacheKey)
			logger.Debug("cache key %s delete", cacheKey)
			// 有时候，我们将压缩包解压到当前目录的某个文件夹里，即
			// /home/a.tar -> /home/a/a.tar 此时 p == /home/a
			// 则只删除/home/a的缓存不够用，还要删除其父目录的缓存
			cacheKey = common.StatCommandCacheKey(cid, filepath.Dir(p))
			common.Cache().Delete(cacheKey)
			logger.Debug("cache key %s delete", cacheKey)
		}
	}(dest, err)
	return err
}

func ValidContainer(containerId string) (err error) {
	_, err = common.DockerClient().ContainerInspect(context.Background(), containerId)
	if err != nil {
		logger.Error(errors.WithStack(err))
		return util.ResourceNotExistError{Msg: fmt.Sprintf("container %s not found", containerId)}
	}
	return nil
}

func GetContainerFileInfo(containerId, path string) (*types.ContainerPathStat, error) {
	stat, err := common.DockerClient().ContainerStatPath(context.Background(), containerId, path)
	if err != nil {
		return nil, errors.Wrapf(err, "get containerpath %s:%s info failed", containerId, path)
	}
	return &stat, nil
}

//
//  ValidPodInfo
//  @Description:    目前不需要支持k8s的接口，暂时废弃
//  @param namespace
//  @param pod
//  @param container
//  @return error
func ValidPodInfo(namespace, pod, container string) error {
	podsClient := common.KubeClientSet().CoreV1().Pods(namespace)
	if podsClient == nil {
		return errors.New("获取 kubePodsClient失败，可能是命名空间不存在或者kubernetes故障！")
	}
	ctx, _ := context.WithTimeout(context.TODO(), K8SOperationTimeout)
	podInfo, err := podsClient.Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "获取pod  %s:%s 失败", namespace, pod)
	}
	flag := false
	for _, c := range podInfo.Spec.Containers {
		if c.Name == container {
			flag = true
			break
		}
	}
	if !flag {
		return &util.ResourceNotExistError{
			Msg: fmt.Sprintf("pod %s:%s dosen't have container %s", namespace, pod, container),
		}
	}

	err = tarVersionCommand(namespace, pod, container)
	if err != nil {
		return &util.TarCmdFailedError{
			Msg: fmt.Sprintf("pod %s:%s:%s dosen't have tar cmd, so We can't upload or download files", namespace, pod, container),
		}
	}
	return nil
}

func tarVersionCommand(namespace, pod, container string) error {
	cmd := []string{"sh", "-c", "tar --version"}
	_, _, err := k8s.Exec(cmd, namespace, pod, container)
	return err
}

func Mkdir(cid, path string) error {
	cmd := []string{"sh", "-c", fmt.Sprintf(`mkdir -p '%s'`, path)}
	outPut, errOut, err := docker.Exec(cid, cmd)
	_, err = must(cmd, outPut, errOut, err)
	return err
}
