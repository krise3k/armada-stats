package models

import (
	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/krise3k/armada-stats/models/armada"
	"github.com/krise3k/armada-stats/utils"
	"strings"
	"sync"
)

var dockerClient *docker.Client

func init() {
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, _ = docker.NewClient(endpoint)
}

type Containers struct {
	Mu            sync.Mutex
	ContainerList []*Container
}

func (containerList *Containers) Add(armadaContainers armada.ArmadaContainerList) {

	waitFirstStats := &sync.WaitGroup{}

	for _, armadaContainer := range armadaContainers {
		utils.GetLogger().WithFields(logrus.Fields{"containerID": armadaContainer.ID, "name": armadaContainer.Name}).Info("Adding container")

		c := &Container{
			ID:           armadaContainer.ID,
			DockerClient: dockerClient,
			Name:         armadaContainer.Name,
			Address:      armadaContainer.Address,
			Tags:         armadaContainer.Tags,
			Status:       armadaContainer.Status,
		}

		containerList.Mu.Lock()
		containerList.ContainerList = append(containerList.ContainerList, c)
		containerList.Mu.Unlock()

		waitFirstStats.Add(1)
		go c.Collect(waitFirstStats)
	}

	waitFirstStats.Wait()
}

func (containerList *Containers) MatchWithArmada(armadaContainers armada.ArmadaContainerList) {
	containersToRemove := []int{}
	containerList.Mu.Lock()

	for j, c := range containerList.ContainerList {
		isFound := false
		c.Mu.Lock()
		if c.Err != nil {
			utils.GetLogger().WithFields(logrus.Fields{"containerID": c.ID, "name": c.Name}).WithError(c.Err).Error("Error getting container stats")
			containersToRemove = append(containersToRemove, j)
			c.Mu.Unlock()
			continue
		}

		for i, armadaContainer := range armadaContainers {
			if strings.HasPrefix(c.ID, armadaContainer.ID) {
				c.Uptime = armadaContainer.Uptime
				c.Status = armadaContainer.Status
				//remove matched container from list
				armadaContainers = append(armadaContainers[:i], armadaContainers[i+1:]...)
				isFound = true
				break
			}
		}

		if !isFound {
			// container not found in armada list, probably stopped
			utils.GetLogger().WithFields(logrus.Fields{"containerID": c.ID, "name": c.Name}).Info("Container not found, removing")

			containersToRemove = append(containersToRemove, j)
		}

		c.Mu.Unlock()
	}

	//remove containers with errors
	for j := len(containersToRemove) - 1; j >= 0; j-- {
		i := containersToRemove[j]
		containerList.ContainerList = append(containerList.ContainerList[:i], containerList.ContainerList[i+1:]...)
	}

	containerList.Mu.Unlock()

	if len(armadaContainers) > 0 {
		containerList.Add(armadaContainers)
	}
}
