package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Members []Member `yaml:"members"`
}

type Member struct {
	RaftAddress string `yaml:"raftAddress"`
	HttpAddress string `yaml:"httpAddress"`
	NodeID string `yaml:"nodeID"`
}

func GetConfiguration() *Configuration {
	var configuration *Configuration
	yamlFile, err := ioutil.ReadFile("cluster.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &configuration)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return configuration
}