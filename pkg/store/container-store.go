/*
Package store
@Time : 2022/1/28 14:33
@Author : sunxy
@File : store
@description:
*/
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sxy/lianfang/pkg/common"
	"github.com/xujiajun/nutsdb"
	"io"
	"os"
	"sync"
	"time"
)

var (
	storeService       *Service
	storeServiceHelper sync.Once
)

const (
	ContainerListKey        = "container_list"
	ContainerStatsKeyPrefix = "container_stats_"
)

func GetService() *Service {
	storeServiceHelper.Do(func() {
		storeService = &Service{
			dockerClient:    common.DockerClient(),
			containerBucket: "default",
			OpBucket:        "operation",
			name:            "default",
			nutsDBHandler:   common.NutsDB,
		}
	})
	return storeService
}

type Service struct {
	dockerClient    *client.Client
	name            string
	containerBucket string
	// 操作表，记录对容器执行的启停操作
	OpBucket      string
	nutsDBHandler *nutsdb.DB
}

func (s *Service) Start() {
	s.clean()
	// 初始化后写一下，确保启动后写库没问题
	err := s.put(s.containerBucket, []byte("initSetUp"), []byte("initSetUp"))
	if err != nil {
		logrus.Errorf("init store service failed, program exit ...")
		os.Exit(1)
	}
	d := time.Second * time.Duration(common.Cfg.UpdateInterval)
	t := time.NewTicker(d)
	defer t.Stop()
	errCount := 0
	for {
		<-t.C
		logrus.Debugf("service [%s] update", s.name)
		err := s.updateContainerList()
		if err != nil {
			// 失败次数增加，跳过容器状态更新，因为容器列表都已经更新失败，状态旧没有必要更新了
			errCount++
			logrus.Errorf("%+v", err)
		} else {
			err = s.updateContainerStats()
			if err != nil {
				errCount++
				logrus.Errorf("%+v", err)
			} else {
				// 如果容器列表与容器状态都更新成功，重置失败次数
				errCount = 0
			}
		}
		if errCount == 3 {
			logrus.Errorf("定时刷新容器信息协程连续失败3次，程序退出")
			os.Exit(1)
		}
	}
}

func (s *Service) clean() {
	logrus.Infof("cleaning nutsdb...")
	err := s.nutsDBHandler.Update(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(s.containerBucket)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				err := tx.Delete(s.containerBucket, entry.Key)
				if err != nil {
					logrus.Warnf("%+v", errors.Wrapf(err, "container-store clean key [%s] error", entry.Key))
				}
			}
			return nil
		})
	if err != nil {
		logrus.Warnf("%+v", errors.Wrap(err, "container-store clean error"))
	}
	logrus.Infof("clean nutsdb finished")
}

func (s *Service) put(bucket string, key, value []byte) error {
	err := s.nutsDBHandler.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucket, key, value, 0); err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return errors.Wrapf(err, "store service [%s] update data failed", s.name)
	}
	return nil
}

func (s Service) get(bucket string, key []byte) (ret []byte, err error) {
	ret = make([]byte, 0)
	err = nil
	err = s.nutsDBHandler.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(bucket, key); err != nil {
				if err == nutsdb.ErrNotFoundKey {
					return nil
				}
				return err
			} else {
				ret = e.Value
			}
			return nil
		})
	return
}

func (s Service) delete(bucket string, key []byte) error {
	if err := s.nutsDBHandler.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Delete(bucket, key); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetContainerList() ([]types.Container, error) {
	csbs, err := s.get(s.containerBucket, []byte(ContainerListKey))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(csbs) == 0 {
		return nil, errors.New("No container list")
	}
	result := make([]types.Container, 0)
	err = json.Unmarshal(csbs, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (s *Service) AddOp(cid, action string) error {
	err := s.put(s.OpBucket, []byte(cid), []byte(action))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Service) DeleteOp(cid string) error {
	err := s.delete(s.OpBucket, []byte(cid))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Service) GetOp(cid string) (string, error) {
	ret, err := s.get(s.OpBucket, []byte(cid))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(ret), nil
}

func (s *Service) GetContainerStats(cid string) ([]byte, error) {
	return s.get(s.containerBucket, s.getContainerStatsKey(cid))
}

// 定期刷新容器状态信息
func (s *Service) updateContainerList() error {
	cs, err := s.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return errors.Wrapf(err, "list container failed")
	}
	csBytes, err := json.Marshal(cs)
	if err != nil {
		return errors.Wrapf(err, "list container failed")
	}
	return s.put(s.containerBucket, []byte(ContainerListKey), csBytes)
}

func (s *Service) updateContainerStats() error {
	cs, err := s.GetContainerList()
	if err != nil {
		return err
	}
	for _, c := range cs {
		key := s.getContainerStatsKey(c.ID)
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		sts, err := common.DockerClient().ContainerStats(ctx, c.ID, false)
		if err != nil {
			return errors.Wrapf(err, "Get container %s stats failed!", c.ID)
		}
		defer sts.Body.Close()
		bs, err := io.ReadAll(sts.Body)
		if err != nil {
			return errors.Wrapf(err, "Read container %s stats failed!", c.ID)
		}
		sts.Body.Close()
		err = s.put(s.containerBucket, key, bs)
		if err != nil {
			return errors.Wrapf(err, "put container %s stats to nutsdb failed!", c.ID)
		}
	}
	return nil
}

func (s Service) getContainerStatsKey(cid string) []byte {
	return []byte(fmt.Sprintf("%s_%s", ContainerStatsKeyPrefix, cid))
}
