package server

import (
	"fmt"
)

type Config struct {
	Host                 string `env:"HOST" env-default:"0.0.0.0"`
	Port                 string `env:"PORT"`
	ServiceName          string `env:"SERVICE_NAME"`
	ConfigServiceHost    string `env:"CONFIG_SERVICE_HOST"`
	ConfigServicePort    int    `env:"CONFIG_SERVICE_PORT"`
	ConfigUpdateInterval int    `env:"CONFIG_UPDATE_INTERVAL" env-default:"60"`

	CorsAllowOrigins         string `config-service:"ALLOWED_HOSTS"`
	AuthServiceHost          string `config-service:"AUTHENTICATION_SERVICE"`
	RegistrationServiceHost  string `config-service:"REGISTRATION_SERVICE"`
	ConfServiceHost          string `config-service:"CONFIG_SERVICE"`
	UsersServiceHost         string `config-service:"USERS_SERVICE"`
	EmailServiceHost         string `config-service:"EMAIL_SERVICE"`
	PostsServiceUrl          string `config-service:"POSTS_SERVICE"`
	HistoryLogsServiceHost   string `config-service:"HISTORY_LOGS_SERVICE"`
	HistoryLogsConsumerHost  string `config-service:"HISTORY_LOGS_CONSUMER"`
	CommentsServiceHost      string `config-service:"COMMENTS_SERVICE"`
	BillingServiceHost       string `config-service:"BILLING_SERVICE"`
	NotificationsServiceHost string `config-service:"NOTIFICATIONS_SERVICE"`
}

//func (cfg *Config) AllowOrigins() []string {
//	array := strings.Split(cfg.CorsAllowOrigins, ",")
//	result := make([]string, 0)
//	for i := 0; i < len(array); i++ {
//		array[i] = strings.TrimSpace(array[i])
//		if array[i] != "" {
//			result = append(result, array[i])
//		}
//	}
//	return result
//}

func (cfg *Config) ConfigServiceUrl() string {
	return fmt.Sprintf("http://%s:%d/api/v1/config/service", cfg.ConfigServiceHost, cfg.ConfigServicePort)
}
