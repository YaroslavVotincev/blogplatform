package notifications

import "time"

type AuthenticationEventData struct {
	At time.Time `json:"at" validate:"required"`
}
