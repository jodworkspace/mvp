package postgresrepo

import (
	"context"

	"github.com/Masterminds/squirrel"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
)

func exists(db postgres.Client, ctx context.Context, table string, col string, val any) (bool, error) {
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

	var found bool
	err = db.Pool().QueryRow(ctx, query, args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}
