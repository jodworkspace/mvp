package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type client struct {
	pgxPool    *pgxpool.Pool
	sqlBuilder squirrel.StatementBuilderType
}

func NewPostgresConnection(dsn string, options ...Option) (Client, error) {
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(dbConfig)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	pgc := &client{
		pgxPool:    pool,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return pgc, nil
}

func (c *client) Stat() *pgxpool.Stat {
	return c.pgxPool.Stat()
}

func (c *client) Pool() Pool {
	return c.pgxPool
}

func (c *client) QueryBuilder() squirrel.StatementBuilderType {
	return c.sqlBuilder
}

func (c *client) Close() {
	c.pgxPool.Close()
}

func (c *client) Ping(ctx context.Context) error {
	return c.pgxPool.Ping(ctx)
}

func (c *client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.pgxPool.Begin(ctx)
}

func (c *client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return c.pgxPool.Query(ctx, sql, args...)
}

func (c *client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.pgxPool.QueryRow(ctx, sql, args...)
}

func (c *client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return c.pgxPool.Exec(ctx, sql, args...)
}
