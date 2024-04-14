package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Cache struct {
		TTl int `yaml:"ttl"`
	} `yaml:"cache"`
	DB struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"db"`
}

const DefaultConfigPath = "./config/database-service/config.yml"

func ReadConfig() (config *Config, err error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = DefaultConfigPath
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	file, err := os.ReadFile(abs)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return
	}
	return config, nil
}
