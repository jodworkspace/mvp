package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/tokpok/mvp/internal/domain"
	"gitlab.com/tokpok/mvp/pkg/db"
)

type UserRepository struct {
	*db.PostgresConnection
}

func NewUserRepository(conn *db.PostgresConnection) *UserRepository {
	return &UserRepository{conn}
}

func (r *UserRepository) Exists(ctx context.Context, colName string, val any) (bool, error) {
	return exists(r.PostgresConnection, ctx, domain.TableUser, colName, val)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := r.SQLBuilder.
		Select(domain.UserPublicCol...).
		From(domain.TableUser).
		Where(squirrel.Eq{domain.ColEmail: email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = r.Pool.QueryRow(ctx, query, args...).Scan(
		&user.UserID,
		&user.Username,
		&user.FullName,
		&user.Email,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
