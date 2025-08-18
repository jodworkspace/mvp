package tracing

import (
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "gitlab.com/jodworkspace/mvp/pkg/monitor/tracing"

type TracedRedisClient struct {
	client goredis.Cmdable
	tracer trace.Tracer
}
