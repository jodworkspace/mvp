package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type pgxPool interface {
	Stat() *pgxpool.Stat
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Metrics struct {
	activeConnections metric.Int64UpDownCounter
	queryCounter      metric.Int64Counter
	queryDuration     metric.Float64Histogram
}

func NewMetrics(
	activeConnections metric.Int64UpDownCounter,
	queryCounter metric.Int64Counter,
	queryDuration metric.Float64Histogram,
) *Metrics {
	return &Metrics{
		activeConnections: activeConnections,
		queryCounter:      queryCounter,
		queryDuration:     queryDuration,
	}
}

func (d *Metrics) InstrumentQuery(ctx context.Context, p pgxPool, op, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	rows, err := p.Query(ctx, sql, args...)
	duration := time.Since(start).Seconds()

	// stat := p.Stat()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", op)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", op)))

	return rows, err
}

func (d *Metrics) InstrumentQueryRow(ctx context.Context, p pgxPool, op, sql string, args ...any) pgx.Row {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	row := p.QueryRow(ctx, sql, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", op)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", op)))

	return row
}

func (d *Metrics) InstrumentExec(ctx context.Context, p pgxPool, op, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	result, err := p.Exec(ctx, sql, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", op)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", op)))

	return result, err
}
