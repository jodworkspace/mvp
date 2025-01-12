package taskuc

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
)

type TaskUsecase interface {
	GetAll(ctx context.Context, page int, pageSize int) ([]*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id string) error
}
