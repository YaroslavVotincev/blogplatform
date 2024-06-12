package blogs

import (
	"github.com/google/uuid"
	"time"
)

type Blog struct {
	ID               uuid.UUID `json:"id"`
	AuthorId         uuid.UUID `json:"author_id"`
	Type             string    `json:"type"`
	Url              string    `json:"url"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"short_description"`
	Status           string    `json:"status"`
	AcceptDonations  bool      `json:"accept_donations"`
	Avatar           *string   `json:"avatar"`
	Cover            *string   `json:"cover"`
	Categories       []string  `json:"categories"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
}

type Category struct {
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Post struct {
	ID               uuid.UUID  `json:"id"`
	BlogId           uuid.UUID  `json:"blog_id"`
	Title            string     `json:"title"`
	Url              string     `json:"url"`
	ShortDescription string     `json:"short_description"`
	TagsString       string     `json:"tags_string"`
	Status           string     `json:"status"`
	Cover            *string    `json:"cover"`
	AccessMode       string     `json:"access_mode"`
	Price            *float64   `json:"price"`
	LikesCount       int        `json:"likes_count"`
	CommentsCount    int        `json:"comments_count"`
	SubscriptionId   *uuid.UUID `json:"subscription_id"`
	Created          time.Time  `json:"created"`
	Updated          time.Time  `json:"updated"`
}

type Content struct {
	ID       uuid.UUID `json:"id"`
	DataJson string    `json:"data_json"`
	DataHtml string    `json:"data_html"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type ContentFile struct {
	ID        uuid.UUID `json:"id"`
	ContentId uuid.UUID `json:"content_id"`
	Type      string    `json:"type"`
	Size      int       `json:"size"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

type Subscription struct {
	ID               uuid.UUID `json:"id"`
	BlogId           uuid.UUID `json:"blog_id"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"short_description"`
	IsFree           bool      `json:"is_free"`
	PriceRub         float64   `json:"price_rub"`
	Cover            *string   `json:"cover"`
	IsActive         bool      `json:"is_active"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
}

type UserSubscription struct {
	ID             uuid.UUID  `json:"id"`
	UserId         uuid.UUID  `json:"user_id"`
	SubscriptionId uuid.UUID  `json:"subscription_id"`
	BlogId         uuid.UUID  `json:"blog_id"`
	Status         string     `json:"status"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at"`
	Created        time.Time  `json:"created"`
	Updated        time.Time  `json:"updated"`
}

type UserFollow struct {
	ID      uuid.UUID `json:"id"`
	UserId  uuid.UUID `json:"user_id"`
	BlogId  uuid.UUID `json:"blog_id"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type PostLike struct {
	ID       uuid.UUID `json:"id"`
	PostId   uuid.UUID `json:"post_id"`
	UserId   uuid.UUID `json:"user_id"`
	Positive bool      `json:"positive"`
	Created  time.Time `json:"created"`
}

type PostPaidAccess struct {
	ID      uuid.UUID `json:"id"`
	PostId  uuid.UUID `json:"post_id"`
	UserId  uuid.UUID `json:"user_id"`
	Created time.Time `json:"created"`
}

type Goal struct {
	ID          uuid.UUID `json:"id"`
	BlogId      uuid.UUID `json:"blog_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Target      int       `json:"target"`
	Current     int       `json:"current"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

type PostView struct {
	ID          uuid.UUID  `json:"id"`
	PostId      uuid.UUID  `json:"post_id"`
	Fingerprint *string    `json:"fingerprint"`
	UserId      *uuid.UUID `json:"user_id"`
	Created     time.Time  `json:"created"`
}

type Donation struct {
	ID               uuid.UUID `json:"id"`
	UserId           uuid.UUID `json:"user_id"`
	BlogId           uuid.UUID `json:"blog_id"`
	Value            float64   `json:"value"`
	Currency         string    `json:"currency"`
	UserComment      string    `json:"user_comment"`
	Status           string    `json:"status"`
	PaymentConfirmed bool      `json:"payment_confirmed"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
}

type BlogIncome struct {
	ID               uuid.UUID `json:"id"`
	BlogId           uuid.UUID `json:"blog_id"`
	UserId           uuid.UUID `json:"user_id"`
	Value            float64   `json:"value"`
	Currency         string    `json:"currency"`
	ItemId           uuid.UUID `json:"item_id"`
	ItemType         string    `json:"item_type"`
	SentToUserWallet bool      `json:"sent_to_user_wallet"`
	Created          time.Time `json:"created"`
}
