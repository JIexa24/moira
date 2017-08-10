package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/moira-alert/moira-alert/database/redis"
	"github.com/moira-alert/moira-alert/logging"
	"github.com/moira-alert/moira-alert/metrics/graphite"
	"gopkg.in/yaml.v2"
)

type config struct {
	Redis    redisConfig    `yaml:"redis"`
	Cache    cacheConfig    `yaml:"cache"`
	Graphite graphiteConfig `yaml:"graphite"`
}

type cacheConfig struct {
	LogLevel        string `yaml:"log_level"`
	LogColor        string `yaml:"log_color"`
	LogFile         string `yaml:"log_file"`
	Listen          string `yaml:"listen"`
	RetentionConfig string `yaml:"retention-config"`
}

type redisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	DBID int    `yaml:"dbid"`
}

type graphiteConfig struct {
	URI      string `yaml:"uri"`
	Prefix   string `yaml:"prefix"`
	Interval int64  `yaml:"interval"`
}

func (graphiteConfig *graphiteConfig) getSettings() graphite.Config {
	return graphite.Config{
		URI:      graphiteConfig.URI,
		Prefix:   graphiteConfig.Prefix,
		Interval: graphiteConfig.Interval,
	}
}

func (config *redisConfig) getSettings() redis.Config {
	return redis.Config{
		Host: config.Host,
		Port: config.Port,
		DBID: config.DBID,
	}
}

func (config *cacheConfig) getLoggerSettings() logging.Config {
	return logging.Config{
		LogFile:  config.LogFile,
		LogColor: toBool(config.LogColor),
		LogLevel: config.LogLevel,
	}
}

func getDefault() config {
	return config{
		Redis: redisConfig{
			Host: "localhost",
			Port: "6379",
			DBID: 0,
		},
		Cache: cacheConfig{
			LogLevel:        "debug",
			LogFile:         "stdout",
			Listen:          ":2003",
			RetentionConfig: "storage-schemas.conf",
		},
		Graphite: graphiteConfig{
			URI:      "localhost:2003",
			Prefix:   "DevOps.Moira",
			Interval: 60,
		},
	}
}

func printDefaultConfig() {
	c := getDefault()
	d, _ := yaml.Marshal(&c)
	fmt.Println(string(d))
}

func readSettings(configFileName string) (*config, error) {
	c := getDefault()
	configYaml, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("Can't read file [%s] [%s]", configFileName, err.Error())
	}
	err = yaml.Unmarshal(configYaml, &c)
	if err != nil {
		return nil, fmt.Errorf("Can't parse config file [%s] [%s]", configFileName, err.Error())
	}
	return &c, nil
}

func toBool(str string) bool {
	switch strings.ToLower(str) {
	case "1", "true", "t", "yes", "y":
		return true
	}
	return false
}
