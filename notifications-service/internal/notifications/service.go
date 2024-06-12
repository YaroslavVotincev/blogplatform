package notifications

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	repository *Repository
	validate   *validator.Validate
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
		validate:   NewValidator(),
	}
}

func (s *Service) ValidateNotificationData(notification *Notification) error {
	switch notification.EventCode {
	case EventCodeAuthentication:
		var data AuthenticationEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodeRegistration:
		var data RegistrationEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodeSubscriptionAuthor:
		var data SubscriptionAuthorEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodeSubscriptionUser:
		var data SubscriptionUserEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodePostPaidAccessAuthor:
		var data PostPaidAccessAuthorEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodePostPaidAccessUser:
		var data PostPaidAccessUserEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodeDonationAuthor:
		var data DonationAuthorEventData
		return s.ValidateStructAndWrite(&data, notification)
	case EventCodeDonationUser:
		var data DonationUserEventData
		return s.ValidateStructAndWrite(&data, notification)
	default:
		return fmt.Errorf("unknown event code: %s", notification.EventCode)
	}
}

func (s *Service) ValidateStructAndWrite(data any, notification *Notification) error {
	err := json.Unmarshal(notification.DataBytes, &data)
	if err != nil {
		return err
	}
	err = s.validate.Struct(data)
	if err := s.validate.Struct(data); err != nil {
		return err
	}
	notification.Data = data
	return nil
}
