package notifications

import (
	"auth-service/pkg/filelogger"
	"auth-service/pkg/queuelogger"
	"encoding/json"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"time"
)

const (
	QueueConfigKey          = "NOTIFICATIONS_QUEUE"
	EventCodeAuthentication = "AUTHENTICATION"
)

type Service struct {
	sender      *Sender
	fileLogger  *filelogger.FileLogger
	queueLogger *queuelogger.RemoteLogger
}

func NewService(sender *Sender, cfgService *configService.ConfigServiceManager,
	fileLogger *filelogger.FileLogger,
	queueLogger *queuelogger.RemoteLogger) *Service {

	service := &Service{
		sender:      sender,
		fileLogger:  fileLogger,
		queueLogger: queueLogger,
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		sender.queue = ss.Value
	}, QueueConfigKey)

	return service
}

func (s *Service) Authentication(userId string) {
	loggingMap := map[string]any{}
	obj := AuthenticationEventData{
		At: time.Now().UTC(),
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal authentication event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodeAuthentication, body)
	if err != nil {
		loggingMap["message"] = "failed to send authentication event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}
