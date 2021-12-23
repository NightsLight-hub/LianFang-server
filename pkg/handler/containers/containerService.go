/*
@Time : 2021/12/21 14:50
@Author : sunxy
@File : containersHandler
@description:
*/
package containers

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	lClient "github.com/sxy/lianfang/pkg/client"
	"golang.org/x/net/context"
	"io"
	"sync"
	"time"
)

var (
	containerServiceInstance     *Service
	containerServiceInstanceOnce sync.Once
)

func GetService() *Service {
	containerServiceInstanceOnce.Do(func() {
		containerServiceInstance = &Service{dockerClient: lClient.GetDockerClient()}
	})
	return containerServiceInstance
}

type Service struct {
	dockerClient *client.Client
}

func (c *Service) List() ([]types.Container, error) {
	cs, err := c.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(err, "list containers failed!")
	}
	return cs, nil
}

// To calculate the values shown by the stats command of the docker cli tool the following formulas can be used:
//
// used_memory = memory_stats.usage - memory_stats.stats.cache
// available_memory = memory_stats.limit
// Memory usage % = (used_memory / available_memory) * 100.0
// cpu_delta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
// system_cpu_delta = cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage
// number_cpus = lenght(cpu_stats.cpu_usage.percpu_usage) or cpu_stats.online_cpus
// CPU usage % = (cpu_delta / system_cpu_delta) * number_cpus * 100.0

// Stats  get the specific container resource status
func (c *Service) Stats(cid string) ([]byte, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	sts, err := c.dockerClient.ContainerStats(ctx, cid, false)
	if err != nil {
		return nil, errors.Wrap(err, "Get container stats failed!")
	}
	defer sts.Body.Close()
	bs, err := io.ReadAll(sts.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Read container stats failed!")
	}
	//var cst = new(types.Stats)
	//err = json.Unmarshal(bs, cst)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "Unmarshal bytes to json failed, bytes is %s", bs)
	//}
	//// used_memory = memory_stats.usage - memory_stats.stats.cache
	//memoryCache, ok := cst.MemoryStats.Stats["cache"]
	//if !ok {
	//	memoryCache = 0
	//}
	//usedMemory := cst.MemoryStats.Usage - memoryCache
	//// available_memory = memory_stats.limit
	//availableMemory := cst.MemoryStats.Limit
	//// Memory usage % = (used_memory / available_memory) * 100.0
	//memoryUsagePercent := util.Round(float64(usedMemory)/float64(availableMemory)*100, 1)
	//// cpu_delta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
	//cpuDelta := cst.CPUStats.CPUUsage.TotalUsage - cst.PreCPUStats.CPUUsage.TotalUsage
	//// system_cpu_delta = cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage
	//systemCpuDeleta := cst.CPUStats.SystemUsage - cst.PreCPUStats.SystemUsage
	//// number_cpus = lenght(cpu_stats.cpu_usage.percpu_usage) or cpu_stats.online_cpus
	//cpuNumber := cst.CPUStats.OnlineCPUs
	//if cpuNumber == 0 {
	//	cpuNumber = uint32(len(cst.CPUStats.CPUUsage.PercpuUsage))
	//}
	//// CPU usage % = (cpu_delta / system_cpu_delta) * number_cpus * 100.0
	//cpuUsage := float64(cpuDelta) / float64(systemCpuDeleta) * float64(cpuNumber) * 100
	return bs, nil
}
