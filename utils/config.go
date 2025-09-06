package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

type AwsConfig struct {
	Region    string `yaml:"region" binding:"required"`
	Bucket    string `yaml:"bucket" binding:"required"`
	AccessKey string `yaml:"access_key" binding:"required"`
	SecretKey string `yaml:"secret_key" binding:"required"`
}

type DestinationsConfig struct {
	Nsfw string `yaml:"nsfw" binding:"required"`
	Sfw  string `yaml:"sfw" binding:"required"`
}

type ApiConfig struct {
	Port int      `yaml:"port" binding:"required"`
	Keys []string `yaml:"keys" binding:"required"`
}

type Config struct {
	BotToken     string             `yaml:"bot_token" binding:"required"`
	ChatId       int64              `yaml:"chat_id" binding:"required"`
	Api          ApiConfig          `yaml:"api" binding:"required"`
	Interval     time.Duration      `yaml:"interval" binding:"required"`
	Redis        string             `yaml:"redis" binding:"required"`
	CacheDir     string             `yaml:"cache_dir" binding:"required"`
	Aws          AwsConfig          `yaml:"aws" binding:"required"`
	Destinations DestinationsConfig `yaml:"destinations" binding:"required"`
	Production   bool               `yaml:"production"`
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
