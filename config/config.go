package config

import "github.com/spf13/viper"

type Config struct {
	TopicName                string `mapstructure:"TOPIC_NAME"`
	KafkaUri                 string `mapstructure:"KAFKA_URI"`
	KafkaCAFilePath          string `mapstructure:"KAFKA_CA_FILE_PATH"`
	KafkaUsername            string `mapstructure:"KAFKA_USERNAME"`
	KafkaPassword            string `mapstructure:"KAFA_PASSWORD"`
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

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}
	return config, nil
}
