package models

import (
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/krise3k/armada-stats/utils"
	"github.com/krise3k/armada-stats/utils/influx"
	"github.com/serenize/snaker"
	"os"
)

type ShipSummaryCounter map[string]int16

func SendMetrics(containers Containers) {
	hostname := getHostname()
	cluster_name := getClusterName()
	influxClient := influx.GetInfluxClient()
	batch := influxClient.CreateBatchPoints()
	summary_by_status := ShipSummaryCounter{}

	for _, container := range containers.ContainerList {
		if len(container.StatusName) > 0 {
			summary_by_status = updateShipSummary(summary_by_status, container)
		}
		point := createPoint(container, hostname, cluster_name)
		batch.AddPoint(point)
	}

	if len(summary_by_status) > 0 {
		point := createSummary(summary_by_status, hostname, cluster_name)
		batch.AddPoint(point)
	}

	influxClient.Save(batch)
}
func updateShipSummary(summary_by_status ShipSummaryCounter, container *Container) ShipSummaryCounter {
	if _, ok := summary_by_status[container.StatusName]; ok {
		summary_by_status[container.StatusName] += 1
	} else {
		summary_by_status[container.StatusName] = 1
	}
	return summary_by_status

}

func createPoint(container *Container, hostname string, cluster_name string) *client.Point {
	tags := map[string]string{
		"id":           container.ID,
		"service":      container.Name,
		"host":         hostname,
		"cluster_name": cluster_name,
	}

	for key, value := range container.Tags {
		tags[key] = value
	}

	fields := map[string]interface{}{
		"address":     container.Address,
		"status":      int8(container.Status),
		"status_name": container.StatusName,
		"uptime":      container.Uptime,
	}

	//add rest measurements
	for name, value := range structs.Map(container.Stats) {
		parsedName := snaker.CamelToSnake(name)
		fields[parsedName] = value
	}

	return influx.GetInfluxClient().CreatePoint("armada", tags, fields)
}

func createSummary(host_summary map[string]int16, hostname string, cluster_name string) *client.Point {
	tags := map[string]string{
		"host":         hostname,
		"cluster_name": cluster_name,
	}

	fields := map[string]interface{}{}

	for key, value := range host_summary {
		fields[key] = value
	}

	return influx.GetInfluxClient().CreatePoint("armada_ship", tags, fields)
}

func getHostname() string {
	if hostname, err := utils.Config.String("hostname"); err == nil {
		return hostname
	}

	hostname, _ := os.Hostname()
	return hostname
}

func getClusterName() string {
	cluster_name, _ := utils.Config.String("armada_cluster_name")
	return cluster_name
}
