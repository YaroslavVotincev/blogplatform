package users

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `select 
			id, login, email, role, deleted, enabled, banned_until, banned_reason, hashed_password
	from users
	where id = $1 and deleted is false and enabled is true`
	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Role,
		&user.Deleted,
		&user.Enabled,
		&user.BannedUntil,
		&user.BannedReason,
		&user.HashedPassword,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ByValue(ctx context.Context, value string) (*User, error) {
	query := `select 
			id, login, email, role, deleted, enabled, banned_until, banned_reason, hashed_password
	from users
	where (lower(login) = lower($1) or email = $1) and deleted is false and enabled is true`
	var user User
	err := r.db.QueryRow(ctx, query, value).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Role,
		&user.Deleted,
		&user.Enabled,
		&user.BannedUntil,
		&user.BannedReason,
		&user.HashedPassword,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
