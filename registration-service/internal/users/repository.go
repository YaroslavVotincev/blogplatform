package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
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
	where id = $1 and deleted is false`
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

func (r *Repository) ExistsById(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `select count(id) from users where id = $1`
	count := 0
	err := r.db.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	query := `select count(id) from users where lower(login) = lower($1)`
	count := 0
	err := r.db.QueryRow(ctx, query, login).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `select count(id) from users where email = $1`
	count := 0
	err := r.db.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *User, profile *Profile) error {
	if user == nil {
		return fmt.Errorf("tried to create a nil user")
	}
	if profile == nil {
		return fmt.Errorf("tried to create a user with nil profile")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `insert into users
	(id, login, email, hashed_password, role, deleted, enabled, email_confirmed_at,
	 erase_at, created, updated, banned_until, banned_reason)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.Exec(ctx, query,
		user.ID,
		user.Login,
		user.Email,
		user.HashedPassword,
		user.Role,
		user.Deleted,
		user.Enabled,
		user.EmailConfirmedAt,
		user.EraseAt,
		user.Created,
		user.Updated,
		user.BannedUntil,
		user.BannedReason,
	)
	if err != nil {
		return err
	}

	query = `insert into profiles
	(id, first_name, last_name, middle_name)
	values ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, query,
		user.ID,
		profile.FirstName,
		profile.LastName,
		profile.MiddleName,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) UserIdByConfirmCode(ctx context.Context, code string) (*uuid.UUID, error) {
	query := `select user_id from signup_confirm_codes where code = $1`
	var userId uuid.UUID
	err := r.db.QueryRow(ctx, query, code).Scan(&userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &userId, nil
}

func (r *Repository) Enable(ctx context.Context, userId uuid.UUID, confirmCode string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `update users set enabled = true, erase_at = null, email_confirmed_at = $2 where id = $1`
	_, err = tx.Exec(ctx, query, userId, time.Now().UTC())
	if err != nil {
		return err
	}

	query = `delete from signup_confirm_codes where code = $1`
	_, err = tx.Exec(ctx, query, confirmCode)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) EraseAllDue(ctx context.Context) error {
	query := `delete from users where erase_at < current_timestamp`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return err
	}
	query = `delete from signup_confirm_codes where expires_at < current_timestamp`
	_, err = r.db.Exec(ctx, query)
	return err
}

func (r *Repository) CreateConfirmationCode(ctx context.Context, code string, userId uuid.UUID, expiresAt time.Time) error {
	query := `insert into signup_confirm_codes(code, user_id, expires_at, created_at) values ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, code, userId, expiresAt, time.Now().UTC())
	return err
}

func (r *Repository) EraseConfirmCode(ctx context.Context, code string) error {
	query := `delete from signup_confirm_codes where code = $1`
	_, err := r.db.Exec(ctx, query, code)
	return err
}
