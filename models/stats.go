package models

import (
	"time"
	"strings"
	"sync"
	"github.com/fsouza/go-dockerclient"
	"github.com/krise3k/armada-stats/utils"
	"github.com/Sirupsen/logrus"
)

type Container struct {
	ID           string
	Name         string
	Address      string
	Status       Status
	Tags         map[string]string
	Uptime       int64
	Stats        struct {
					 CPUPercentage     float64
					 CPUCorePercentage float64
					 Memory            float64
					 MemoryLimit       float64
					 MemoryPercentage  float64
					 Swap              float64
					 NetworkRx         float64
					 NetworkTx         float64
					 BlockRead         float64
					 BlockWrite        float64
				 }
	DockerClient *docker.Client
	Mu           sync.RWMutex
	Err          error
}

func (c *Container) Collect() {
	var (
		previousCPU uint64
		previousSystem uint64
	)

	utils.GetLogger().WithFields(logrus.Fields{"containerID": c.ID, "name":c.Name}).Info("Getting stats for container")

	go func() {
		var memPercent = 0.0
		var cpuPercent = 0.0
		var cpuCorePercent = 0.0

		stats, err := c.getContainerStats()
		if err != nil {
			c.Mu.Lock()
			c.Err = err
			c.Mu.Unlock()
			return
		}

		memoryUsage := float64(stats.MemoryStats.Usage - stats.MemoryStats.Stats.TotalCache)
		if stats.MemoryStats.Limit != 0 {
			memPercent = float64(memoryUsage / float64(stats.MemoryStats.Limit)) * 100.0
		}

		previousCPU = stats.PreCPUStats.CPUUsage.TotalUsage
		previousSystem = stats.PreCPUStats.SystemCPUUsage
		cpuPercent, cpuCorePercent = calculateCPUPercent(previousCPU, previousSystem, stats.CPUStats)
		blkRead, blkWrite := calculateBlockIO(stats.BlkioStats.IOServiceBytesRecursive)

		c.Mu.Lock()
		c.Stats.CPUPercentage = cpuPercent
		c.Stats.CPUCorePercentage = cpuCorePercent
		c.Stats.Memory = memoryUsage
		c.Stats.MemoryLimit = float64(stats.MemoryStats.Limit)
		c.Stats.MemoryPercentage = memPercent
		c.Stats.Swap = float64(stats.MemoryStats.Stats.Swap)
		c.Stats.NetworkRx, c.Stats.NetworkTx = calculateNetwork(stats.Networks)
		c.Stats.BlockRead = blkRead
		c.Stats.BlockWrite = blkWrite
		c.Mu.Unlock()
	}()
}


func (c *Container) getContainerStats() (stats *docker.Stats, err error) {
	errC := make(chan error, 1)
	statsC := make(chan *docker.Stats, 1)
	timeout := 10 * time.Second
	go func() {
		errC <- c.DockerClient.Stats(docker.StatsOptions{ID:c.ID, Stats:statsC, Stream:false, Timeout: timeout})
	}()
	err = <-errC
	defer close(errC)

	if err != nil {
		return nil, err
	}
	stats, _ = <-statsC

	return stats, err
}

func calculateCPUPercent(previousCPU, previousSystem uint64, v docker.CPUStats) (float64, float64) {

	cpuPercent := 0.0
	cpuCorePercent := 0.0
	// calculate the change for the cpu usage of the container in between readings
	cpuDelta := float64(v.CPUUsage.TotalUsage) - float64(previousCPU)
	// calculate the change for the entire system between readings
	systemDelta := float64(v.SystemCPUUsage) - float64(previousSystem)


	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuCorePercent = (cpuDelta / systemDelta) * 100.0
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUUsage.PercpuUsage)) * 100.0
	}

	return cpuPercent, cpuCorePercent
}

func calculateBlockIO(blkio []docker.BlkioStatsEntry) (blkRead float64, blkWrite float64) {

	for _, bioEntry := range blkio {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + float64(bioEntry.Value)
		case "write":
			blkWrite = blkWrite + float64(bioEntry.Value)
		}
	}

	return blkRead, blkWrite
}


func calculateNetwork(network map[string]docker.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, interface_stats := range network {
		rx += float64(interface_stats.RxBytes)
		tx += float64(interface_stats.TxBytes)
	}
	return rx, tx
}
