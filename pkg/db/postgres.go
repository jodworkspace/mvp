package db

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/tokpok/mvp/config"
	"time"
)

type PostgresConnection struct {
	Pool       *pgxpool.Pool
	SQLBuilder squirrel.StatementBuilderType
}

type PostgresOptions struct {
	MaxPoolSize        int32
	ConnectionAttempts int
	RetryTimeout       time.Duration
}

const (
	defaultMaxPoolSize        = 10
	defaultConnectionAttempts = 5
	defaultRetryTimeout       = 5 * time.Second
)

func applyDefaults(opts PostgresOptions) {
	if opts.MaxPoolSize == 0 {
		opts.MaxPoolSize = defaultMaxPoolSize
	}
	if opts.ConnectionAttempts == 0 {
		opts.ConnectionAttempts = defaultConnectionAttempts
	}
	if opts.RetryTimeout == 0 {
		opts.RetryTimeout = defaultRetryTimeout
	}
}

func getConnectionString(postgresConfig *config.PostgresConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		postgresConfig.User,
		postgresConfig.Password,
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Database,
	)
}

func NewPostgresConnection(cfg *config.Config, options PostgresOptions) (*PostgresConnection, error) {
	pgc := &PostgresConnection{
		SQLBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	dsn := getConnectionString(cfg.Postgres)
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	applyDefaults(options)
	dbConfig.MaxConns = options.MaxPoolSize

	for c := options.ConnectionAttempts; c > 0; c-- {
		pgc.Pool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
		if err == nil {
			return pgc, nil
		}

		if c > 1 {
			time.Sleep(options.RetryTimeout)
		}
	}

	return nil, err
}
