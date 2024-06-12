package users

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID               uuid.UUID
	Login            string
	Email            string
	HashedPassword   string
	Role             string
	Deleted          bool
	Enabled          bool
	EmailConfirmedAt *time.Time
	EraseAt          *time.Time
	Created          time.Time
	Updated          time.Time
	BannedUntil      *time.Time
	BannedReason     *string
}

type Profile struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	MiddleName string
	Avatar     *string
}

type Wallet struct {
	ID         uuid.UUID `json:"id"`
	PublicKey  string    `json:"publicKey"`
	SecretKey  string    `json:"secretKey"`
	Mnemonic   []string  `json:"mnemonic"`
	Address    string    `json:"address"`
	BalanceRub float64   `json:"balanceRub"`
	Created    time.Time `json:"created"`
}
