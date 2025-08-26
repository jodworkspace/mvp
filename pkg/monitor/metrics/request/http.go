package request

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	requestsCounter metric.Int64Counter
	requestDuration metric.Int64Histogram
}

func NewHTTPMetrics(
	requestsCounter metric.Int64Counter,
	requestDuration metric.Int64Histogram,
) *HTTPMetrics {
	return &HTTPMetrics{
		requestsCounter: requestsCounter,
		requestDuration: requestDuration,
	}
}

func (h *HTTPMetrics) Handle(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(rr, r)
		duration := time.Since(start).Milliseconds()

		h.requestsCounter.Add(r.Context(), 1, metric.WithAttributes(
			attribute.String("route", route),
			attribute.String("method", r.Method),
			attribute.Int("status_code", rr.status),
		))

		h.requestDuration.Record(r.Context(), duration, metric.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("route", route),
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
