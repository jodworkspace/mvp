package instrument

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	pgxmetrics "gitlab.com/jodworkspace/mvp/pkg/monitor"
)

type InstrumentedPostgresClient struct {
	postgres.Client
	metrics *pgxmetrics.PgxMonitor
}

func NewInstrumentedPostgresClient(client postgres.Client, metrics *pgxmetrics.PgxMonitor) *InstrumentedPostgresClient {
	return &InstrumentedPostgresClient{
		Client:  client,
		metrics: metrics,
	}
}

func (ic *InstrumentedPostgresClient) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	op := opFromSQL(sql)
	return ic.metrics.InstrumentQuery(ctx, ic.Pool(), op, sql, args...)
}

func (ic *InstrumentedPostgresClient) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	op := opFromSQL(sql)
	return ic.metrics.InstrumentExec(ctx, ic.Pool(), op, sql, args...)
}

func (ic *InstrumentedPostgresClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	op := opFromSQL(sql)
	return ic.metrics.InstrumentQueryRow(ctx, ic.Pool(), op, sql, args...)
}

func opFromSQL(sql string) string {
	if len(sql) >= 6 {
		prefix := sql[:6]
		switch {
		case prefix == "SELECT":
			return "SELECT"
		case prefix == "INSERT":
			return "INSERT"
		case prefix == "UPDATE":
			return "UPDATE"
		case prefix == "DELETE":
			return "DELETE"
		}
	}
	return "OTHER"
}
