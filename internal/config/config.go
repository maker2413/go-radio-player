package config

import "github.com/spf13/viper"

type Config struct {
	StationName string `mapstructure:"STATION_NAME"`
}

func GetConfig() (config Config, err error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
