package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	GitHubToken string `mapstructure:"github_token"`
	DefaultUser string `mapstructure:"default_user"`
}

var AppConfig Config

func LoadConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	viper.AddConfigPath(home)
	viper.SetConfigName(".ghpm")
	viper.SetConfigType("yaml")

	// Set default values
	viper.SetDefault("default_user", "")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(&AppConfig)
}

func SaveConfig() error {
	return viper.WriteConfig()
}
