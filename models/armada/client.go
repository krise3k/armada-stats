package armada

import (
	"encoding/json"
	"fmt"
	"github.com/krise3k/armada-stats/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ArmadaClient struct {
	Host string
	Port string
}

type ArmadaContainer struct {
	ID         string
	Name       string
	Address    string
	Status     Status
	StatusName string
	Tags       map[string]string
	Uptime     int64
}

type APIContainerList struct {
	Status string               `json:"status"`
	Result []ArmadaAPIContainer `json:"result"`
}

type ArmadaAPIContainer struct {
	Name           string            `json:"name"`
	Address        string            `json:"address"`
	ID             string            `json:"container_id"`
	Status         string            `json:"status"`
	Tags           map[string]string `json:"tags"`
	StartTimestamp string            `json:"start_timestamp"`
}

type Status int

const (
	passing Status = iota
	warning
	critical
)

type ArmadaContainerList []ArmadaContainer

func getUptime(timestamp string) (int64, error) {
	startTimestamp, err := strconv.Atoi(timestamp)
	if err != nil {
		return 0, err
	}

	started := time.Unix(int64(startTimestamp), 0)

	return int64(time.Since(started).Seconds()), nil
}

func parseStatus(statusStr string) Status {
	switch statusStr {
	case "passing", "recovering":
		return passing
	case "warning":
		return warning
	case "crashed", "critical", "not-recovered":
		return critical
	default:
		utils.GetLogger().Error("Unknown container status " + statusStr)

		return -1
	}

}

func parseContainer(apiContainer ArmadaAPIContainer) ArmadaContainer {
	uptime, err := getUptime(apiContainer.StartTimestamp)
	if err != nil {
		utils.GetLogger().WithError(err).Error("Error getting container uptime")
	}
	container := ArmadaContainer{
		Name:       apiContainer.Name,
		Address:    apiContainer.Address,
		ID:         apiContainer.ID,
		Status:     parseStatus(apiContainer.Status),
		StatusName: apiContainer.Status,
		Tags:       apiContainer.Tags,
		Uptime:     uptime,
	}

	return container
}

func isSubService(containerName string) bool {
	return strings.Contains(containerName, ":")
}

func convertToArmadaContainer(apiContainerList []ArmadaAPIContainer) ArmadaContainerList {
	armadaContainerList := ArmadaContainerList{}
	for _, container := range apiContainerList {
		if isSubService(container.Name) {
			continue
		}
		armadaContainer := parseContainer(container)
		armadaContainerList = append(armadaContainerList, armadaContainer)
	}

	return armadaContainerList
}
func NewArmadaClient(host string, port string) *ArmadaClient {
	return &ArmadaClient{
		Host: host,
		Port: port,
	}
}

func (c *ArmadaClient) GetLocalContainerList() ArmadaContainerList {

	query_string := "local=true"
	return c.getList(query_string)
}

func (c *ArmadaClient) GetContainerListByServiceName(service_name string) ArmadaContainerList {

	query_string := fmt.Sprintf("microservice_name=%s", service_name)
	return c.getList(query_string)
}

func (c *ArmadaClient) getList(query string) ArmadaContainerList {
	apiContainerList := new(APIContainerList)
	url := fmt.Sprintf("http://%s:%s/list?%s", c.Host, c.Port, query)
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
