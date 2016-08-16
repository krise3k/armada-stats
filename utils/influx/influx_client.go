package influx

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/krise3k/armada-stats/models/armada"
	"github.com/krise3k/armada-stats/utils"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

const initialTriggerToFindNewInflux = 5
const maxTriggerToFindNewInflux = 1000

var influxClient *InfluxClient

type InfluxClient struct {
	Client                      client.Client
	Mu                          sync.Mutex
	host                        string
	port                        int
	errorCounter                int16
	errorsToTrigerFindingInflux int16
}

func initInfluxClient() {
	host, port := getInfluxAddr()
	influxClient = NewInfluxClient(host, port)
}
func GetInfluxClient() *InfluxClient {
	if influxClient == nil {
		initInfluxClient()
	}

	return influxClient
}

func getInfluxAddr() (string, int) {
	armadaShips, _ := utils.Config.List("armada_cluster_with_influx")
	armadaPort, _ := utils.Config.String("armada_port")
	influxServiceName, _ := utils.Config.String("influx_service_name")
	for _, host := range armadaShips {
		armadaClient := armada.NewArmadaClient(host.(string), armadaPort)

		influxServices := armadaClient.GetContainerListByServiceName(influxServiceName)
		if len(influxServices) == 0 {
			continue
		}

		addr := strings.SplitN(influxServices[0].Address, ":", 2)
		host := addr[0]
		port, _ := strconv.Atoi(addr[1])
		return host, port
	}
	utils.GetLogger().Panic("Could not get influxdb addres from armada")

	return "", 0
}

func initClient(host string, port int) client.Client {
	db, _ := utils.Config.String("influx_database")
	user, _ := utils.Config.String("influx_user")
	password, _ := utils.Config.String("influx_password")

	addr := fmt.Sprintf("http://%s", host)
	if port != 80 {
		addr += fmt.Sprintf(":%d", port)
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: password,
		Timeout:  10 * time.Second,
	})

	if err != nil {
		utils.GetLogger().WithError(err).Panic("Cannot connect to influx")
	}

	defer c.Close()

	dbCreateQuery := fmt.Sprintf("CREATE DATABASE %s", db)
	q := client.NewQuery(dbCreateQuery, "", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		utils.GetLogger().Info(response.Results)
	}
	return c
}

func NewInfluxClient(host string, port int) *InfluxClient {
	c := initClient(host, port)
	return &InfluxClient{
		host:                        host,
		port:                        port,
		Client:                      c,
		errorCounter:                0,
		errorsToTrigerFindingInflux: initialTriggerToFindNewInflux,
	}
}

func (c *InfluxClient) CreateBatchPoints() client.BatchPoints {
	db, _ := utils.Config.String("influx_database")
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: "s",
	})

	if err != nil {
		utils.GetLogger().WithError(err).Panic("Error creating batch")
	}
	return bp
}

func (c *InfluxClient) CreatePoint(name string, tags map[string]string, fields map[string]interface{}) *client.Point {
	pt, err := client.NewPoint(name, tags, fields, time.Now().UTC())
	if err != nil {
		utils.GetLogger().WithError(err).Error("Error sending to influx")
	}

	return pt
}

func (c *InfluxClient) Save(points client.BatchPoints) {
	err := c.Client.Write(points)
	if err == nil {
		return
	}
	c.errorCounter++
	utils.GetLogger().WithError(err).Error("Error sending to influx")
	utils.GetLogger().Debugf("Influx error counter %d", c.errorCounter)
	if c.errorCounter%c.errorsToTrigerFindingInflux == 0 {
		c.findNewInfluxdbAddress()
	}

}

func (c *InfluxClient) findNewInfluxdbAddress() {
	c.Mu.Lock()
	utils.GetLogger().Infof("Start finding new influxd")
	host, port := getInfluxAddr()
	c.errorCounter = 0

	if host == c.host && port == c.port {
		c.errorsToTrigerFindingInflux = int16(math.Min(float64(2*c.errorsToTrigerFindingInflux), float64(maxTriggerToFindNewInflux)))
		c.Mu.Unlock()
		utils.GetLogger().Warning("Cannot find new influxdb address")
		return
	}
	utils.GetLogger().Infof("New influxdb address %s:%d has been found", host, port)
	c.Client = initClient(host, port)
	c.host = host
	c.port = port
	c.errorsToTrigerFindingInflux = initialTriggerToFindNewInflux
	c.Mu.Unlock()
}
