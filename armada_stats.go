package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/krise3k/armada-stats/models"
	"github.com/krise3k/armada-stats/models/armada"
	"github.com/krise3k/armada-stats/utils"
	"os"
	"sync"
	"time"
)

var (
	configPath = flag.String("config", "/etc/armada-stats/armada-stats.yml", "config file location")
	v          = flag.Bool("v", false, "prints current version and exits")
	logger     *logrus.Logger
	version    string
)

func main() {

	flag.Parse()
	if *v {
		fmt.Printf("Version %s \n", version)
		os.Exit(0)
	}
	utils.InitConfig(*configPath)
	utils.Config.Set("version", version)
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

	armadaHost, _ := utils.Config.String("armada_host")
	armadaPort, _ := utils.Config.String("armada_port")
	armadaClient := armada.NewArmadaClient(armadaHost, armadaPort)
	armadaContainers := armadaClient.GetLocalContainerList()

	containers := new(models.Containers)
	containers.Add(armadaContainers)

	for range time.Tick(stats_update_frequency) {
		armadaContainers = armadaClient.GetLocalContainerList()
		containers.MatchWithArmada(armadaContainers)

		if len(containers.ContainerList) == 0 {
			logger.Warn("ContainerList is empty, check is armada running")
			continue
		}

		containers.Mu.Lock()

		waitCollectAll := &sync.WaitGroup{}
		waitCollectAll.Add(len(containers.ContainerList))
		for _, cl := range containers.ContainerList {
			go cl.Collect(waitCollectAll)
		}
		waitCollectAll.Wait()

		go models.SendMetrics(*containers)

		containers.Mu.Unlock()
	}

	return
}
