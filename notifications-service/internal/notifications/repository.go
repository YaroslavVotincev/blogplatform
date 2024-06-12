package notifications

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AllByUser(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	query := `select id, event_code, user_id, seen, data, created, updated
	from notifications
	where user_id = $1
	order by created desc`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Notification, 0)
	var item Notification
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.EventCode,
			&item.UserID,
			&item.Seen,
			&item.DataBytes,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(item.DataBytes, &item.Data)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) Create(ctx context.Context, notification Notification) error {
	query := `insert into notifications(id, event_code, user_id, seen, data, created, updated) 
			values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query,
		notification.ID,
		notification.EventCode,
		notification.UserID,
		notification.Seen,
		notification.DataBytes,
		notification.Created,
		notification.Updated,
	)
	return err
}

func (r *Repository) CountUnseenForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `select count(id) from notifications where user_id = $1 and seen = false`
	row := r.db.QueryRow(ctx, query, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) UpdateAllToSeenForUser(ctx context.Context, userID uuid.UUID) error {
	query := `update notifications set seen = true, updated = current_timestamp where user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
