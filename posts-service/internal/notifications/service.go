package notifications

import (
	"encoding/json"
	configService "posts-service/pkg/config-client"
	"posts-service/pkg/filelogger"
	"posts-service/pkg/queuelogger"
	"time"
)

const (
	QueueConfigKey                = "NOTIFICATIONS_QUEUE"
	EventCodeSubscriptionAuthor   = "SUBSCRIPTION_AUTHOR"
	EventCodeSubscriptionUser     = "SUBSCRIPTION_USER"
	EventCodePostPaidAccessAuthor = "POST_PAID_ACCESS_AUTHOR"
	EventCodePostPaidAccessUser   = "POST_PAID_ACCESS_USER"
	EventCodeDonationAuthor       = "DONATION_AUTHOR"
	EventCodeDonationUser         = "DONATION_USER"
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

func (s *Service) SubscriptionAuthor(userId, blogId, subscriptionId, fromUserId string, incomeValue float64, incomeCurrency string) {
	loggingMap := map[string]any{}
	obj := SubscriptionAuthorEventData{
		At:             time.Now().UTC(),
		BlogId:         blogId,
		SubscriptionId: subscriptionId,
		FromUserId:     fromUserId,
		IncomeValue:    incomeValue,
		IncomeCurrency: incomeCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal SUBSCRIPTION_AUTHOR event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodeSubscriptionAuthor, body)
	if err != nil {
		loggingMap["message"] = "failed to send SUBSCRIPTION_AUTHOR event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) SubscriptionUser(userId, blogId, subscriptionId string, paymentValue float64, paymentCurrency string) {
	loggingMap := map[string]any{}
	obj := SubscriptionUserEventData{
		At:              time.Now().UTC(),
		BlogId:          blogId,
		SubscriptionId:  subscriptionId,
		PaymentValue:    paymentValue,
		PaymentCurrency: paymentCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal SUBSCRIPTION_USER event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodeSubscriptionUser, body)
	if err != nil {
		loggingMap["message"] = "failed to send SUBSCRIPTION_USER event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) PostPaidAccessAuthor(userId, blogId, postId, fromUserId string, incomeValue float64, incomeCurrency string) {
	loggingMap := map[string]any{}
	obj := PostPaidAccessAuthorEventData{
		At:             time.Now().UTC(),
		BlogId:         blogId,
		PostId:         postId,
		FromUserId:     fromUserId,
		IncomeValue:    incomeValue,
		IncomeCurrency: incomeCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal POST_PAID_ACCESS_AUTHOR event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodePostPaidAccessAuthor, body)
	if err != nil {
		loggingMap["message"] = "failed to send POST_PAID_ACCESS_AUTHOR event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) PostPaidAccessUser(userId, blogId, postId string, paymentValue float64, paymentCurrency string) {
	loggingMap := map[string]any{}
	obj := PostPaidAccessUserEventData{
		At:              time.Now().UTC(),
		BlogId:          blogId,
		PostId:          postId,
		PaymentValue:    paymentValue,
		PaymentCurrency: paymentCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal POST_PAID_ACCESS_USER event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodePostPaidAccessUser, body)
	if err != nil {
		loggingMap["message"] = "failed to send POST_PAID_ACCESS_USER event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) DonationAuthor(userId, blogId, fromUserId, comment string, incomeValue float64, incomeCurrency string) {
	loggingMap := map[string]any{}
	obj := DonationAuthorEventData{
		At:             time.Now().UTC(),
		BlogId:         blogId,
		FromUserId:     fromUserId,
		Comment:        comment,
		IncomeValue:    incomeValue,
		IncomeCurrency: incomeCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal DONATION_AUTHOR event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodeDonationAuthor, body)
	if err != nil {
		loggingMap["message"] = "failed to send DONATION_AUTHOR event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}

func (s *Service) DonationUser(userId, blogId string, paymentValue float64, paymentCurrency string) {
	loggingMap := map[string]any{}
	obj := DonationUserEventData{
		At:              time.Now().UTC(),
		BlogId:          blogId,
		PaymentValue:    paymentValue,
		PaymentCurrency: paymentCurrency,
	}
	body, err := json.Marshal(obj)
	if err != nil {
		loggingMap["message"] = "failed to marshal DONATION_USER event data"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
	err = s.sender.publishMessage(userId, EventCodeDonationUser, body)
	if err != nil {
		loggingMap["message"] = "failed to send DONATION_USER event message to notification queue"
		loggingMap["error"] = err.Error()
		s.fileLogger.Error("error occurred", loggingMap)
		_ = s.queueLogger.Error(nil, loggingMap)
	}
}
