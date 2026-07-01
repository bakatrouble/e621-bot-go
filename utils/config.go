package utils

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	_ "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type AwsConfig struct {
	Region    string `koanf:"region"`
	Bucket    string `koanf:"bucket"`
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
}

type DestinationsConfig struct {
	Nsfw string `koanf:"nsfw"`
	Sfw  string `koanf:"sfw"`
}

type ApiConfig struct {
	Bind string   `koanf:"bind"`
	Keys []string `koanf:"keys"`
}

type Config struct {
	BotToken     string             `koanf:"bot_token"`
	ChatId       int64              `koanf:"chat_id"`
	Api          ApiConfig          `koanf:"api"`
	Interval     time.Duration      `koanf:"interval"`
	Redis        string             `koanf:"redis"`
	CacheDir     string             `koanf:"cache_dir"`
	Aws          AwsConfig          `koanf:"aws"`
	Destinations DestinationsConfig `koanf:"destinations"`
	Production   bool               `koanf:"production"`
}

func ParseConfig(configFile string) (*Config, error) {
	var err error

	k := koanf.New(".")
	if configFile != "" {
		if err = k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			panic(err)
		}
	}
	if err = k.Load(env.Provider(".", env.Opt{
		TransformFunc: func(k, v string) (string, any) {
			return strings.ToLower(k), v
		},
	}), nil); err != nil {
		panic(err)
	}

	config := Config{}
	k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf"})

	if config.BotToken == "" {
		panic("bot_token is not set in configuration")
	}

	if config.ChatId == 0 {
		panic("chat_id is not set in configuration")
	}

	if config.Interval == 0 {
		panic("interval is not set in configuration")
	}

	if config.Redis == "" {
		panic("redis is not set in configuration")
	}

	if config.CacheDir == "" {
		config.CacheDir = "cache"
	}

	if config.CacheDir, err = filepath.Abs(config.CacheDir); err != nil {
		return nil, err
	}

	if err = os.MkdirAll(config.CacheDir, os.ModePerm); err != nil {
		return nil, err
	}

	return &config, nil
}
