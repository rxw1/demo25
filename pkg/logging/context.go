package logging

import (
	"context"
	"log/slog"
)

// Context wiring
// Prefer passing *slog.Logger explicitly. When that’s noisy (eg HTTP handlers),
// use context to carry a request-scoped child logger. These helpers keep it
// tidy.

type ctxKey struct{}

// Into attaches l to ctx for later retrieval.
func Into(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From retrieves a *slog.Logger from ctx or returns slog.Default().
func From(ctx context.Context) *slog.Logger {
	if v := ctx.Value(ctxKey{}); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}

// With returns a context holding a child logger with added attributes.
func With(ctx context.Context, attrs ...any) context.Context {
	return Into(ctx, From(ctx).With(attrs...))
}
