package models
import (
	"net/http"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/krise3k/armada-stats/utils"
)

type ArmadaContainer struct {
	ID      string
	Name    string
	Address string
	Status  Status
	Tags    map[string]string
	Uptime  int64
}

type APIContainerList struct {
	Status string `json:"status"`
	Result []ArmadaAPIContainer `json:"result"`
}

type ArmadaAPIContainer struct {
	Name           string             `json:"name"`
	Address        string             `json:"address"`
	ID             string             `json:"container_id"`
	Status         string             `json:"status"`
	Tags           map[string]string  `json:"tags"`
	StartTimestamp string             `json:"start_timestamp"`
}

type Status int

const (
	passing Status = iota
	warning
	critical
)

type ArmadaContainerList [] ArmadaContainer

func getUptime(timestamp string) (int64, error) {
	startTimestamp, err := strconv.Atoi(timestamp)
	if err != nil {
		return 0, err
	}

	started := time.Unix(int64(startTimestamp), 0)

	return int64(time.Since(started).Seconds()), nil
}

func parseStatus(statusStr string) (Status) {
	switch statusStr {
	case "passing":
		return passing
	case "warning":
		return warning
	case "critical":
		return critical
	default:
		utils.GetLogger().Error("Unknown container status " + statusStr)

		return -1
	}
}

func parseContainer(apiContainer ArmadaAPIContainer) (ArmadaContainer) {
	uptime, err := getUptime(apiContainer.StartTimestamp)
	if err != nil {
		utils.GetLogger().WithError(err).Error("Error getting container uptime")
	}
	container := ArmadaContainer{
		Name:           apiContainer.Name,
		Address:        apiContainer.Address,
		ID:             apiContainer.ID,
		Status:         parseStatus(apiContainer.Status),
		Tags:           apiContainer.Tags,
		Uptime:         uptime,
	}

	return container
}

func isSubService(containerName string) bool {
	return strings.Contains(containerName, ":")
}

func convertToArmadaContainer(apiContainerList []ArmadaAPIContainer) (ArmadaContainerList) {
	armadaContainerList := ArmadaContainerList{}
	for _, container := range (apiContainerList) {
		if isSubService(container.Name) {
			continue
		}
		armadaContainer := parseContainer(container)
		armadaContainerList = append(armadaContainerList, armadaContainer)
	}

	return armadaContainerList
}

func GetArmadaContainerList() (ArmadaContainerList) {
	apiContainerList := new(APIContainerList)
	host, _ := utils.Config.String("armada_host")
	port, _ := utils.Config.String("armada_port")

	query_string := "list?local=true"

	url := fmt.Sprintf("http://%s:%s/%s", host, port, query_string)
	resp, err := http.Get(url)

	defer resp.Body.Close()

	if err != nil {
		utils.GetLogger().WithError(err).Panic("Cannot get container list from armada")
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiContainerList); err != nil {
		utils.GetLogger().WithError(err).Panic("Cannot decode json with container list")
	}
	containerList := convertToArmadaContainer(apiContainerList.Result)

	return containerList
}