package email

import (
	"bytes"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"html/template"
	"registration-service/pkg/filelogger"
	"registration-service/pkg/queuelogger"
)

const (
	EmailQueueConfigKey = "EMAIL_QUEUE"
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
	}, EmailQueueConfigKey)

	return service
}

func (s *Service) SendSignupConfirmEmail(login, email, code string) {
	data := SignupConfirmEmailData{
		Code: code,
	}

	t, _ := template.ParseFiles("./templates/signup.html")
	var tpl bytes.Buffer
	_ = t.Execute(&tpl, data)
	html := tpl.String()
	ToName := login

	s.sender.handleMessage(QueueMessage{
		From:    nil,
		Name:    nil,
		Subject: "Регистрация аккаунта",
		To:      email,
		ToName:  &ToName,
		Html:    html,
	})
}
