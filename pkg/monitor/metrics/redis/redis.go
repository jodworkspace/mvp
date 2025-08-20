package redis

import "go.opentelemetry.io/otel/metric"

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
