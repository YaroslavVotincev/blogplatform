package server

import "fmt"

type Config struct {
	Host                 string `env:"HOST" env-default:"0.0.0.0"`
	Port                 string `env:"PORT" env-default:"8000"`
	ServiceName          string `env:"SERVICE_NAME"`
	ConfigServiceHost    string `env:"CONFIG_SERVICE_HOST"`
	ConfigServicePort    int    `env:"CONFIG_SERVICE_PORT"`
	ConfigUpdateInterval int    `env:"CONFIG_UPDATE_INTERVAL" env-default:"60"`

	FileQueue string `config-service:"FILE_QUEUE"`

	LogQueue   string `config-service:"LOG_QUEUE"`
	MqHost     string `config-service:"MQ_HOST"`
	MqPort     string `config-service:"MQ_PORT"`
	MqUser     string `config-service:"MQ_USER"`
	MqPassword string `config-service:"MQ_PASSWORD"`
}

func (cfg *Config) ConfigServiceUrl() string {
	return fmt.Sprintf("http://%s:%d/api/v1/config/service", cfg.ConfigServiceHost, cfg.ConfigServicePort)
}

func (cfg *Config) MqUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", cfg.MqUser, cfg.MqPassword, cfg.MqHost, cfg.MqPort)
}
