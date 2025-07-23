package postgres

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Pool() *pgxpool.Pool
	QueryBuilder() squirrel.StatementBuilderType
}

type postgresConn struct {
	PgxPool    *pgxpool.Pool
	SQLBuilder squirrel.StatementBuilderType
}

func (c *postgresConn) Pool() *pgxpool.Pool {
	return c.PgxPool
}

func (c *postgresConn) QueryBuilder() squirrel.StatementBuilderType {
	return c.SQLBuilder
}

func MustNewPostgresConnection(dsn string, options ...Option) Client {
	c, err := NewPostgresConnection(dsn, options...)
	if err != nil {
		panic(err)
	}

	return c
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

	pgc := &postgresConn{
		SQLBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		PgxPool:    pool,
	}

	return pgc, nil
}
