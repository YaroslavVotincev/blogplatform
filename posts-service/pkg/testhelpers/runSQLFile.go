package testhelpers

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Execute(ctx context.Context, db *pgxpool.Pool, path string) error {
	c, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sql := string(c)
	_, err = db.Exec(ctx, sql)
	if err != nil {
		return err
	}
	return nil
}
