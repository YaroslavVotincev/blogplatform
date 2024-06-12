package pgutils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DBPool(ctx context.Context, dbUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, fmt.Errorf("database connection fail cause %v", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("fail to ping database cause %v", err)
	}
	return pool, nil
}
