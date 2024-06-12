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
