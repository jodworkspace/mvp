package helper

import "context"

func ContextWithValues(ctx context.Context, kv map[string]any) context.Context {
	for k, v := range kv {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}
