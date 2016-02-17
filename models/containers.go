package models

import (
	"log"
	"strings"
	"sync"
	"github.com/fsouza/go-dockerclient"
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

func (containerList *Containers ) Add(armadaContainers ArmadaContainerList) {
	for _, armadaContainer := range (armadaContainers) {
		log.Printf("Adding container %s ID %s", armadaContainer.Name, armadaContainer.ID)

		c := &Container{
			ID: armadaContainer.ID,
			DockerClient: dockerClient,
			Name: armadaContainer.Name,
			Address: armadaContainer.Address,
			Tags: armadaContainer.Tags,
			Status: armadaContainer.Status,
		}

		containerList.Mu.Lock()
		containerList.ContainerList = append(containerList.ContainerList, c)
		containerList.Mu.Unlock()

		go c.Collect()
	}

}

func (containerList *Containers ) MatchWithArmada(armadaContainers ArmadaContainerList) {
	containersToRemove := []int{}
	containerList.Mu.Lock()

	for j, c := range containerList.ContainerList {
		isFound := false
		c.Mu.Lock()
		if c.Err != nil {
			log.Printf("Errors getting container %s ID %s stats: %v", c.Name, c.ID, c.Err)
			containersToRemove = append(containersToRemove, j)
			c.Mu.Unlock()
			continue
		}

		for i, armadaContainer := range armadaContainers {
			if strings.HasPrefix(c.ID, armadaContainer.ID) {
				c.Uptime = armadaContainer.Uptime
				c.Status = armadaContainer.Status
				//remove matched container from list
				armadaContainers = append(armadaContainers[:i], armadaContainers[i + 1:]...)
				isFound = true
				break
			}
		}

		if !isFound {
			// container not found in armada list, probably stopped
			log.Printf("Container %s ID %s not found, removing.", c.Name, c.ID)
			containersToRemove = append(containersToRemove, j)
		}

		c.Mu.Unlock()
	}

	//remove containers with errors
	for j := len(containersToRemove) - 1; j >= 0; j-- {
		i := containersToRemove[j]
		containerList.ContainerList = append(containerList.ContainerList[:i], containerList.ContainerList[i + 1:]...)
	}

	containerList.Mu.Unlock()

	if len(armadaContainers) > 0 {
		log.Printf("containers no: %v, list :%v",len(armadaContainers), armadaContainers )
		containerList.Add(armadaContainers)
	}
}
