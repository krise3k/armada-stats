package utils

import (
	"io/ioutil"
	"log"
	"github.com/olebedev/config"
)

var Config *config.Config
var defaultConfigPath = "/etc/default/armada-stats.yml"


func InitConfig(path string) {
	defaultConfig := readConfig(defaultConfigPath)
	customConfig := readConfig(path)
	Config, _ = defaultConfig.Extend(customConfig)
}

func readConfig(path string) *config.Config {
	yamlConfig, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicf("Could not load config: %v", err)
	}

	cfg, err := config.ParseYaml(string(yamlConfig))
	if err != nil {
		log.Panicf("Could not parse config: %v", err)
	}
	return cfg
}
