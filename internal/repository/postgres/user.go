package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
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
	return exists(r.Postgres, ctx, domain.TableUsers, col, val)
}

func (r *UserRepository) Insert(ctx context.Context, user *domain.User, tx ...pgx.Tx) error {
	query, args, err := r.QueryBuilder().
		Insert(domain.TableUsers).
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

	if len(tx) > 0 {
		return tx[0].QueryRow(ctx, query, args...).Scan(&user.ID)
	}

	return r.Pool().QueryRow(ctx, query, args...).Scan(&user.ID)
}

func (r *UserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.UserPublicCols...).
		From(domain.TableUsers).
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

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := r.QueryBuilder().
		Select(domain.UserPublicCols...).
		From(domain.TableUsers).
		Where(squirrel.Eq{domain.ColEmail: email}).
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
