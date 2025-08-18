package metrics

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	dbMetricsName    = "gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/db"
	dbQueryCounter   metric.Int64Counter
	dbQueryHistogram metric.Float64Histogram
)

func InitDBMetrics() {
	meter := provider.Meter(dbMetricsName)
	dbQueryCounter, _ = meter.Int64Counter("db_query_total")
	dbQueryHistogram, _ = meter.Float64Histogram("db_query_duration_seconds")
}

func InstrumentQuery(ctx context.Context, db *sql.DB, operation, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	dbQueryCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", operation)))
	dbQueryHistogram.Record(ctx, duration, metric.WithAttributes(attribute.String("operation", operation)))

	return rows, err
}
