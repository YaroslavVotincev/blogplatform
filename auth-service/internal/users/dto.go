package users

import (
	"github.com/google/uuid"
)

type LoginRequest struct {
	Login      string `json:"login" validate:"required,min=2"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserInfoResponse struct {
	ID           uuid.UUID `json:"id"`
	Role         string    `json:"role"`
	Login        string    `json:"login"`
	Email        string    `json:"email"`
	IsBanned     bool      `json:"is_banned"`
	BannedReason *string   `json:"banned_reason"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
