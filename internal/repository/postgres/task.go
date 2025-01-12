package postgresrepo

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type TaskRepository struct {
	db.PostgresConn
}

func NewTaskRepository(conn db.PostgresConn) *TaskRepository {
	return &TaskRepository{conn}
}

func (r *TaskRepository) List(ctx context.Context, page, pageSize uint64) ([]*domain.Task, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.TaskAllColumns...).
		From(domain.TableTask).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Details,
			&task.PriorityLevel,
			&task.IsCompleted,
			&task.StartDate,
			&task.EstimatedDuration,
			&task.DueDate,
			&task.OwnerUserID,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query, args, err := r.QueryBuilder().
		Insert(domain.TableTask).
		Columns(domain.TaskAllColumns...).
		Values(
			task.ID,
			task.Title,
			task.Details,
			task.PriorityLevel,
			task.IsCompleted,
			task.StartDate,
			task.EstimatedDuration,
			task.DueDate,
			task.OwnerUserID,
			task.CreatedAt,
			task.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.Pool().Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uint64) error {
	return nil
}
