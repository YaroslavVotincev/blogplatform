package users

import (
	"github.com/google/uuid"
	"time"
)

type AdminGetUserRequest struct {
	ID               uuid.UUID  `json:"id"`
	Login            string     `json:"login"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	Deleted          bool       `json:"deleted"`
	Enabled          bool       `json:"enabled"`
	EmailConfirmedAt *time.Time `json:"email_confirmed_at,omitempty"`
	EraseAt          *time.Time `json:"erase_at,omitempty"`
	BannedUntil      *time.Time `json:"banned_until,omitempty"`
	BannedReason     *string    `json:"banned_reason,omitempty"`
	Created          time.Time  `json:"created"`
	Updated          time.Time  `json:"updated"`
}

type AdminCreateUserRequest struct {
	Login    string `json:"login" validate:"required,min=2,max=30,login"`
	Email    string `json:"email" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=admin moderator user"`
	Password string `json:"password" validate:"required,password"`
	Enabled  bool   `json:"enabled"`
}

type AdminUpdateUserRequest struct {
	Login        string     `json:"login" validate:"required,min=2,max=30,login"`
	Email        string     `json:"email" validate:"required"`
	Role         string     `json:"role" validate:"required,oneof=admin moderator user"`
	Deleted      bool       `json:"deleted"`
	Enabled      bool       `json:"enabled"`
	BannedUntil  *time.Time `json:"banned_until"`
	BannedReason *string    `json:"banned_reason"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
