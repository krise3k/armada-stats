package models

import (
	"os"
	"github.com/fatih/structs"
	"github.com/krise3k/armada-stats/utils/influx"
)

func SendToInflux(container Container) {
	batch := influx.CreateBatchPoints()
	hostname, _ := os.Hostname()

	fields := map[string]interface{}{
		"address": container.Address,
		"status": int8(container.Status),
		"uptime": container.Uptime,
	}

	tags := map[string]string{
		"id": container.ID,
		"name": container.Name,
		"hostname": hostname,
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




