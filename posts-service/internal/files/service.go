package files

import (
	configService "posts-service/pkg/config-client"
	"posts-service/pkg/filelogger"
	"posts-service/pkg/queuelogger"
)

const (
	QueueConfigKey    = "FILE_QUEUE"
	EndpointConfigKey = "FILE_GET_URL"
	SetMethod         = "set"
	DeleteMethod      = "delete"
)

type Service struct {
	sender *Sender

	fileGetEndpoint string

	fileLogger  *filelogger.FileLogger
	queueLogger *queuelogger.RemoteLogger
}

func NewService(sender *Sender, fileGetEndpoint string, cfgService *configService.ConfigServiceManager,
	fileLogger *filelogger.FileLogger,
	queueLogger *queuelogger.RemoteLogger) *Service {

	service := &Service{
		sender:          sender,
		fileGetEndpoint: fileGetEndpoint,
		fileLogger:      fileLogger,
		queueLogger:     queueLogger,
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		sender.queue = ss.Value
	}, QueueConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.fileGetEndpoint = ss.Value
	}, EndpointConfigKey)

	return service
}

func (s *Service) SendFile(id string, body []byte) {
	err := s.sender.publishMessage(id, SetMethod, body)
	if err != nil {
		loggingMap := map[string]any{}
		loggingMap["message"] = "failed to send set file message to queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) DeleteFile(id string) {
	err := s.sender.publishMessage(id, DeleteMethod, nil)
	if err != nil {
		loggingMap := map[string]any{}
		loggingMap["message"] = "failed to send delete file message to queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) GetFileEndpointUrl() string {
	return s.fileGetEndpoint
}
