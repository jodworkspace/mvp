package postgresrepo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
)

type TransactionManager struct {
	db postgres.Client
}

func NewTransactionManager(db postgres.Client) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, txFunc func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := tm.db.Pool().Begin(ctx)
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
