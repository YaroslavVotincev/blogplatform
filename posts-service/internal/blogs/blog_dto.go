package blogs

import "github.com/google/uuid"

type BlogUpdateRequest struct {
	Title            string   `json:"title" validate:"required,min=2,max=50"`
	ShortDescription string   `json:"short_description" validate:"required,min=2,max=200"`
	Url              string   `json:"url" validate:"required,min=2,max=50,urlpath"`
	Status           string   `json:"status" validate:"required,oneof=draft public"`
	AcceptDonations  bool     `json:"accept_donations"`
	Categories       []string `json:"categories"`
}

type BlogUpdateTitleRequest struct {
	Title            string `json:"title" validate:"required,min=2,max=200"`
	ShortDescription string `json:"short_description"`
}

type BlogUpdateUrlRequest struct {
	Url string `json:"url" validate:"required,min=2,max=200"`
}

type BlogUpdateContentRequest struct {
	DataJson string `json:"data_json" validate:"required"`
	DataHtml string `json:"data_html"`
}

type BlogUpdateAcceptDonationsRequest struct {
	AcceptDonations bool `json:"accept_donations"`
}

type BlogUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft public"`
}

type SubscriptionUpdateInfoRequest struct {
	Title            string  `json:"title"`
	ShortDescription string  `json:"short_description"`
	PriceRub         float64 `json:"price_rub"`
	IsActive         bool    `json:"is_active"`
}

//type GrantSubscriptionServiceRequest struct {
//	SubscriptionId uuid.UUID `json:"subscription_id" validate:"required"`
//	UserId         uuid.UUID `json:"user_id" validate:"required"`
//}

type CreateGoalRequest struct {
	Type        string `json:"type" validate:"required,oneof=1 2 3 4"`
	Target      int    `json:"target" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type UpdateGoalRequest struct {
	Target      int    `json:"target" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type BlogStatsResponse struct {
	CommentsCount        int `json:"comments_count"`
	PostsCount           int `json:"posts_count"`
	FollowersCount       int `json:"followers_count"`
	PaidSubscribersCount int `json:"paid_subscribers_count"`
}

//type ConfirmDonationServiceRequest struct {
//	DonationId uuid.UUID `json:"donation_id" validate:"required"`
//	UserId     uuid.UUID `json:"user_id" validate:"required"`
//}

type MakeDonationRequest struct {
	Value   float64 `json:"value" validate:"required"`
	Comment string  `json:"comment"`
}

type GrantItemServiceRequest struct {
	ItemId   uuid.UUID `json:"item_id" validate:"required"`
	UserId   uuid.UUID `json:"user_id" validate:"required"`
	Value    float64   `json:"value" validate:"required"`
	Currency string    `json:"currency" validate:"required"`
}
