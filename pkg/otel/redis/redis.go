package otelredis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	meterName  = "otel/redis"
	tracerName = "otel/redis"
)

type Monitor struct {
	activeConnections metric.Int64UpDownCounter
	queryCounter      metric.Int64Counter
	queryDuration     metric.Float64Histogram
}

func NewMonitor() (*Monitor, error) {
	meter := otel.Meter(meterName)

	activeConnections, err := meter.Int64UpDownCounter("db.client.connections.active",
		metric.WithDescription("Number of active Redis client connections"),
	)
	if err != nil {
		return nil, err
	}

	queryCounter, err := meter.Int64Counter("db.client.query.count",
		metric.WithDescription("Number of Redis commands executed"),
	)
	if err != nil {
		return nil, err
	}

	queryDuration, err := meter.Float64Histogram("db.client.query.duration.ms",
		metric.WithDescription("Redis command duration in milliseconds"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 500, 1000, 2000),
	)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		activeConnections: activeConnections,
		queryCounter:      queryCounter,
		queryDuration:     queryDuration,
	}, nil
}

// TODO

type Hook struct {
	monitor *Monitor
}

func NewHook(m *Monitor) *Hook {
	return &Hook{monitor: m}
}

func (h *Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "redis.command",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemNameRedis,
			semconv.DBOperationNameKey.String(cmd.Name()),
		),
	)
	return context.WithValue(ctx, spanKey{}, span), nil
}

func (h *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := ctx.Value(spanKey{}).(trace.Span)
	if cmd.Err() != nil && !errors.Is(cmd.Err(), redis.Nil) {
		span.RecordError(cmd.Err())
		span.SetStatus(codes.Error, cmd.Err().Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()

	attrs := []attribute.KeyValue{
		semconv.DBOperationNameKey.String(cmd.Name()),
	}

	h.monitor.queryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	return nil
}

func (h *Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "redis.pipeline",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	return context.WithValue(ctx, spanKey{}, span), nil
}

func (h *Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := ctx.Value(spanKey{}).(trace.Span)
	for _, cmd := range cmds {
		if cmd.Err() != nil && !errors.Is(cmd.Err(), redis.Nil) {
			span.RecordError(cmd.Err())
			span.SetStatus(codes.Error, cmd.Err().Error())
		}
	}
	span.End()
	h.monitor.queryCounter.Add(ctx, int64(len(cmds)))

	return nil
}

type spanKey struct{}
