package postgresrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/gookie/mvp/pkg/db"
)

func exists(db db.PostgresConn, ctx context.Context, table string, col string, val any) (bool, error) {
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
