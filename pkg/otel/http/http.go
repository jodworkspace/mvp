package otelhttp

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	meterName  = "otel/http"
	tracerName = "otel/http"
)

type Monitor struct {
	requestsCounter metric.Int64Counter
	requestDuration metric.Int64Histogram
	responseSize    metric.Int64Histogram
	errorCounter    metric.Int64Counter
}

func NewMonitor() (*Monitor, error) {
	meter := otel.Meter(meterName)

	requestsCounter, err := meter.Int64Counter("http.client.request.count",
		metric.WithDescription("Number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Int64Histogram("http.client.request.duration.ms",
		metric.WithDescription("Duration of HTTP requests in milliseconds"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 500, 1000, 2000))
	if err != nil {
		return nil, err
	}

	responseSize, err := meter.Int64Histogram("http.client.response.size.bytes",
		metric.WithDescription("Size of HTTP responses in bytes"),
		metric.WithUnit("By"),
		metric.WithExplicitBucketBoundaries(100, 500, 1000, 5000, 10000, 50000, 100000))
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter("http.client.error.count",
		metric.WithDescription("Number of HTTP request errors"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		requestsCounter: requestsCounter,
		requestDuration: requestDuration,
		responseSize:    responseSize,
		errorCounter:    errorCounter,
	}, nil
}

type writerRecorder struct {
	http.ResponseWriter
	status  int
	written int64
}

func (w *writerRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *writerRecorder) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)
	return n, err
}

// Handle provides an HTTP middleware that instruments requests with OpenTelemetry tracing and metrics.
func (h *Monitor) Handle(route string, next http.Handler) http.Handler {
	tracer := otel.Tracer(tracerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		propagator := otel.GetTextMapPropagator()

		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, route, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		recorder := &writerRecorder{ResponseWriter: w, status: 200}
		start := time.Now()
		next.ServeHTTP(recorder, r.WithContext(ctx))
		duration := time.Since(start).Milliseconds()

		attrs := []attribute.KeyValue{
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.URLPathKey.String(r.URL.Path),
			semconv.HTTPResponseStatusCodeKey.Int(recorder.status),
			semconv.NetworkPeerAddressKey.String(r.URL.Hostname()),
		}

		h.requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		h.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
		h.responseSize.Record(ctx, recorder.written, metric.WithAttributes(attrs...))
		if recorder.status >= 400 {
			h.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		}

		span.SetAttributes(
			attribute.String("route", route),
			attribute.Int64("process_time_ms", duration),
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.HTTPResponseStatusCodeKey.Int(recorder.status),
			semconv.HTTPResponseBodySizeKey.Int64(recorder.written),
		)
	})
}

type tracingTransport struct {
	transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface, starting a span for the outgoing request and injecting context.
func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tracer := otel.Tracer(tracerName)

	spanName := fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)
	ctx, span := tracer.Start(req.Context(), spanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	parent := trace.SpanContextFromContext(req.Context())
	if !parent.IsValid() {
		return t.transport.RoundTrip(req)
	}

	propagator := otel.GetTextMapPropagator()
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
		attribute.String("url", req.URL.String()),
		attribute.String("method", req.Method),
		attribute.Int("status_code", resp.StatusCode),
		attribute.Int64("process_time_ms", processTime),
	)

	return resp, err
}

func (h *Monitor) TransportWithTracing() http.RoundTripper {
	return &tracingTransport{
		transport: http.DefaultTransport,
	}
}
