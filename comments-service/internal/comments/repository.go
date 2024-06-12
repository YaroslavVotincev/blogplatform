package comments

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

func (r *Repository) ByParentId(ctx context.Context, parentId uuid.UUID) ([]Comment, error) {
	query := `SELECT id, parent_id, author_id, content, created, updated FROM comments 
            WHERE parent_id = $1 order by created desc`
	rows, err := r.db.Query(ctx, query, parentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultArray := make([]Comment, 0)
	var item Comment
	item.Children = make([]Comment, 0)
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.ParentId,
			&item.AuthorId,
			&item.Content,
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

func (r *Repository) ByParentIdMap(ctx context.Context, parentId uuid.UUID) (map[uuid.UUID]Comment, error) {
	query := `SELECT id, parent_id, author_id, content, created, updated FROM comments 
            WHERE parent_id = $1 order by created desc`
	rows, err := r.db.Query(ctx, query, parentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultMap := make(map[uuid.UUID]Comment)
	var item Comment
	item.Children = make([]Comment, 0)
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.ParentId,
			&item.AuthorId,
			&item.Content,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultMap[item.ID] = item
	}
	return resultMap, nil
}

func (r *Repository) ById(ctx context.Context, id uuid.UUID) (*Comment, error) {
	query := `SELECT id, parent_id, author_id, content, created, updated FROM comments WHERE id = $1`
	var item Comment
	item.Children = make([]Comment, 0)
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.ParentId,
		&item.AuthorId,
		&item.Content,
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

func (r *Repository) ByParentIdList(ctx context.Context, parentIds []uuid.UUID) ([]Comment, error) {

	if len(parentIds) == 0 {
		return nil, nil
	}

	query := `SELECT id, parent_id, author_id, content, created, updated 
			FROM comments WHERE parent_id = ANY($1) 
			order by created desc`
	rows, err := r.db.Query(ctx, query, parentIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultArray := make([]Comment, 0)
	var item Comment
	item.Children = make([]Comment, 0)
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.ParentId,
			&item.AuthorId,
			&item.Content,
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

func (r *Repository) ByParentId2Levels(ctx context.Context, parentId uuid.UUID) ([]Comment, error) {

	query := `
			SELECT id, parent_id, author_id, content, created, updated
			FROM comments
			WHERE parent_id = $1 OR parent_id IN (
    			SELECT id FROM comments WHERE parent_id = $1
			)
			ORDER BY created`

	rows, err := r.db.Query(ctx, query, parentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topComments := make([]Comment, 0)
	childComments := make([]Comment, 0)
	var item Comment
	item.Children = make([]Comment, 0)
	for rows.Next() {

		err = rows.Scan(
			&item.ID,
			&item.ParentId,
			&item.AuthorId,
			&item.Content,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		if item.ParentId == parentId {
			topComments = append(topComments, item)
		} else {
			childComments = append(childComments, item)
		}
	}

	for i := range topComments {
		for j := range childComments {
			if topComments[i].ID == childComments[j].ParentId {
				topComments[i].Children = append(topComments[i].Children, childComments[j])
			}
		}
	}

	return topComments, nil
}

func (r *Repository) ByParentId2LevelsCount(ctx context.Context, parentId uuid.UUID) (int, error) {

	query := `
			SELECT count(id)
			FROM comments
			WHERE parent_id = $1 OR parent_id IN (
    			SELECT id FROM comments WHERE parent_id = $1
			)`
	count := 0
	err := r.db.QueryRow(ctx, query, parentId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) Create(ctx context.Context, item *Comment) error {
	query := `INSERT INTO comments (id, parent_id, author_id, content, created, updated) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, item.ID, item.ParentId, item.AuthorId, item.Content, item.Created, item.Updated)
	return err
}
