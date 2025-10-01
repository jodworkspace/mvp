package task

import (
	"context"

	"gitlab.com/jodworkspace/mvp/internal/domain"
)

type Repository interface {
	Count(ctx context.Context, ownerID string) (int64, error)
	List(ctx context.Context, page, pageSize uint64, ownerID string) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Get(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Delete(ctx context.Context, id string) error
}
