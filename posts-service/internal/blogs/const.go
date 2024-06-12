package blogs

import "encoding/json"

var defaultContentObject map[string]any = map[string]any{
	"type": "doc",
	"content": []map[string]any{
		{
			"type": "paragraph",
			"attrs": map[string]any{
				"textAlign": "left",
			},
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Начните писать...",
				},
			},
		},
	},
}

var defaultContentJsonBytes, _ = json.Marshal(defaultContentObject)
var defaultContentJson = string(defaultContentJsonBytes)
var defaultContentHtml = "<p>Начните писать...</p>"

const (
	BlogTypePersonal = "personal"
	BlogTypeThematic = "thematic"
)

const (
	BlogStatusDraft  = "draft"
	BlogStatusPublic = "public"
)

const (
	PostStatusDraft  = "draft"
	PostStatusPublic = "public"
)

const (
	PaymentItemTypeSubscription = "subscription"
	PaymentItemTypePost         = "post"
	PaymentItemTypeDonation     = "donation"
)

const (
	UserSubscriptionStatusRecurrent = "recurrent"
	UserSubscriptionStatusCancelled = "cancelled"
	UserSubscriptionStatusExpired   = "expired"
	UserSubscriptionStatusLifetime  = "lifetime"
)

const (
	DonationStatusNew       = "new"
	DonationStatusConfirmed = "confirmed"
)

const (
	CurrencyRub = "rub"
	CurrencyTon = "toncoin"
)
