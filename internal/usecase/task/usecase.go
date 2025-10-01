package task

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"go.uber.org/zap"
)

type UseCase struct {
	taskRepo Repository
	logger   *logger.ZapLogger
}

func NewUseCase(taskRepo Repository, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		taskRepo: taskRepo,
		logger:   zl,
	}
}
func (u *UseCase) Count(ctx context.Context, ownerID string) (int64, error) {
	count, err := u.taskRepo.Count(ctx, ownerID)
	if err != nil {
		u.logger.Error("taskUseCase - taskRepo.Count", zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (u *UseCase) List(ctx context.Context, page, pageSize uint64, ownerID string) ([]*domain.Task, error) {
	tasks, err := u.taskRepo.List(ctx, page, pageSize, ownerID)
	if err != nil {
		u.logger.Error(
			"taskUseCase - taskRepo.List",
			zap.Uint64("page", page),
			zap.Uint64("page_size", pageSize),
			zap.String("ownerID", ownerID),
			zap.Error(err),
		)
		return nil, err
	}

	return tasks, nil
}

func (u *UseCase) Create(ctx context.Context, task *domain.Task) error {
	task.ID = uuid.NewString()
	task.IsCompleted = false
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := u.taskRepo.Create(ctx, task)
	if err != nil {
		u.logger.Error(
			"taskUseCase- taskRepo.Create",
			zap.String("owner_id", task.OwnerID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (u *UseCase) Get(ctx context.Context, id string) (*domain.Task, error) {
	return nil, nil
}

func (u *UseCase) Update(ctx context.Context, task *domain.Task) error {
	return nil
}

func (u *UseCase) Delete(ctx context.Context, id string) error {
	err := u.taskRepo.Delete(ctx, id)
	if err != nil {
		u.logger.Error(
			"taskUseCase - taskRepo.Delete",
			zap.String("task_id", id),
			zap.Error(err),
		)
		return err
	}

	return nil
}
