package metrics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	httpMetricsName     = "gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/http"
	httpRequestsCounter metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
)

func InitHTTPMetrics() {
	meter := provider.Meter(httpMetricsName)
	httpRequestsCounter, _ = meter.Int64Counter("http_request_count")
	httpRequestDuration, _ = meter.Float64Histogram("http_request_duration_seconds")
}

func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(rr, r)
		duration := time.Since(start).Seconds()

		routePattern := chi.RouteContext(r.Context()).RoutePattern()

		httpRequestsCounter.Add(r.Context(), 1, metric.WithAttributes(
			attribute.String("route", routePattern),
			attribute.String("method", r.Method),
			attribute.Int("status_code", rr.status),
		))

		httpRequestDuration.Record(r.Context(), duration, metric.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("route", routePattern),
		))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}
