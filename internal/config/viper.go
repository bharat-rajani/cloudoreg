package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	configName = "configs"
	configType = "yaml"
)

// configPaths is the list of locations that will be searched for the configs file.
var configPaths = []string{"/etc/rainmaker-api/", "./configs/"}

// loadWithViper loads the configs using spf13/viper.
// Panic is allowed here because configs are crucial to the application.
func loadWithViper() *Config {
	// Specifying the configs file name and type to viper.
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// Adding configs paths to viper.
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Reading the configs file.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error in ReadInConfig: %w", err))
	}

	model := &Config{}
	// Unmarshalling into the model instance.
	if err := viper.Unmarshal(model, func(c *mapstructure.DecoderConfig) { c.TagName = configType }); err != nil {
		panic(fmt.Errorf("error in Unmarshal: %w", err))
	}

	return model
}
