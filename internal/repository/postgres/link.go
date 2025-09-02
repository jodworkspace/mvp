package postgresrepo

import (
	"context"

	"github.com/Masterminds/squirrel"
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

func (r *LinkRepository) Get(ctx context.Context, userID, issuer string) (*domain.Link, error) {
	query, args, err := r.client.QueryBuilder().
		Select(domain.LinkAllCols...).
		From(domain.TableLinks).
		Where(squirrel.And{
			squirrel.Eq{domain.ColUserID: userID},
			squirrel.Eq{domain.ColIssuer: issuer},
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var link domain.Link
	err = r.client.Pool().QueryRow(ctx, query, args...).Scan(
		&link.UserID,
		&link.Issuer,
		&link.ExternalID,
		&link.AccessToken,
		&link.RefreshToken,
		&link.AccessTokenExpiredAt,
		&link.RefreshTokenExpiredAt,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) Update(ctx context.Context, link *domain.Link) error {
	query, args, err := r.client.QueryBuilder().
		Update(domain.TableLinks).
		Set(domain.ColAccessToken, link.AccessToken).
		Set(domain.ColRefreshToken, link.RefreshToken).
		Set(domain.ColAccessTokenExpiresAt, link.AccessTokenExpiredAt).
		Set(domain.ColRefreshTokenExpiresAt, link.RefreshTokenExpiredAt).
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
