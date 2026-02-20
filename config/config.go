package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	TopicName                string `mapstructure:"TOPIC_NAME"`
	KafkaUri                 string `mapstructure:"KAFKA_URI"`
	KafkaCAFilePath          string `mapstructure:"KAFKA_CA_FILE_PATH"`
	KafkaUsername            string `mapstructure:"KAFKA_USERNAME"`
	KafkaPassword            string `mapstructure:"KAFKA_PASSWORD"`
	KafkaConsumerGroup       string `mapstructure:"KAFKA_CONSUMER_GROUP"`
	SchemaRegistryURL        string `mapstructure:"SCHEMA_REGISTRY_URL"`
	SchemaRegistryUsername   string `mapstructure:"SCHEMA_REGISTRY_USERNAME"`
	SchemaRegistryPassword   string `mapstructure:"SCHEMA_REGISTRY_PASSWORD"`
	ServiceName              string `mapstructure:"SERVICE_NAME"`
	OTELExporterOTLPEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELExporterOTLPHeaders  string `mapstructure:"OTEL_EXPORTER_OTLP_HEADERS"`
	OTELResourceAttreibutes  string `mapstructure:"OTEL_RESOURCE_ATTRIBUTES"`
	DBSource                 string `mapstructure:"DB_SOURCE"`
	ClerkKey                 string `mapstructure:"CLERK_KEY"`
	LogFilePath              string `mapstructure:"LOG_FILE_PATH"`
	LokiURL                  string `mapstructure:"LOKI_URL"`
	SyslogAddress            string `mapstructure:"SYSLOG_ADDRESS"`
	SyslogNetwork            string `mapstructure:"SYSLOG_NETWORK"`
}

// LoadConfig loads configuration from environment variables
// Optionally loads from .env file if path is provided and file exists
func LoadConfig(path string) (config Config, err error) {
	// Bind all environment variables
	viper.AutomaticEnv()

	// Explicitly bind each config key to its environment variable
	_ = viper.BindEnv("TOPIC_NAME")
	_ = viper.BindEnv("KAFKA_URI")
	_ = viper.BindEnv("KAFKA_CA_FILE_PATH")
	_ = viper.BindEnv("KAFKA_USERNAME")
	_ = viper.BindEnv("KAFKA_PASSWORD")
	_ = viper.BindEnv("KAFKA_CONSUMER_GROUP")
	_ = viper.BindEnv("SCHEMA_REGISTRY_URL")
	_ = viper.BindEnv("SCHEMA_REGISTRY_USERNAME")
	_ = viper.BindEnv("SCHEMA_REGISTRY_PASSWORD")
	_ = viper.BindEnv("SERVICE_NAME")
	_ = viper.BindEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	_ = viper.BindEnv("OTEL_EXPORTER_OTLP_HEADERS")
	_ = viper.BindEnv("OTEL_RESOURCE_ATTRIBUTES")
	_ = viper.BindEnv("DB_SOURCE")
	_ = viper.BindEnv("CLERK_KEY")
	_ = viper.BindEnv("LOG_FILE_PATH")
	_ = viper.BindEnv("LOKI_URL")
	_ = viper.BindEnv("SYSLOG_ADDRESS")
	_ = viper.BindEnv("SYSLOG_NETWORK")

	// Unmarshal into config struct (from env vars and/or config file)
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
