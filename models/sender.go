package models

import (
	"os"
	"github.com/fatih/structs"
	"github.com/krise3k/armada-stats/utils/influx"
	"github.com/krise3k/armada-stats/utils"
)

func SendToInflux(container Container) {
	hostname := getHostname()
	batch := influx.CreateBatchPoints()
	fields := map[string]interface{}{
		"address": container.Address,
		"status": int8(container.Status),
		"uptime": container.Uptime,
	}

	tags := map[string]string{
		"id": container.ID,
		"serviceName": container.Name,
		"host": hostname,
	}

	for key, value := range (container.Tags) {
		tags[key] = value
	}
	//add Status measurement
	point := influx.CreatePoint("Status", tags, fields)
	batch.AddPoint(point)

	//add rest measurements
	for name, value := range (structs.Map(container.Stats)) {
		fields := map[string]interface{}{
			"value": value,
		}
		point := influx.CreatePoint(name, tags, fields)
		batch.AddPoint(point)
	}

	influx.Save(batch)
}

func getHostname() string {
	if hostname, err := utils.Config.String("hostname"); err == nil {
		return hostname
	}

	hostname, _ := os.Hostname()
	return hostname
}