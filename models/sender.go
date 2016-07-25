package models

import (
	"os"
	"github.com/fatih/structs"
	"github.com/krise3k/armada-stats/utils"
	"github.com/krise3k/armada-stats/utils/influx"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/serenize/snaker"
)

func SendMetrics(containers Containers) {
	hostname := getHostname()
	batch := influx.CreateBatchPoints()

	for _, container := range containers.ContainerList {
		point := createPoint(container, hostname)
		batch.AddPoint(point)
	}

	influx.Save(batch)
}

func createPoint(container *Container, hostname string) *client.Point {
	tags := map[string]string{
		"id": container.ID,
		"service": container.Name,
		"host": hostname,
	}

	for key, value := range (container.Tags) {
		tags[key] = value
	}

	fields := map[string]interface{}{
		"address": container.Address,
		"status": int8(container.Status),
		"uptime": container.Uptime,
	}

	//add rest measurements
	for name, value := range (structs.Map(container.Stats)) {
		parsedName := snaker.CamelToSnake(name)
		fields[parsedName] = value
	}

	return influx.CreatePoint("armada", tags, fields)
}

func getHostname() string {
	if hostname, err := utils.Config.String("hostname"); err == nil {
		return hostname
	}

	hostname, _ := os.Hostname()
	return hostname
}