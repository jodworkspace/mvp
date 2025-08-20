package instrument

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	pgxmetrics "gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/pgx"
)

type InstrumentedClient struct {
	client  postgres.Client
	metrics *pgxmetrics.Metrics
}

func NewInstrumentedClient(client postgres.Client, metrics *pgxmetrics.Metrics) *InstrumentedClient {
	return &InstrumentedClient{
		client:  client,
		metrics: metrics,
	}
}

func (ic *InstrumentedClient) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return ic.metrics.InstrumentQuery(ctx, ic.client.Pool(), "", query, args...)
}

func (ic *InstrumentedClient) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return ic.metrics.InstrumentExec(ctx, ic.client.Pool(), "", query, args...)
}

func (ic *InstrumentedClient) Close() {
	ic.client.Close()
}

func (ic *InstrumentedClient) Ping(ctx context.Context) error {
	return ic.client.Ping(ctx)
}

func (ic *InstrumentedClient) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return ic.Acquire(ctx)
}

func (ic *InstrumentedClient) Begin(ctx context.Context) (pgx.Tx, error) {
	return ic.client.Begin(ctx)
}

func (ic *InstrumentedClient) Pool() *pgxpool.Pool {
	return ic.client.Pool()
}

func (ic *InstrumentedClient) QueryBuilder() squirrel.StatementBuilderType {
	return ic.client.QueryBuilder()
}

func (ic *InstrumentedClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return ic.client.QueryRow(ctx, sql, args...)
}
