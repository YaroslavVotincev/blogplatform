package historylogs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	query := "select count(id) from history_logs"
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) All(ctx context.Context, levels, services []string, userIDs []*uuid.UUID, startTime, endTime *time.Time, limit, skip *int) ([]HistoryLog, error) {
	query := `SELECT id, level, service, user_id, data, created FROM history_logs WHERE 1=1`
	var args []interface{}
	var conditions []string

	if len(levels) > 0 {
		var placeholders []string
		for i, level := range levels {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			args = append(args, level)
		}
		conditions = append(conditions, fmt.Sprintf("level IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(services) > 0 {
		var placeholders []string
		for _, service := range services {
			idx := len(args) + 1
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, service)
		}
		conditions = append(conditions, fmt.Sprintf("service IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(userIDs) > 0 {
		var placeholders []string
		includeNull := false

		for _, userID := range userIDs {
			idx := len(args) + 1
			if userID != nil {
				placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
				args = append(args, *userID)
			} else {
				includeNull = true
			}
		}

		condition := fmt.Sprintf("user_id IN (%s)", strings.Join(placeholders, ", "))
		if includeNull {
			condition = fmt.Sprintf("(%s OR user_id IS NULL)", condition)
		}

		conditions = append(conditions, condition)
	}

	if startTime != nil && endTime != nil {
		idxStart := len(args) + 1
		idxEnd := len(args) + 2
		conditions = append(conditions, fmt.Sprintf("created BETWEEN $%d AND $%d", idxStart, idxEnd))
		args = append(args, *startTime, *endTime)
	} else if startTime != nil {
		idx := len(args) + 1
		conditions = append(conditions, fmt.Sprintf("created >= $%d", idx))
		args = append(args, *startTime)
	} else if endTime != nil {
		idx := len(args) + 1
		conditions = append(conditions, fmt.Sprintf("created <= $%d", idx))
		args = append(args, *endTime)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created DESC"

	if limit != nil {
		idx := len(args) + 1
		query += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, *limit)
	}

	if skip != nil {
		idx := len(args) + 1
		query += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, *skip)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]HistoryLog, 0)
	var item HistoryLog
	for rows.Next() {
		var userID uuid.UUID
		var userIDPtr *uuid.UUID = nil
		err = rows.Scan(
			&item.ID,
			&item.Level,
			&item.Service,
			&userID,
			&item.DataRaw,
			&item.Created,
		)
		if err != nil {
			return nil, err
		}
		if userID != uuid.Nil {
			userIDPtr = &userID
		}
		item.UserID = userIDPtr
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) ByID(ctx context.Context, id uuid.UUID) (*HistoryLog, error) {
	query := `select id, level, service, user_id, data, created
	from history_logs
	where id = $1`
	var item HistoryLog
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Level,
		&item.Service,
		&item.UserID,
		&item.DataRaw,
		&item.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
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

func (r *Repository) Levels(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT level FROM history_logs`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]string, 0)
	var item string
	for rows.Next() {
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}

	return resultArray, nil
}

func (r *Repository) Services(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service FROM history_logs`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]string, 0)
	var item string
	for rows.Next() {
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}

	return resultArray, nil
}
