package config

import "github.com/confluentinc/confluent-kafka-go/kafka"

// Config represents the configs model.
type Config struct {
	// Application is the model of application configs.
	Application struct {
		Name      string `yaml:"name"`
		LogLevel  string `yaml:"logLevel"`
		LogPretty bool   `yaml:"logPretty"`
	} `yaml:"application"`

	// HTTPServer is the model of the HTTP Server configs.
	HTTPServer struct {
		// Addr is the address of the HTTP server.
		Addr string `yaml:"addr"`
	} `yaml:"httpServer"`

	RHSMService RHSMService `yaml:"rhsmService"`

	SourcesService SourcesService `yaml:"sourcesService"`

	Kafka KafkaConfig `yaml:"kafka"`
}

type SourcesService struct {
	URL string `yaml:"url"`
	PSK string `yaml:"psk"`
}

type RHSMService struct {
	URL string `yaml:"url"`
}

type KafkaConfig struct {
	BootstrapServers     string           `yaml:"server"`
	SourceEventTopic     string           `yaml:"eventTopic"`
	SourceStatusTopic    string           `yaml:"statusTopic"`
	SourceEventConsumer  *kafka.ConfigMap `yaml:"consumer"`
	SourceStatusProducer *kafka.ConfigMap `yaml:"producer"`
}

// Load loads and returns the config value.
func Load() *Config {
	return loadWithViper()
}

// LoadMock provides a mock instance of the config for testing purposes.
func LoadMock() *Config {
	cfg := &Config{}

	cfg.Application.Name = "rainmaker-mock-application"
	cfg.Application.LogLevel = "trace"
	cfg.Application.LogPretty = false

	cfg.HTTPServer.Addr = "localhost:8080"

	return cfg
}
