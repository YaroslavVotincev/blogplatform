package robokassa

import (
	"github.com/google/uuid"
	"time"
)

type Invoice struct {
	ID          int       `json:"id"`
	OutSum      float64   `json:"out_sum"`
	ItemId      uuid.UUID `json:"item_id"`
	ItemType    string    `json:"item_type"`
	UserId      uuid.UUID `json:"user_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"`
	PaymentLink string    `json:"payment_link"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

const (
	InvoiceStatusNew       = "new"
	InvoiceStatusConfirmed = "confirmed"
	InvoiceStatusExpired   = "expired"
	InvoiceStatusFailed    = "failed"
)

const (
	InvoiceItemTypeSubscription = "subscription"
	InvoiceItemTypePost         = "post"
	InvoiceItemTypeDonation     = "donation"
)

const (
	MerchantLoginConfigKey = "ROBOKASSA_MERCHANT_LOGIN"
	Password1ConfigKey     = "ROBOKASSA_PASSWORD1"
	Password2ConfigKey     = "ROBOKASSA_PASSWORD2"
	IsTestConfigKey        = "ROBOKASSA_IS_TEST"
	TestPassword1ConfigKey = "ROBOKASSA_TEST_PASSWORD1"
	TestPassword2ConfigKey = "ROBOKASSA_TEST_PASSWORD2"
)

const (
	PaymentLinkPrefixUrl = "https://auth.robokassa.ru/Merchant/Index/"
)

const (
	CurrencyRub = "rub"
	CurrencyTon = "toncoin"
)
