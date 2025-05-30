package xlog

import "context"

type key int

const (
	xlogKey key = 0
)

func NewContext(ctx context.Context, xl *Logger) context.Context {
	return context.WithValue(ctx, xlogKey, xl)
}

func FromContext(ctx context.Context) (xl *Logger, ok bool) {
	xl, ok = ctx.Value(xlogKey).(*Logger)
	return
}

func FromContextSafe(ctx context.Context) *Logger {
	xl, ok := ctx.Value(xlogKey).(*Logger)
	if !ok {
		xl = New()
	}
	return xl
}
