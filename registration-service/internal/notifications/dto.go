package notifications

import "time"

type AuthenticationEventData struct {
	At time.Time `json:"at" validate:"required"`
}

type RegistrationEventData struct {
	At    time.Time `json:"at" validate:"required"`
	Email string    `json:"email" validate:"required"`
	Login string    `json:"login" validate:"required"`
}
