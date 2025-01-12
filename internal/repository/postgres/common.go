package postgresrepo

import (
	"context"
	"gitlab.com/tokpok/mvp/pkg/db"
)

func exists(db *db.PostgresConnection, ctx context.Context, tabName string, colName string, val any) (bool, error) {
	query, args, err := db.SQLBuilder.
		Select("1").
		Prefix("SELECT EXISTS(").
		From(tabName).
		Suffix(")").
		ToSql()
	if err != nil {
		return false, err
	}

	var exists bool
	err = db.Pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}
