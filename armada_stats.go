package main

import (
	"flag"
	"fmt"
	"github.com/krise3k/armada-stats/models"
	"github.com/krise3k/armada-stats/utils"
	"log"
	"time"
)

var (
	configPath = flag.String("config", "/etc/armada-stats/armada-stats.yml", "config file location")
)

func main() {
	flag.Parse()

	utils.InitConfig(*configPath)
	ravenClient := utils.GetRaven()

	if err := GatherStats(); err != nil {
		ravenClient.CaptureErrorAndWait(err, nil)
		log.Println(err)
	}
}

func GatherStats() (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Panic: %v", r)
			}
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
			log.Println("ContainerList is empty, is armada running")
		}

		containers.Mu.Unlock()
	}

	return
}
