package monitor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type pgxPool interface {
	Stat() *pgxpool.Stat
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type pgxMetrics struct {
	activeConnections metric.Int64UpDownCounter
	queryCounter      metric.Int64Counter
	queryDuration     metric.Float64Histogram
}

type PgxMonitor struct {
	meter  metric.Meter
	tracer trace.Tracer
	pgxMetrics
}

func NewPgxMonitor(meter metric.Meter, tracer trace.Tracer) (*PgxMonitor, error) {
	queryCounter, err := meter.Int64Counter("db_query_total")
	if err != nil {
		return nil, err
	}

	queryDuration, err := meter.Float64Histogram("db_query_duration_milliseconds",
		metric.WithDescription("Database query duration in milliseconds"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 500, 1000, 2000),
	)
	if err != nil {
		return nil, err
	}

	activeConnections, err := meter.Int64UpDownCounter(
		"active_connections",
		metric.WithDescription("Number of active connections"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return &PgxMonitor{
		meter:  meter,
		tracer: tracer,
		pgxMetrics: pgxMetrics{
			activeConnections: activeConnections,
			queryCounter:      queryCounter,
			queryDuration:     queryDuration,
		},
	}, nil
}

func (d *PgxMonitor) InstrumentQuery(ctx context.Context, p pgxPool, op, sql string, args ...any) (pgx.Rows, error) {
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

func (d *PgxMonitor) InstrumentQueryRow(ctx context.Context, p pgxPool, op, sql string, args ...any) pgx.Row {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	row := p.QueryRow(ctx, sql, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", op)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", op)))

	return row
}

func (d *PgxMonitor) InstrumentExec(ctx context.Context, p pgxPool, op, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	d.activeConnections.Add(ctx, 1)
	result, err := p.Exec(ctx, sql, args...)
	duration := time.Since(start).Seconds()

	d.activeConnections.Add(ctx, -1)
	d.queryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", op)))
	d.queryDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", op)))

	return result, err
}
