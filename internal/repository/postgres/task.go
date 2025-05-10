package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type TaskRepository struct {
	db.Postgres
}

func NewTaskRepository(conn db.Postgres) *TaskRepository {
	return &TaskRepository{conn}
}

func (r *TaskRepository) Count(ctx context.Context) (int64, error) {
	query, args, err := r.QueryBuilder().
		Select("count(*)").
		From(domain.TableTask).
		ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.Pool().QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
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

func (r *TaskRepository) Get(ctx context.Context, id uint64) (*domain.Task, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.TaskAllColumns...).
		From(domain.TableTask).
		Where(squirrel.Eq{domain.ColID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var task domain.Task
	err = r.Pool().QueryRow(ctx, query, args...).Scan(
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

	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	return task, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uint64) error {
	query, args, err := r.QueryBuilder().
		Delete(domain.TableTask).
		Where(squirrel.Eq{domain.ColID: id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.Pool().Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
