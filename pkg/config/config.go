package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	DBType     string `yaml:"db-type"`
	DBDsn      string `yaml:"db-dsn"`
	ListenAddr string `yaml:"listen-addr"`
	APIUrl     string `yaml:"api-url"`
	APIPrefix  string `yaml:"api-prefix"`
}

func Init(file string) (*Config, error) {
	cfgFile := "gh-pr.yml"
	if len(file) > 0 {
		cfgFile = file
	}
	log.Infof("Using config file: %s", cfgFile)
	fi, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err = yaml.NewDecoder(fi).Decode(&config); err != nil {
		return nil, err
	}
	if config.DBType != "sqlite3" {
		return nil, fmt.Errorf("unsupported db type: %s", config.DBType)
	}

	return &config, nil
}
