package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Values struct {
	JwtCodePhrase string `yaml:"jwt_code_phrase"`
	Cache         struct {
		TTl int `yaml:"ttl"`
	} `yaml:"cache"`
	DB struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"db"`
}

func (cfg Values) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
}

const DefaultConfigPath = "./config/database-service/config.yml"

func ReadConfig() (config *Values, err error) {
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
