package db

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres interface {
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

func MustNewPostgresConnection(dsn string, options ...PostgresOption) Postgres {
	c, err := NewPostgresConnection(dsn, options...)
	if err != nil {
		panic(err)
	}
	return c
}

func NewPostgresConnection(dsn string, options ...PostgresOption) (Postgres, error) {
	pgc := &postgresConn{
		SQLBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(dbConfig)
	}

	pgc.PgxPool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err == nil {
		return pgc, nil
	}

	return nil, err
}

type PostgresOption func(*pgxpool.Config)

func WithMinConns(minConns int32) PostgresOption {
	return func(config *pgxpool.Config) {
		config.MinConns = minConns
	}
}

func WithMaxConns(maxConns int32) PostgresOption {
	return func(config *pgxpool.Config) {
		config.MaxConns = maxConns
	}
}
