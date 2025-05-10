package postgresrepo

import (
	"context"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type FederatedUserRepository struct {
	db.Postgres
}

func NewFederatedUserRepository(conn db.Postgres) *FederatedUserRepository {
	return &FederatedUserRepository{conn}
}

func (r *FederatedUserRepository) Insert(ctx context.Context, user *domain.FederatedUser) error {
	query, args, err := r.QueryBuilder().
		Insert(domain.TableFederatedUser).
		Columns(domain.FederatedUserAllCols...).
		Values(
			user.ID,
			user.UserID,
			user.Provider,
			user.ExternalID,
			user.AccessToken,
			user.RefreshToken,
			user.CreatedAt,
			user.UpdatedAt,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return err
	}

	return r.Pool().QueryRow(ctx, query, args...).Scan(&user.UserID)
}
