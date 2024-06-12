package notifications

import (
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	ID        uuid.UUID `json:"id"`
	EventCode string    `json:"event_code"`
	UserID    uuid.UUID `json:"user_id"`
	Seen      bool      `json:"seen"`
	Data      any       `json:"data"`
	DataBytes []byte    `json:"-"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

const (
	EventCodeAuthentication       = "AUTHENTICATION"
	EventCodeRegistration         = "REGISTRATION"
	EventCodeSubscriptionAuthor   = "SUBSCRIPTION_AUTHOR"
	EventCodeSubscriptionUser     = "SUBSCRIPTION_USER"
	EventCodePostPaidAccessAuthor = "POST_PAID_ACCESS_AUTHOR"
	EventCodePostPaidAccessUser   = "POST_PAID_ACCESS_USER"
	EventCodeDonationAuthor       = "DONATION_AUTHOR"
	EventCodeDonationUser         = "DONATION_USER"
)
