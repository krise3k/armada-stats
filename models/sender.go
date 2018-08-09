package models

import (
	"github.com/fatih/structs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/krise3k/armada-stats/utils"
	"github.com/krise3k/armada-stats/utils/influx"
	"github.com/serenize/snaker"
	"os"
	"time"
	"github.com/krise3k/armada-stats/utils/kafka"
)

type ShipSummaryCounter map[string]int16

func SendMetrics(containers Containers) {
	hostname := getHostname()
	cluster_name := getClusterName()
	summary_by_status := ShipSummaryCounter{}
	var metrics []telegraf.Metric
	timestamp := time.Now().UTC()

	for _, container := range containers.ContainerList {
		if len(container.StatusName) > 0 {
			summary_by_status = updateShipSummary(summary_by_status, container)
		}
		metrics = append(metrics, createMetric(container, hostname, cluster_name, timestamp))
	}

	if len(summary_by_status) > 0 {
		metrics = append(metrics, createSummary(summary_by_status, hostname, cluster_name, timestamp))
	}

	if useKafka, _ := utils.Config.Bool("send_to_kafka"); useKafka == true {
		sendToKafka(metrics)
	}

	if useInflux, _ := utils.Config.Bool("send_to_influx"); useInflux == true {
		sendToInflux(metrics)
	}
}

func sendToKafka(metrics []telegraf.Metric) {
	kafkaClient := kafka.GetKafkaClient()
	err := kafkaClient.Write(metrics)
	if err != nil {
		utils.GetLogger().WithError(err).Error("Error sending to kafka")
	}
}

func sendToInflux(metrics []telegraf.Metric) {
	influxClient := influx.GetInfluxClient()
	batch := influxClient.CreateBatchPoints()
	for _, m := range metrics {
		batch.AddPoint(influxClient.CreatePoint(m.Name(), m.Tags(), m.Fields()))
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

func createMetric(container *Container, hostname string, cluster_name string, timestamp time.Time) telegraf.Metric {
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
	m, _ := metric.New("armada", tags, fields, timestamp)

	return m

}

func createSummary(host_summary map[string]int16, hostname string, cluster_name string, timestamp time.Time) telegraf.Metric {
	tags := map[string]string{
		"host":         hostname,
		"cluster_name": cluster_name,
	}

	fields := map[string]interface{}{}

	for key, value := range host_summary {
		fields[key] = value
	}
	m, _ := metric.New("armada_ship", tags, fields, timestamp)

	return m
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
