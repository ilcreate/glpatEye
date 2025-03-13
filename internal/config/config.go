package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Gitlab struct {
		BaseUrl            string `yaml:"base_url"`
		Pattern            string `yaml:"pattern"`
		ResponseObjectSize int    `yaml:"responseObjectsSize"`
		PoolSize           int    `yaml:"poolSize"`
		Cron               string `yaml:"cron"`
	} `yaml:"gitlab"`
}

func (c *Config) loadConfig(path string) error {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		err := fmt.Errorf("Error during read a config file: %w", err)
		log.Default().Output(2, err.Error())
		log.Printf("Use environment variables, because config file isn't set.")
		return err
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return fmt.Errorf("error during unmarshal yaml file; %v", err)
	}
	return nil
}

func (c *Config) InitAppConfig(path string) {
	err := c.loadConfig(path)
	if err != nil {
		log.Printf("failed to load configuration; %v", err)
	}
}

func (c Config) DefaultConfig(envKey string, defaultValue interface{}) interface{} {
	if value, exists := os.LookupEnv(envKey); exists {
		switch defaultValue.(type) {
		case string:
			return value
		case int:
			if intValue, err := strconv.Atoi(value); err == nil {
				return intValue
			} else {
				log.Printf("invalid value for %s: %v. Using default: %v", envKey, err, defaultValue)
			}
		case bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				return boolValue
			} else {
				log.Printf("invalid value for %s: %v. Using default: %v", envKey, err, defaultValue)
			}
		default:
			log.Printf("unsupported type of variable: %s. Using default: %v", envKey, defaultValue)
		}
	}
	return defaultValue
}
