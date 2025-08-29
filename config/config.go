package config

import (
	"context"

	"github.com/spf13/viper"
)

type Config struct {
	RestServer restServer `mapstructure:"restServer"`
	Postgres   postgres   `mapstructure:"postgres"`
	Redis      redis      `mapstructure:"redis"`
}

type restServer struct {
	Port int `mapstructure:"port"`
}

type postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
}

type redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

var config *Config

func ResolveConfigFromFile(ctx context.Context, configPath string) (*Config, error) {
	// Read Configuration File
	viper.SetConfigName(configPath)
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetConfig() *Config {
	return config
}
