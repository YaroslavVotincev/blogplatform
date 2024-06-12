package robokassa

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

func (r *Repository) InvoicesByParams(ctx context.Context,
	invIds []int,
	itemIds, itemTypes, userIds, statuses []string,
	startTime, endTime *time.Time, limit, offset *int,
) ([]Invoice, error) {

	query := `SELECT id, out_sum, item_id, item_type, user_id, expires_at, status, payment_link, created, updated 
              FROM robokassa_invoices`

	var args []interface{}
	var conditions []string

	if len(invIds) > 0 {
		ids := make([]string, len(invIds))
		for i, id := range invIds {
			ids[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("id IN (%s)", strings.Join(ids, ",")))
	}

	if len(itemIds) > 0 {
		ids := make([]string, len(itemIds))
		for i, id := range itemIds {
			ids[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("item_id IN (%s)", strings.Join(ids, ",")))
	}

	if len(itemTypes) > 0 {
		types := make([]string, len(itemTypes))
		for i, t := range itemTypes {
			types[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, t)
		}
		conditions = append(conditions, fmt.Sprintf("item_type IN (%s)", strings.Join(types, ",")))
	}

	if len(userIds) > 0 {
		uids := make([]string, len(userIds))
		for i, id := range userIds {
			uids[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("user_id IN (%s)", strings.Join(uids, ",")))
	}

	if len(statuses) > 0 {
		sts := make([]string, len(statuses))
		for i, s := range statuses {
			sts[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, s)
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(sts, ",")))
	}

	if startTime != nil {
		conditions = append(conditions, fmt.Sprintf("created >= $%d", len(args)+1))
		args = append(args, *startTime)
	}

	if endTime != nil {
		conditions = append(conditions, fmt.Sprintf("created <= $%d", len(args)+1))
		args = append(args, *endTime)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created DESC"

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, *limit)
	}

	if offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
		args = append(args, *offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Invoice, 0)
	var item Invoice
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.OutSum,
			&item.ItemId,
			&item.ItemType,
			&item.UserId,
			&item.ExpiresAt,
			&item.Status,
			&item.PaymentLink,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}

	return resultArray, nil
}

func (r *Repository) InvoiceById(ctx context.Context, id int) (*Invoice, error) {
	query := `select id, out_sum, item_id, item_type, user_id, expires_at, status, payment_link, created, updated 
		from robokassa_invoices	
		where id = $1`

	var item Invoice
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.OutSum,
		&item.ItemId,
		&item.ItemType,
		&item.UserId,
		&item.ExpiresAt,
		&item.Status,
		&item.PaymentLink,
		&item.Created,
		&item.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) InvoicesByUserIdAndItemId(ctx context.Context, userId, subId uuid.UUID) ([]Invoice, error) {
	query := `select id, out_sum, item_id, item_type, user_id, expires_at, status, payment_link, created, updated 
		from robokassa_invoices
		where user_id = $1 and item_id = $2
		order by created desc 
		`

	rows, err := r.db.Query(ctx, query, userId, subId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Invoice, 0)
	var item Invoice
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.OutSum,
			&item.ItemId,
			&item.ItemType,
			&item.UserId,
			&item.ExpiresAt,
			&item.Status,
			&item.PaymentLink,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) CreateInvoice(ctx context.Context, item *Invoice) error {
	query := `insert into robokassa_invoices
    (out_sum, item_id, item_type, user_id, expires_at, status, payment_link, created, updated) 
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	var id int
	err := r.db.QueryRow(ctx, query,
		item.OutSum,
		item.ItemId,
		item.ItemType,
		item.UserId,
		item.ExpiresAt,
		item.Status,
		item.PaymentLink,
		item.Created,
		item.Updated,
	).Scan(&id)
	if err != nil {
		return err
	}
	item.ID = id
	return nil
}

func (r *Repository) UpdateInvoice(ctx context.Context, item *Invoice) error {
	query := `update robokassa_invoices set
            out_sum = $2, 
            item_id = $3,
            item_type = $4,
            user_id = $5, 
            expires_at = $6, 
            status = $7,
            payment_link = $8,
            created = $9, 
            updated = $10
			where id = $1
	`
	_, err := r.db.Exec(ctx, query,
		item.ID,
		item.OutSum,
		item.ItemId,
		item.ItemType,
		item.UserId,
		item.ExpiresAt,
		item.Status,
		item.PaymentLink,
		item.Created,
		item.Updated,
	)
	return err
}
