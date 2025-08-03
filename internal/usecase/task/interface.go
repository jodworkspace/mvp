package taskuc

import (
	"context"
	"gitlab.com/jodworkspace/mvp/internal/domain"
)

type TaskUseCase interface {
	Count(ctx context.Context, filter *domain.Filter) (int64, error)
	List(ctx context.Context, filter *domain.Filter) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) error
	Get(ctx context.Context, id uint64) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id uint64) error
}

type TaskRepository interface {
	Count(ctx context.Context, filter *domain.Filter) (int64, error)
	List(ctx context.Context, filter *domain.Filter) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Get(ctx context.Context, id uint64) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) (*domain.Task, error)
	Delete(ctx context.Context, id uint64) error
}
