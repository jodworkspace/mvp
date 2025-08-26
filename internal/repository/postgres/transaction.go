package postgresrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
)

type TransactionManager struct {
	client postgres.Client
}

func NewTransactionManager(pgc postgres.Client) *TransactionManager {
	return &TransactionManager{
		client: pgc,
	}
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, txFunc func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := tm.client.Pool().Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	return txFunc(ctx, tx)
}
