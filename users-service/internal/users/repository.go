package users

import (
	"context"
	"errors"
	"fmt"
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

func (r *Repository) Count(ctx context.Context) (int64, error) {
	query := "select count(id) from users"
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) All(ctx context.Context) ([]User, error) {
	query := `select id, login, email, role, deleted, enabled,
			email_confirmed_at, erase_at, banned_until, banned_reason,
			created, updated
			from users
			order by login`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]User, 0)
	var user User
	for rows.Next() {
		err = rows.Scan(
			&user.ID,
			&user.Login,
			&user.Email,
			&user.Role,
			&user.Deleted,
			&user.Enabled,
			&user.EmailConfirmedAt,
			&user.EraseAt,
			&user.BannedUntil,
			&user.BannedReason,
			&user.Created,
			&user.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, user)
	}
	return resultArray, nil
}

func (r *Repository) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `select 
			id, login, email, role, deleted, enabled,
			email_confirmed_at, erase_at, banned_until, banned_reason,
			created, updated
	from users
	where id = $1`
	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Role,
		&user.Deleted,
		&user.Enabled,
		&user.EmailConfirmedAt,
		&user.EraseAt,
		&user.BannedUntil,
		&user.BannedReason,
		&user.Created,
		&user.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ByIDList(ctx context.Context, idList []string) ([]User, error) {
	query := `select 
			id, login, email, role, deleted, enabled,
			email_confirmed_at, erase_at, banned_until, banned_reason,
			created, updated
	from users
	where id = Any($1)`
	rows, err := r.db.Query(ctx, query, idList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]User, 0)
	var user User
	for rows.Next() {
		err = rows.Scan(
			&user.ID,
			&user.Login,
			&user.Email,
			&user.Role,
			&user.Deleted,
			&user.Enabled,
			&user.EmailConfirmedAt,
			&user.EraseAt,
			&user.BannedUntil,
			&user.BannedReason,
			&user.Created,
			&user.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, user)
	}
	return resultArray, nil
}

func (r *Repository) ByLogin(ctx context.Context, login string) (*User, error) {
	query := `select 
			id, login, email, role, deleted, enabled,
			email_confirmed_at, erase_at, banned_until, banned_reason,
			created, updated
	from users
	where lower(login) = lower($1)`
	var user User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Role,
		&user.Deleted,
		&user.Enabled,
		&user.EmailConfirmedAt,
		&user.EraseAt,
		&user.BannedUntil,
		&user.BannedReason,
		&user.Created,
		&user.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ByEmail(ctx context.Context, email string) (*User, error) {
	query := `select 
			id, login, email, role, deleted, enabled,
			email_confirmed_at, erase_at, banned_until, banned_reason,
			created, updated
	from users
	where email = $1`
	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Role,
		&user.Deleted,
		&user.Enabled,
		&user.EmailConfirmedAt,
		&user.EraseAt,
		&user.BannedUntil,
		&user.BannedReason,
		&user.Created,
		&user.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *User, profile *Profile) error {
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

func (r *Repository) Update(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("tried to update a nil user")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `update users set
		login = $2,
		email = $3,
		role = $4,
		deleted = $5,
		enabled = $6,
		updated = $7,
		banned_until = $8,
		banned_reason = $9
	where id = $1`
	_, err = tx.Exec(ctx, query,
		user.ID,
		user.Login,
		user.Email,
		user.Role,
		user.Deleted,
		user.Enabled,
		user.Updated,
		user.BannedUntil,
		user.BannedReason,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) EraseById(ctx context.Context, id uuid.UUID) error {
	query := `delete from users where id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *Repository) WalletByUserId(ctx context.Context, userId uuid.UUID) (*Wallet, error) {
	query := `select id, address, publicKey, secretKey, mnemonic, balance_rub, created from wallets where id = $1`
	var wallet Wallet
	err := r.db.QueryRow(ctx, query, userId).Scan(
		&wallet.ID,
		&wallet.Address,
		&wallet.PublicKey,
		&wallet.SecretKey,
		&wallet.Mnemonic,
		&wallet.BalanceRub,
		&wallet.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *Repository) CreateWallet(ctx context.Context, wallet *Wallet) error {
	query := `insert into wallets (id, address, publicKey, secretKey, mnemonic, balance_rub, created) 
			values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query,
		wallet.ID,
		wallet.Address,
		wallet.PublicKey,
		wallet.SecretKey,
		wallet.Mnemonic,
		wallet.BalanceRub,
		wallet.Created,
	)
	return err
}

func (r *Repository) UpdateWalletRubBalance(ctx context.Context, wallet *Wallet) error {
	query := `update wallets set
		balance_rub = $2
	where id = $1`
	_, err := r.db.Exec(ctx, query,
		wallet.ID,
		wallet.BalanceRub,
	)
	return err
}

func (r *Repository) ProfileById(ctx context.Context, userId uuid.UUID) (*Profile, error) {
	query := `select id, first_name, last_name, middle_name, avatar from profiles where id = $1`
	var profile Profile
	err := r.db.QueryRow(ctx, query, userId).Scan(
		&profile.ID,
		&profile.FirstName,
		&profile.LastName,
		&profile.MiddleName,
		&profile.Avatar,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, profile *Profile) error {
	query := `update profiles set
		first_name = $2,
		last_name = $3,
		middle_name = $4,
		avatar = $5
	where id = $1`
	_, err := r.db.Exec(ctx, query,
		profile.ID,
		profile.FirstName,
		profile.LastName,
		profile.MiddleName,
		profile.Avatar,
	)
	return err
}

func (r *Repository) ByIdHashedPassword(ctx context.Context, id uuid.UUID) (*string, error) {
	query := `select 
			hashed_password
	from users
	where id = $1`
	var hash string
	err := r.db.QueryRow(ctx, query, id).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &hash, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, password string) error {
	query := `update users set
		hashed_password = $2
	where id = $1`
	_, err := r.db.Exec(ctx, query, id, password)
	return err
}
