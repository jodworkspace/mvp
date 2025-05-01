package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type UserRepository struct {
	db.PostgresConn
}

func NewUserRepository(conn db.PostgresConn) *UserRepository {
	return &UserRepository{conn}
}

func (r *UserRepository) Exists(ctx context.Context, col string, val any) (bool, error) {
	return exists(r.PostgresConn, ctx, domain.TableUser, col, val)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.UserPublicCol...).
		From(domain.TableUser).
		Where(squirrel.Eq{domain.ColUserEmail: email}).
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
