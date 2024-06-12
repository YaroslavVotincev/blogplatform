package notifications

import "time"

type SubscriptionAuthorEventData struct {
	At             time.Time `json:"at" validate:"required"`
	BlogId         string    `json:"blog_id" validate:"required"`
	SubscriptionId string    `json:"subscription_id" validate:"required"`
	FromUserId     string    `json:"from_user_id" validate:"required"`
	IncomeValue    float64   `json:"income_value" validate:"required"`
	IncomeCurrency string    `json:"income_currency" validate:"required"`
}

type SubscriptionUserEventData struct {
	At              time.Time `json:"at" validate:"required"`
	BlogId          string    `json:"blog_id" validate:"required"`
	SubscriptionId  string    `json:"subscription_id" validate:"required"`
	PaymentValue    float64   `json:"payment_value" validate:"required"`
	PaymentCurrency string    `json:"payment_currency" validate:"required"`
}

type PostPaidAccessAuthorEventData struct {
	At             time.Time `json:"at" validate:"required"`
	BlogId         string    `json:"blog_id" validate:"required"`
	PostId         string    `json:"post_id" validate:"required"`
	FromUserId     string    `json:"from_user_id" validate:"required"`
	IncomeValue    float64   `json:"income_value" validate:"required"`
	IncomeCurrency string    `json:"income_currency" validate:"required"`
}

type PostPaidAccessUserEventData struct {
	At              time.Time `json:"at" validate:"required"`
	BlogId          string    `json:"blog_id" validate:"required"`
	PostId          string    `json:"post_id" validate:"required"`
	PaymentValue    float64   `json:"payment_value" validate:"required"`
	PaymentCurrency string    `json:"payment_currency" validate:"required"`
}

type DonationAuthorEventData struct {
	At             time.Time `json:"at" validate:"required"`
	BlogId         string    `json:"blog_id" validate:"required"`
	FromUserId     string    `json:"from_user_id" validate:"required"`
	Comment        string    `json:"comment"`
	IncomeValue    float64   `json:"income_value" validate:"required"`
	IncomeCurrency string    `json:"income_currency" validate:"required"`
}

type DonationUserEventData struct {
	At              time.Time `json:"at" validate:"required"`
	BlogId          string    `json:"blog_id" validate:"required"`
	PaymentValue    float64   `json:"payment_value" validate:"required"`
	PaymentCurrency string    `json:"payment_currency" validate:"required"`
}
