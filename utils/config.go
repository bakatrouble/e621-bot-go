package utils

import (
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type AwsConfig struct {
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type Config struct {
	BotToken string        `yaml:"bot_token"`
	ChatId   int64         `yaml:"chat_id"`
	ApiPort  int           `yaml:"api_port"`
	Interval time.Duration `yaml:"interval"`
	Redis    string        `yaml:"redis"`
	Aws      AwsConfig     `yaml:"aws"`
}

func ParseConfig(configFile string) (*Config, error) {
	config := &Config{}
	dat, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(dat, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
