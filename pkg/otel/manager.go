package otel

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
)

type Config struct {
	ServiceName    string
	MetricInterval time.Duration
}

type Manager struct {
	cfg          *Config
	grpcConn     *grpc.ClientConn
	promRegistry *prometheus.Registry
}

type Option func(manager *Manager)

func WithGRPCConn(conn *grpc.ClientConn) Option {
	return func(manager *Manager) {
		manager.grpcConn = conn
	}
}

func WithCustomPrometheus() Option {
	return func(manager *Manager) {
		registry := prometheus.NewRegistry()
		registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewBuildInfoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		manager.promRegistry = registry
	}
}

func NewManager(cfg *Config, opts ...Option) *Manager {
	m := &Manager{
		cfg: cfg,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) SetupOtelSDK(ctx context.Context) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(m.cfg.ServiceName),
		))
	if err != nil {
		return nil, err
	}

	var shutdownFuncs []func(context.Context) error

	meterProvider, err := m.newMeterProvider(ctx, res)
	if err != nil {
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := m.newTracerProvider(ctx, res)
	if err != nil {
		return shutdown(shutdownFuncs...), err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	loggerProvider, err := newLoggerProvider()
	if err != nil {
		return shutdown(shutdownFuncs...), err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown(shutdownFuncs...), nil
}

func (m *Manager) newMeterProvider(ctx context.Context, r *resource.Resource) (*sdkmetric.MeterProvider, error) {
	promExporter, err := otelprom.New(
		otelprom.WithRegisterer(m.promRegistry),
	)
	if err != nil {
		return nil, err
	}

	otlpExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(m.grpcConn),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(r),
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter,
			sdkmetric.WithInterval(m.cfg.MetricInterval),
		)),
	)

	return meterProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (m *Manager) newTracerProvider(ctx context.Context, r *resource.Resource) (*sdktrace.TracerProvider, error) {
	otlpExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(m.grpcConn),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(r),
		sdktrace.WithBatcher(otlpExporter, sdktrace.WithBatchTimeout(5*time.Second)),
	)

	return tracerProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	return loggerProvider, nil
}

func shutdown(shutdownFn ...func(context.Context) error) func(ctx context.Context) error {
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
		MaxRequestsInFlight: 10,
		Registry:            m.promRegistry,
	}

	return promhttp.HandlerFor(m.promRegistry, opts)
}
