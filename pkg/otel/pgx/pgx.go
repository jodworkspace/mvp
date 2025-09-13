package otelpgx

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	meterName  = "otel/pgx"
	tracerName = "otel/pgx"
)

type Monitor struct {
	connections   metric.Int64UpDownCounter
	queryCount    metric.Int64Counter
	queryDuration metric.Float64Histogram
	poolIdle      metric.Int64ObservableGauge
	poolAcquired  metric.Int64ObservableGauge
	poolMax       metric.Int64ObservableGauge
	poolWaiting   metric.Int64ObservableGauge
}

func NewMonitor() (*Monitor, error) {
	meter := otel.Meter(meterName)

	queryCount, err := meter.Int64Counter(
		"db.client.query.count",
		metric.WithDescription("Number of database queries executed"),
	)
	if err != nil {
		return nil, err
	}

	queryDuration, err := meter.Float64Histogram(
		"db.client.query.duration.ms",
		metric.WithDescription("Database query duration in milliseconds"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 500, 1000, 2000),
	)
	if err != nil {
		return nil, err
	}

	activeConnections, err := meter.Int64UpDownCounter(
		"db.client.connections.active",
		metric.WithDescription("Number of active database client connections"),
	)
	if err != nil {
		return nil, err
	}

	idleConnections, err := meter.Int64ObservableGauge(
		"db.client.connections.idle",
		metric.WithDescription("Number of idle database client connections"),
	)
	if err != nil {
		return nil, err
	}

	acquiredConnections, err := meter.Int64ObservableGauge(
		"db.client.connections.acquired",
		metric.WithDescription("Number of acquired database client connections"),
	)
	if err != nil {
		return nil, err
	}

	maxConnections, err := meter.Int64ObservableGauge(
		"db.client.connections.max",
		metric.WithDescription("Maximum number of database client connections allowed"),
	)
	if err != nil {
		return nil, err
	}

	waitingConnections, err := meter.Int64ObservableGauge(
		"db.client.connections.waiting",
		metric.WithDescription("Number of database client connections waiting for a connection"),
	)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		queryCount:    queryCount,
		queryDuration: queryDuration,
		connections:   activeConnections,
		poolIdle:      idleConnections,
		poolAcquired:  acquiredConnections,
		poolMax:       maxConnections,
		poolWaiting:   waitingConnections,
	}, nil
}

func (m *Monitor) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	tracer := otel.Tracer(tracerName)

	operation := opFromSQL(data.SQL)
	ctx, span := tracer.Start(ctx, strings.ToLower(operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemNamePostgreSQL,
			semconv.DBQueryTextKey.String(data.SQL),
			semconv.DBOperationNameKey.String(operation),
		),
	)

	return trace.ContextWithSpan(ctx, span)
}

func (m *Monitor) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		if data.Err != nil {
			span.RecordError(data.Err)
			span.SetStatus(codes.Error, data.Err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
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
	return "PGX_UNKNOWN"
}
