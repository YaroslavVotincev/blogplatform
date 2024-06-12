package comments

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	service := Service{repository: repository}
	return &service
}

func (s *Service) ByParentId(ctx context.Context, parentId uuid.UUID) ([]Comment, error) {
	return s.repository.ByParentId2Levels(ctx, parentId)
}

func (s *Service) ById(ctx context.Context, id uuid.UUID) (*Comment, error) {
	return s.repository.ById(ctx, id)
}

func (s *Service) CreateFromRequest(ctx context.Context, parentId, authorId uuid.UUID, req *CommentCreateRequest) (*Comment, error) {
	timeNow := time.Now().UTC()
	comment := Comment{
		ID:       uuid.New(),
		ParentId: parentId,
		AuthorId: authorId,
		Content:  req.Content,
		Created:  timeNow,
		Updated:  timeNow,
	}
	return &comment, s.repository.Create(ctx, &comment)
}
