package blogs

import "github.com/google/uuid"

type PostUpdateRequest struct {
	Title            string     `json:"title" validate:"required,min=2,max=150"`
	ShortDescription string     `json:"short_description" validate:"required,min=2,max=200"`
	Url              string     `json:"url" validate:"required,min=2,max=50,urlpath"`
	TagsString       string     `json:"tags_string" validate:"required,min=2,max=150"`
	Status           string     `json:"status" validate:"required,oneof=draft public"`
	AccessMode       string     `json:"access_mode" validate:"required,oneof=1 2 3 4"`
	Price            *float64   `json:"price"`
	SubscriptionId   *uuid.UUID `json:"subscription_id"`
}

type PostUpdateTitleRequest struct {
	Title            string `json:"title" validate:"required,min=2,max=200"`
	ShortDescription string `json:"short_description"`
}

type PostUpdateUrlRequest struct {
	Url string `json:"url" validate:"required,min=2,max=200"`
}

type PostUpdateContentRequest struct {
	DataJson string `json:"data_json" validate:"required"`
	DataHtml string `json:"data_html"`
}

type PostUpdateTagsRequest struct {
	TagsString string `json:"tags_string"`
}

type PostUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft public"`
}

type PostUpdateAccessModeRequest struct {
	AccessMode     string     `json:"access_mode" validate:"required,oneof=1 2 3 4"`
	Price          *float64   `json:"price"`
	SubscriptionId *uuid.UUID `json:"subscription_id"`
}

type PostLikesInfoResponse struct {
	Likes     int  `json:"likes"`
	Dislikes  int  `json:"dislikes"`
	MyLike    bool `json:"my_like"`
	MyDislike bool `json:"my_dislike"`
}

type PostMyContentAccessResponse struct {
	HaveAccess   bool          `json:"have_access"`
	UserId       uuid.UUID     `json:"user_id"`
	PostId       uuid.UUID     `json:"post_id"`
	AccessMode   string        `json:"access_mode"`
	Price        float64       `json:"price"`
	Subscription *Subscription `json:"subscription"`
}

//type GrantPostPaidAccessServiceRequest struct {
//	PostId uuid.UUID `json:"post_id" validate:"required"`
//	UserId uuid.UUID `json:"user_id" validate:"required"`
//}
