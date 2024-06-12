package users

import (
	"github.com/google/uuid"
)

type UserInfoResponse struct {
	ID           uuid.UUID `json:"id"`
	Role         string    `json:"role"`
	Login        string    `json:"login"`
	Email        string    `json:"email"`
	IsBanned     bool      `json:"is_banned"`
	BannedReason *string   `json:"banned_reason"`
}

type MyProfileWalletResponse struct {
	BalanceTon float64 `json:"balance_ton"`
	BalanceRub float64 `json:"balance_rub"`
	Address    string  `json:"address"`
}

type MyProfileFioRequest struct {
	FirstName  string `json:"first_name" validate:"max=30"`
	LastName   string `json:"last_name" validate:"max=30"`
	MiddleName string `json:"middle_name" validate:"max=30"`
}

type ChangeMyPasswordRequest struct {
	Old string `json:"old" validate:"required"`
	New string `json:"new" validate:"required,password"`
}

type AddRubToBalanceServiceRequest struct {
	Value float64 `json:"value" validate:"required"`
}
