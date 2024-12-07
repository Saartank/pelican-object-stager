package config

import (
	"bytes"
	"os"

	_ "embed"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	Pelican struct {
		BinaryPath string `mapstructure:"binary_path"`
	} `mapstructure:"pelican"`
}

var AppConfig Config

//go:embed resources/default_config.yaml
var defaultConfig []byte

func LoadConfig(configPath string) {
	viper.SetConfigType("yaml")

	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		logrus.Fatalf("Error loading embedded default configuration: %v", err)
	}

	logrus.Info("Default configuration loaded successfully")

	// Check if additional configuration file exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, merge the additional configuration
		viper.SetConfigFile(configPath)
		if err := viper.MergeInConfig(); err != nil {
			logrus.Fatalf("Error merging additional configuration file: %v", err)
		}
		logrus.Infof("Additional configuration merged successfully from: %s", configPath)
	} else if !os.IsNotExist(err) {
		// Other errors (e.g., permission issues)
		logrus.Fatalf("Error checking additional configuration file: %v", err)
	}

	// Unmarshal the final configuration
	if err := viper.Unmarshal(&AppConfig); err != nil {
		logrus.Fatalf("Unable to decode configuration: %v", err)
	}

	logrus.Infof("Configuration loading complete")
}
