package server

import (
	configservice "config-service/internal/config-service"
	"context"
	"fmt"
)

type Config struct {
	Host        string `env:"HOST" env-default:"0.0.0.0"`
	Port        string `env:"PORT"`
	ServiceName string `env:"SERVICE_NAME" env-default:"config_service"`

	DbHost     string `env:"DB_HOST"`
	DbPort     string `env:"DB_PORT"`
	DbName     string `env:"DB_NAME"`
	DbUser     string `env:"DB_USER"`
	DbPassword string `env:"DB_PASSWORD"`

	MQHost     string
	LogQueue   string
	MQPassword string
	MQPort     string
	MQUser     string
}

func (cfg *Config) DbUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)
}

func (cfg *Config) MqUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s",
		cfg.MQUser, cfg.MQPassword, cfg.MQHost, cfg.MQPort)
}

func (cfg *Config) UpdateConfigFromService(ctx context.Context, service *configservice.Service) error {
	settings, err := service.SettingsByService(ctx, cfg.ServiceName)
	if err != nil {
		return err
	}
	for _, s := range settings {
		switch s.Key {
		case "MQ_HOST":
			cfg.MQHost = s.Value
		case "MQ_PORT":
			cfg.MQPort = s.Value
		case "MQ_USER":
			cfg.MQUser = s.Value
		case "MQ_PASSWORD":
			cfg.MQPassword = s.Value
		case "LOG_QUEUE":
			cfg.LogQueue = s.Value
		}
	}
	return nil
}
