package config

import (
	"errors"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Station struct {
	StationName string `koanf:"name"`
	StreamURL   string `koanf:"url"`
}

type Config struct {
	Stations []Station `koanf:"stations"`

	MaxDisplayedTitleSize int  `koanf:"maxDisplayedTitleSize"`
	Debug                 bool `koanf:"debug"`
}

func GetConfig() (config Config, err error) {
	var k = koanf.New(".")

	if err = k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return config, err
	}

	err = k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "__", ".")
	}), nil)
	if err != nil {
		return Config{}, err
	}

	err = k.Unmarshal("", &config)

	if len(config.Stations) == 0 {
		return config, errors.New("no stations configured")
	}

	return config, err
}
