package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/krise3k/armada-stats/models"
	"github.com/krise3k/armada-stats/utils"
	"time"
)

var (
	configPath = flag.String("config", "/etc/armada-stats/armada-stats.yml", "config file location")
	logger     *logrus.Logger
)

func main() {
	flag.Parse()

	utils.InitConfig(*configPath)
	logger = utils.GetLogger()

	GatherStats()
}

func GatherStats() {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			var err error
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Panic: %v", r)
			}
			logger.WithError(err).Fatal("Captured panic")
		}
	}()

	suf, _ := utils.Config.Int("stats_update_frequency")
	stats_update_frequency := time.Duration(suf) * time.Second
	containers := new(models.Containers)

	armadaContainers := models.GetArmadaContainerList()
	containers.Add(armadaContainers)

	// quick pause for getting initial values
	//@todo handle better
	time.Sleep(stats_update_frequency)

	for range time.Tick(stats_update_frequency) {
		armadaContainers = models.GetArmadaContainerList()

		containers.MatchWithArmada(armadaContainers)
		containers.Mu.Lock()

		for _, s := range containers.ContainerList {
			go models.SendToInflux(*s)
			go s.Collect()
		}

		if len(containers.ContainerList) == 0 {
			logger.Warn("ContainerList is empty, is armada running")
		}

		containers.Mu.Unlock()
	}

	return
}
