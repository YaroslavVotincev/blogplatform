package robokassa

import "github.com/google/uuid"

type PaymentLinkRequest struct {
	ItemId      uuid.UUID `json:"item_id"`
	ItemType    string    `json:"item_type"`
	UserId      uuid.UUID `json:"user_id"`
	Sum         float64   `json:"sum"`
	Description string    `json:"description"`
}

type PaymentLinkResponse struct {
	Url string `json:"url"`
}
