package server

import (
	"fmt"
)

type Config struct {
	Host                 string `env:"HOST" env-default:"0.0.0.0"`
	Port                 string `env:"PORT"`
	ServiceName          string `env:"SERVICE_NAME"`
	ConfigServiceHost    string `env:"CONFIG_SERVICE_HOST"`
	ConfigServicePort    string `env:"CONFIG_SERVICE_PORT"`
	ConfigUpdateInterval int    `env:"CONFIG_UPDATE_INTERVAL" env-default:"60"`

	DbHost     string `config-service:"DB_HOST"`
	DbName     string `config-service:"DB_NAME"`
	DbPassword string `config-service:"DB_PASSWORD"`
	DbPort     string `config-service:"DB_PORT"`
	DbUser     string `config-service:"DB_USER"`
}

func (cfg *Config) DbUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
	)
}

func (cfg *Config) ConfigServiceUrl() string {
	return fmt.Sprintf("http://%s:%s/api/v1/config/service", cfg.ConfigServiceHost, cfg.ConfigServicePort)
}
