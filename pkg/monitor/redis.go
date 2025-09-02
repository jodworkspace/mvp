package monitor

import "go.opentelemetry.io/otel/metric"

type RedisMonitor struct {
	activeConnections metric.Int64UpDownCounter
	queryCounter      metric.Int64Counter
	queryDuration     metric.Float64Histogram
}

func newRedisMonitor(
	activeConnections metric.Int64UpDownCounter,
	queryCounter metric.Int64Counter,
	queryDuration metric.Float64Histogram,
) *RedisMonitor {
	return &RedisMonitor{
		activeConnections: activeConnections,
		queryCounter:      queryCounter,
		queryDuration:     queryDuration,
	}
}
