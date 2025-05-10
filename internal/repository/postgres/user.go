package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type UserRepository struct {
	db.Postgres
}

func NewUserRepository(conn db.Postgres) *UserRepository {
	return &UserRepository{conn}
}

func (r *UserRepository) Exists(ctx context.Context, col string, val any) (bool, error) {
	return exists(r.Postgres, ctx, domain.TableUser, col, val)
}

func (r *UserRepository) Insert(ctx context.Context, user *domain.User) error {
	query, args, err := r.QueryBuilder().
		Insert(domain.TableUser).
		Columns(domain.UserPublicCols...).
		Values(
			user.ID,
			user.DisplayName,
			user.Email,
			user.EmailVerified,
			user.AvatarURL,
			user.PreferredLanguage,
			user.Active,
			user.CreatedAt,
			user.UpdatedAt,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return err
	}

	return r.Pool().QueryRow(ctx, query, args...).Scan(&user.ID)
}

func (r *UserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.UserPublicCols...).
		From(domain.TableUser).
		Where(squirrel.Eq{domain.ColID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = r.Pool().QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.DisplayName,
		&user.Email,
		&user.EmailVerified,
		&user.AvatarURL,
		&user.PreferredLanguage,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
