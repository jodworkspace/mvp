package tracing

import (
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "gitlab.com/jodworkspace/mvp/pkg/monitor/tracing"

type TracedRedisClient struct {
	client *redis.Client
	tracer trace.Tracer
}

func NewTracedRedisClient(client *redis.Client) *redis.Client {
	hooks := &redisHooks{
		tracer: otel.GetTracerProvider().Tracer(tracerName),
	}

	client.AddHook(hooks)
	return client
}

type redisHooks struct {
	tracer trace.Tracer
}

func (r redisHooks) DialHook(next redis.DialHook) redis.DialHook {
	//TODO implement me
	panic("implement me")
}

func (r redisHooks) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	//TODO implement me
	panic("implement me")
}

func (r redisHooks) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	//TODO implement me
	panic("implement me")
}
