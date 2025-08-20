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

func (d *Metrics) InstrumentQuery(ctx context.Context, db *pgxpool.Pool, operation, query string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	rows, err := db.Query(ctx, query, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", operation)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", operation)))

	return rows, err
}

func (d *Metrics) InstrumentExec(ctx context.Context, db *pgxpool.Pool, operation, query string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	result, err := db.Exec(ctx, query, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", operation)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", operation)))

	return result, err
}
