package taskuc

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/gookie/mvp/internal/domain"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	"gitlab.com/gookie/mvp/pkg/logger"
	"time"
)

type taskUsecase struct {
	repo *postgresrepo.TaskRepository
	zl   *logger.ZapLogger
}

func NewTaskUsecase(repo *postgresrepo.TaskRepository, zl *logger.ZapLogger) TaskUsecase {
	return &taskUsecase{repo: repo, zl: zl}
}

func (u *taskUsecase) GetAll(ctx context.Context, page int, pageSize int) ([]*domain.Task, error) {
	return nil, nil
}

func (u *taskUsecase) Create(ctx context.Context, task *domain.Task) error {
	task.ID = uuid.NewString()
	task.IsCompleted = false
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := u.repo.Create(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (u *taskUsecase) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return nil, nil
}

func (u *taskUsecase) Update(ctx context.Context, task *domain.Task) error {
	return nil
}

func (u *taskUsecase) Delete(ctx context.Context, id string) error {
	return nil
}
