package monitor

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/propagation"
)

var propagator = propagation.TraceContext{}

type httpTracing struct {
}

type httpMetrics struct {
	requestsCounter metric.Int64Counter
	requestDuration metric.Int64Histogram
}

type HTTPMonitor struct {
	meter  metric.Meter
	tracer trace.Tracer
	httpMetrics
	httpTracing
}

func NewHTTPMonitor(meter metric.Meter, tracer trace.Tracer) (*HTTPMonitor, error) {
	requestsCounter, err := meter.Int64Counter("http_request_count",
		metric.WithDescription("Number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Int64Histogram("http_request_duration_milliseconds",
		metric.WithDescription("Request processing duration in milliseconds"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 500, 1000, 2000))
	if err != nil {
		return nil, err
	}

	return &HTTPMonitor{
		meter:  meter,
		tracer: tracer,
		httpMetrics: httpMetrics{
			requestsCounter: requestsCounter,
			requestDuration: requestDuration,
		},
	}, nil
}

type writerRecorder struct {
	http.ResponseWriter
	status int
}

func (w *writerRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (h *HTTPMonitor) Handle(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := h.tracer.Start(ctx, "http-server",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		wr := &writerRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(wr, r.WithContext(ctx))
		duration := time.Since(start).Milliseconds()

		h.requestsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("route", route),
			attribute.String("method", r.Method),
			attribute.Int("status_code", wr.status),
		))
		h.requestDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("route", route),
			attribute.String("method", r.Method),
			attribute.Int("status_code", wr.status),
		))

		// Add route info to span
		span.SetAttributes(
			attribute.String("route", route),
			attribute.String("method", r.Method),
			attribute.Int("status_code", wr.status),
			attribute.Int64("process_time_ms", duration),
		)
	})
}

type tracingTransport struct {
	transport http.RoundTripper
	tracer    trace.Tracer
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.Start(req.Context(), "http-client")
	defer span.End()

	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	processTime := time.Since(start).Milliseconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	span.SetAttributes(
		attribute.String("external_url", req.URL.String()),
		attribute.String("external_method", req.Method),
		attribute.Int("external_status_code", resp.StatusCode),
		attribute.Int64("external_process_time_ms", processTime),
	)

	return resp, err
}

func (h *HTTPMonitor) NewTracingClient() http.Client {
	client := http.Client{}

	client.Transport = &tracingTransport{
		transport: http.DefaultTransport,
		tracer:    h.tracer,
	}

	return client
}
