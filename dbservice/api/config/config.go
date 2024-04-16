package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type Values struct {
	JwtCodePhrase string `yaml:"jwt_code_phrase"`
	Cache         struct {
		TTlMilli  int    `yaml:"ttl"`
		KeyPrefix string `yaml:"key_prefix"`
	} `yaml:"cache"`
	DB struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"db"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Db       int    `yaml:"db"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}

func (cfg Values) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
}

func (cfg Values) GetCacheTTL() time.Duration {
	return time.Duration(cfg.Cache.TTlMilli) * time.Millisecond
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
