package configservice

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AllServices(ctx context.Context) ([]ServiceModel, error) {
	query := `select service, created, updated from settings_services order by created, service`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]ServiceModel, 0)
	var service ServiceModel
	for rows.Next() {
		err = rows.Scan(&service.Service, &service.Created, &service.Updated)
		if err != nil {
			return nil, err
		}
		results = append(results, service)
	}
	return results, err
}

func (r *Repository) ServiceByName(ctx context.Context, name string) (*ServiceModel, error) {
	query := `select service, created, updated from settings_services WHERE service = $1`
	var service ServiceModel
	err := r.db.QueryRow(ctx, query, name).Scan(&service.Service, &service.Created, &service.Updated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &service, nil
}

func (r *Repository) CreateService(ctx context.Context, service *ServiceModel) error {
	query := `insert into settings_services (service, created, updated) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, service.Service, service.Created, service.Updated)
	return err
}

func (r *Repository) UpdateService(ctx context.Context, oldName, newName string) error {
	query := `update settings_services set service = $1, updated = $2 where service = $3`
	_, err := r.db.Exec(ctx, query, newName, time.Now().UTC(), oldName)
	return err
}

func (r *Repository) DeleteService(ctx context.Context, name string) error {
	query := `delete from settings_services where service = $1`
	_, err := r.db.Exec(ctx, query, name)
	return err
}

func (r *Repository) SettingsByService(ctx context.Context, name string) ([]Setting, error) {
	query := `select key, value, created, updated from settings_items where service = $1`
	rows, err := r.db.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var item = Setting{Service: name}
	results := make([]Setting, 0)
	for rows.Next() {
		err = rows.Scan(&item.Key, &item.Value, &item.Created, &item.Updated)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func (r *Repository) SetSettingsToService(ctx context.Context, service string, items []Setting) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	query := `delete from settings_items where service = $1`
	_, err = r.db.Exec(ctx, query, service)
	if err != nil {
		return err
	}
	for _, item := range items {
		query = `insert into settings_items (service, key, value, created, updated) values ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(ctx, query, service, item.Key, item.Value, item.Created, item.Updated)
		if err != nil {
			return err
		}
	}
	query = `update settings_services set updated = $2 WHERE service = $1`
	_, err = tx.Exec(ctx, query, service, time.Now().UTC())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
