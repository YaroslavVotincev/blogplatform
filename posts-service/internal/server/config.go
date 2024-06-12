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

	FileQueue          string `config-service:"FILE_QUEUE"`
	FileGetEndpointUrl string `config-service:"FILE_GET_URL"`

	NotificationQueue string `config-service:"NOTIFICATION_QUEUE"`

	DbHost     string `config-service:"DB_HOST"`
	DbPort     string `config-service:"DB_PORT"`
	DbUser     string `config-service:"DB_USER"`
	DbPassword string `config-service:"DB_PASSWORD"`
	DbName     string `config-service:"DB_NAME"`

	MQHost     string `config-service:"MQ_HOST"`
	MQPort     string `config-service:"MQ_PORT"`
	MQUser     string `config-service:"MQ_USER"`
	MQPassword string `config-service:"MQ_PASSWORD"`
	LogQueue   string `config-service:"LOG_QUEUE"`

	BillingServiceUrl  string `config-service:"BILLING_SERVICE_URL"`
	CommentsServiceUrl string `config-service:"COMMENTS_SERVICE_URL"`
	UsersServiceUrl    string `config-service:"USERS_SERVICE_URL"`

	MainPageLikesRequirement    int `config-service:"MAIN_PAGE_LIKES_REQUIREMENT"`
	MainPageCommentsRequirement int `config-service:"MAIN_PAGE_COMMENTS_REQUIREMENT"`
	MainPageViewsRequirement    int `config-service:"MAIN_PAGE_VIEWS_REQUIREMENT"`
	MainPageDislikesRequirement int `config-service:"MAIN_PAGE_DISLIKES_REQUIREMENT"`

	DonationsRobokassaMinValue float64 `config-service:"DONATIONS_ROBOKASSA_MIN_VALUE"`
	DonationsToncoinMinValue   float64 `config-service:"DONATIONS_TONCOIN_MIN_VALUE"`
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

func (cfg *Config) MqUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.MQUser, cfg.MQPassword, cfg.MQHost, cfg.MQPort)
}
