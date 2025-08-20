package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/pgx"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/request"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
)

type Manager struct {
	serviceName string
	conn        *grpc.ClientConn

	registry *prometheus.Registry
	provider *sdkmetric.MeterProvider
}

func NewManager(serviceName string, conn *grpc.ClientConn) *Manager {
	return &Manager{
		serviceName: serviceName,
		conn:        conn,
	}
}

func (m *Manager) Init(ctx context.Context) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(m.serviceName),
		))
	if err != nil {
		return nil, err
	}

	m.registry = prometheus.NewRegistry()
	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	promExporter, err := otelprom.New(
		otelprom.WithRegisterer(m.registry),
	)
	if err != nil {
		return nil, err
	}

	otlpExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(m.conn),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	m.provider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter)),
	)
	otel.SetMeterProvider(m.provider)

	return m.provider.Shutdown, nil
}

func (m *Manager) NewHTTPMetrics() (*request.HTTPMetrics, error) {
	metricsName := fmt.Sprintf("%s/%s", m.serviceName, "/monitor/metrics/http")
	meter := m.provider.Meter(metricsName)

	requestsCounter, err := meter.Int64Counter("http_request_count")
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram("http_request_duration_seconds")
	if err != nil {
		return nil, err
	}

	return request.NewHTTPMetrics(requestsCounter, requestDuration), nil
}

func (m *Manager) NewPgxMetrics() (*pgx.Metrics, error) {
	metricsName := fmt.Sprintf("%s/%s", m.serviceName, "/monitor/metrics/pgx")
	meter := m.provider.Meter(metricsName)

	queryCounter, err := meter.Int64Counter("db_query_total")
	if err != nil {
		return nil, err
	}

	queryDuration, err := meter.Float64Histogram("db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2),
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

	return pgx.NewMetrics(activeConnections, queryCounter, queryDuration), nil
}

func (m *Manager) PrometheusHandler() http.Handler {
	opts := promhttp.HandlerOpts{
		Registry: m.registry,
	}

	return promhttp.HandlerFor(m.registry, opts)
}
