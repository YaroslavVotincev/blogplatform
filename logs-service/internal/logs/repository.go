package logs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item *HistoryLog) error {
	query := `insert into history_logs(id, level, service, user_id, data, created)
	values($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query,
		item.ID,
		item.Level,
		item.Service,
		item.UserID,
		item.DataRaw,
		item.Created,
	)
	return err
}
