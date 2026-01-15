package config

import (
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

	Debug bool `koanf:"debug"`
}

func GetConfig() (config Config, err error) {
	var k = koanf.New(".")

	// 1. Load the structured list from YAML
	if err = k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return config, err
	}

	// 2. Load Environment Variables for overrides (e.g. DEBUG=true)
	k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "__", ".")
	}), nil)

	err = k.Unmarshal("", &config)

	return config, err
}
