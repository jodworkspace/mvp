package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	registry *prometheus.Registry
	provider *metric.MeterProvider
)

func Init() error {
	registry = prometheus.NewRegistry()

	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return err
	}

	provider = metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	return nil
}

func PrometheusHandler() http.Handler {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	opts := promhttp.HandlerOpts{
		Registry: registry,
	}

	return promhttp.HandlerFor(registry, opts)
}
