package postgresrepo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/internal/domain"
	"gitlab.com/gookie/mvp/pkg/db"
)

type FederatedUserRepository struct {
	db.Postgres
}

func NewFederatedUserRepository(conn db.Postgres) *FederatedUserRepository {
	return &FederatedUserRepository{
		Postgres: conn,
	}
}

func (r *FederatedUserRepository) Insert(ctx context.Context, federatedUser *domain.FederatedUser, tx ...pgx.Tx) error {
	query, args, err := r.QueryBuilder().
		Insert(domain.TableFederatedUser).
		Columns(domain.FederatedUserAllCols...).
		Values(
			federatedUser.ID,
			federatedUser.UserID,
			federatedUser.Issuer,
			federatedUser.ExternalID,
			federatedUser.AccessToken,
			federatedUser.RefreshToken,
			federatedUser.CreatedAt,
			federatedUser.UpdatedAt,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return err
	}

	if len(tx) > 0 {
		return tx[0].QueryRow(ctx, query, args...).Scan(&federatedUser.ID)
	}

	return r.Pool().QueryRow(ctx, query, args...).Scan(&federatedUser.UserID)
}
