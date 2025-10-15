package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type db struct {
	pgxPool    *pgxpool.Pool
	sqlBuilder squirrel.StatementBuilderType
}

func NewPostgresDB(dsn string, options ...Option) (DB, error) {
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(dbConfig)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	pgc := &db{
		pgxPool:    pool,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return pgc, nil
}

func (c *db) Pool() Pool {
	return c.pgxPool
}

func (c *db) QueryBuilder() squirrel.StatementBuilderType {
	return c.sqlBuilder
}
