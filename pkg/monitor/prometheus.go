package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type PrometheusClient struct {
	registry *prometheus.Registry
}

func NewPrometheusClient() *PrometheusClient {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &PrometheusClient{
		registry: registry,
	}
}

func (p *PrometheusClient) HTTPHandler() http.Handler {
	opts := promhttp.HandlerOpts{
		Registry: p.registry,
	}

	return promhttp.HandlerFor(p.registry, opts)
}
