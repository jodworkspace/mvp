package postgresrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
)

type LinkRepository struct {
	client postgres.Client
}

func NewLinkRepository(pgc postgres.Client) *LinkRepository {
	return &LinkRepository{
		client: pgc,
	}
}

func (r *LinkRepository) Insert(ctx context.Context, link *domain.Link, tx ...pgx.Tx) error {
	query, args, err := r.client.QueryBuilder().
		Insert(domain.TableLinks).
		Columns(domain.LinkAllCols...).
		Values(
			link.UserID,
			link.Issuer,
			link.ExternalID,
			link.AccessToken,
			link.RefreshToken,
			link.AccessTokenExpiredAt,
			link.RefreshTokenExpiredAt,
			link.CreatedAt,
			link.UpdatedAt,
		).
		Suffix("RETURNING " + domain.ColUserID).
		ToSql()
	if err != nil {
		return err
	}

	if len(tx) > 0 {
		return tx[0].QueryRow(ctx, query, args...).Scan(&link.UserID)
	}

	return r.client.Pool().QueryRow(ctx, query, args...).Scan(&link.UserID)
}
