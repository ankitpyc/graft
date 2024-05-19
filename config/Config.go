package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	WALFilePath          string `json:"wal_log_file"`
	Port                 string `json:"port"`
	Host                 string `json:"host"`
	ClusterName          string `json:"cluster_name"`
	ClusterUUID          string `json:"cluster_id"`
	ServiceDiscoveryAddr string `json:"service_discovery_addr"`
}

func (conf *Config) LoadConfig(configFile string) (*Config, error) {
	var config Config
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewConfig() *Config {
	return &Config{}
}
