package taskuc

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"go.uber.org/zap"
	"time"
)

type taskUsecase struct {
	taskRepo TaskRepository
	logger   *logger.ZapLogger
}

func NewTaskUsecase(taskRepo TaskRepository, zl *logger.ZapLogger) TaskUseCase {
	return &taskUsecase{
		taskRepo: taskRepo,
		logger:   zl,
	}
}
func (u *taskUsecase) Count(ctx context.Context) (int64, error) {
	count, err := u.taskRepo.Count(ctx)
	if err != nil {
		u.logger.Error("taskUsecase - taskRepo.Count", zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (u *taskUsecase) List(ctx context.Context, page uint64, pageSize uint64) ([]*domain.Task, error) {
	tasks, err := u.taskRepo.List(ctx, page, pageSize)
	if err != nil {
		u.logger.Error("taskUsecase- taskRepo.List", zap.Error(err))
		return nil, err
	}
	return tasks, nil
}

func (u *taskUsecase) Create(ctx context.Context, task *domain.Task) error {
	task.ID = uuid.NewString()
	task.IsCompleted = false
	task.OwnerUserID = ""
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := u.taskRepo.Create(ctx, task)
	if err != nil {
		u.logger.Error("taskUsecase- taskRepo.Create", zap.Error(err))
		return err
	}
	return nil
}

func (u *taskUsecase) Get(ctx context.Context, id uint64) (*domain.Task, error) {
	return nil, nil
}

func (u *taskUsecase) Update(ctx context.Context, task *domain.Task) error {
	return nil
}

func (u *taskUsecase) Delete(ctx context.Context, id uint64) error {
	return nil
}
