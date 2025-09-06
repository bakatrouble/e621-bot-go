package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

type AwsConfig struct {
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type DestinationsConfig struct {
	Nsfw string `yaml:"nsfw"`
	Sfw  string `yaml:"sfw"`
}

type Config struct {
	BotToken     string             `yaml:"bot_token"`
	ChatId       int64              `yaml:"chat_id"`
	SharedChatId int64              `yaml:"shared_chat_id"`
	ApiPort      int                `yaml:"api_port"`
	Interval     time.Duration      `yaml:"interval"`
	Redis        string             `yaml:"redis"`
	CacheDir     string             `yaml:"cache_dir"`
	Aws          AwsConfig          `yaml:"aws"`
	Destinations DestinationsConfig `yaml:"destinations"`
}

func ParseConfig(configFile string) (*Config, error) {
	config := &Config{}
	var dat []byte
	var err error
	if dat, err = os.ReadFile(configFile); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(dat, config); err != nil {
		return nil, err
	}

	if config.CacheDir, err = filepath.Abs(config.CacheDir); err != nil {
		return nil, err
	}

	if err = os.MkdirAll(config.CacheDir, os.ModePerm); err != nil {
		return nil, err
	}

	return config, nil
}
