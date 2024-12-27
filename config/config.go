package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"time"

	"github.com/pelicanplatform/pelicanobjectstager/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Initialize the zap logger for the "config" component
var log = logger.With(zap.String("component", "config"))

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	Pelican struct {
		BinaryPath string `mapstructure:"binary_path"`
	} `mapstructure:"pelican"`

	Staging struct {
		TempDestination string `mapstructure:"temp_destination"`
		Workers         int    `mapstructure:"workers"`
	}

	Database struct {
		Location               string        `mapstructure:"location"`
		RefreshInterval        time.Duration `mapstructure:"refresh_interval"`
		MaxRecordStaleDuration time.Duration `mapstructure:"max_record_stale_duration"`
	} `mapstructure:"database"`
}

var AppConfig Config

//go:embed resources/default_config.yaml
var defaultConfig []byte

func LoadConfig(configPath string) {
	viper.SetConfigType("yaml")

	// Load the embedded default configuration
	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		log.Fatal("Error loading embedded default configuration", zap.Error(err))
	}

	log.Info("Default configuration loaded successfully")

	// Check if additional configuration file exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, merge the additional configuration
		viper.SetConfigFile(configPath)
		if err := viper.MergeInConfig(); err != nil {
			log.Fatal("Error merging additional configuration file", zap.String("path", configPath), zap.Error(err))
		}
		log.Info("Additional configuration merged successfully", zap.String("path", configPath))
	} else if !os.IsNotExist(err) {
		// Other errors (e.g., permission issues)
		log.Fatal("Error checking additional configuration file", zap.String("path", configPath), zap.Error(err))
	}

	// Unmarshal the final configuration
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatal("Unable to decode configuration", zap.Error(err))
	}

	// Serialize the final configuration for logging
	configBytes, err := json.MarshalIndent(AppConfig, "", "  ")
	if err != nil {
		log.Error("Failed to serialize configuration for logging", zap.Error(err))
	} else {
		log.Info("Final configuration loaded", zap.String("config", string(configBytes)))
	}

	log.Info("Configuration loading complete")
}
