package postgresrepo

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
)

type TaskRepository struct {
	client postgres.Client
}

func NewTaskRepository(pgc postgres.Client) *TaskRepository {
	return &TaskRepository{
		client: pgc,
	}
}

func (r *TaskRepository) Count(ctx context.Context, filter *domain.Filter) (int64, error) {
	builder := r.client.QueryBuilder().
		Select("count(*)").
		From(domain.TableTask)

	for key, value := range filter.Conditions {
		builder = builder.Where(squirrel.Eq{key: value})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.client.Pool().QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *TaskRepository) List(ctx context.Context, filter *domain.Filter) ([]*domain.Task, error) {
	builder := r.client.QueryBuilder().
		Select(domain.TaskAllColumns...).
		From(domain.TableTask).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		OrderBy(fmt.Sprintf("%s DESC", domain.ColCreatedAt))

	for key, value := range filter.Conditions {
		builder = builder.Where(squirrel.Eq{key: value})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.client.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var task domain.Task
		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Details,
			&task.Priority,
			&task.IsCompleted,
			&task.StartDate,
			&task.DueDate,
			&task.OwnerID,
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
	query, args, err := r.client.QueryBuilder().
		Insert(domain.TableTask).
		Columns(domain.TaskAllColumns...).
		Values(
			task.ID,
			task.Title,
			task.Details,
			task.Priority,
			task.IsCompleted,
			task.StartDate,
			task.DueDate,
			task.OwnerID,
			task.CreatedAt,
			task.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.client.Pool().Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) Get(ctx context.Context, id string) (*domain.Task, error) {
	query, args, err := r.client.QueryBuilder().
		Select(domain.TaskAllColumns...).
		From(domain.TableTask).
		Where(squirrel.Eq{domain.ColID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var task domain.Task
	err = r.client.Pool().QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&task.Title,
		&task.Details,
		&task.Priority,
		&task.IsCompleted,
		&task.StartDate,
		&task.DueDate,
		&task.OwnerID,
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

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query, args, err := r.client.QueryBuilder().
		Delete(domain.TableTask).
		Where(squirrel.Eq{domain.ColID: id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.client.Pool().Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
