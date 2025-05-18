package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.com/gookie/mvp/pkg/db"
)

func exists(db db.Postgres, ctx context.Context, table string, col string, val any) (bool, error) {
	query, args, err := db.QueryBuilder().
		Select("1").
		Prefix("SELECT EXISTS(").
		From(table).
		Suffix(")").
		Where(squirrel.Eq{col: val}).
		ToSql()
	if err != nil {
		return false, err
	}

	var exists bool
	err = db.Pool().QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

type TransactionManager struct {
	db db.Postgres
}

func NewTransactionManager(db db.Postgres) *TransactionManager {
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
