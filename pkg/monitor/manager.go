package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
)

type Manager struct {
	serviceName        string
	metricInterval     time.Duration
	conn               *grpc.ClientConn
	prometheusRegistry *prometheus.Registry
	meterProvider      *sdkmetric.MeterProvider
	tracerProvider     *sdktrace.TracerProvider
}

func NewManager(serviceName string, metricInterval time.Duration, conn *grpc.ClientConn) *Manager {
	return &Manager{
		serviceName:    serviceName,
		metricInterval: metricInterval,
		conn:           conn,
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

	registry := prometheus.NewRegistry()
	registry.MustRegister(
	//collectors.NewGoCollector(),
	//collectors.NewBuildInfoCollector(),
	//collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	m.prometheusRegistry = registry

	meterProvider, err := m.initMeter(ctx, res, registry)
	if err != nil {
		return nil, err
	}
	m.meterProvider = meterProvider
	otel.SetMeterProvider(meterProvider)

	tracerProvider, err := m.initTracer(ctx, res)
	if err != nil {
		return nil, err
	}
	m.tracerProvider = tracerProvider
	otel.SetTracerProvider(tracerProvider)

	return shutdownMonitor(meterProvider.Shutdown, tracerProvider.Shutdown), nil
}

func (m *Manager) initMeter(ctx context.Context, r *resource.Resource, reg *prometheus.Registry) (*sdkmetric.MeterProvider, error) {
	promExporter, err := otelprom.New(
		otelprom.WithRegisterer(reg),
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

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(r),
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter,
			sdkmetric.WithInterval(m.metricInterval),
		)),
	)
	otel.SetMeterProvider(m.meterProvider)

	return meterProvider, nil
}

func (m *Manager) initTracer(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	otlpExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(m.conn),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlpExporter, sdktrace.WithBatchTimeout(5*time.Second)),
	)
	otel.SetTracerProvider(m.tracerProvider)

	return tracerProvider, nil
}

func shutdownMonitor(shutdownFn ...func(context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var shutDownErrs error

		for _, fn := range shutdownFn {
			shutDownErr := fn(ctx)
			if shutDownErr != nil {
				shutDownErrs = errors.Join(shutDownErrs, shutDownErr)
			}
		}

		return shutDownErrs
	}
}

func (m *Manager) PrometheusHandler() http.Handler {
	opts := promhttp.HandlerOpts{
		Registry: m.prometheusRegistry,
	}

	return promhttp.HandlerFor(m.prometheusRegistry, opts)
}

func (m *Manager) NewHTTPMonitor() (*HTTPMonitor, error) {
	name := fmt.Sprintf("%s/%s", m.serviceName, "/monitor/http")
	meter := m.meterProvider.Meter(name)
	tracer := m.tracerProvider.Tracer(name)

	return NewHTTPMonitor(meter, tracer)
}

func (m *Manager) NewPgxMonitor() (*PgxMonitor, error) {
	name := fmt.Sprintf("%s/%s", m.serviceName, "/monitor/pgx")
	meter := m.meterProvider.Meter(name)
	tracer := m.tracerProvider.Tracer(name)

	return NewPgxMonitor(meter, tracer)
}
