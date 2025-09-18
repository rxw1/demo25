package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level       string
	JSON        bool
	AddSource   bool
	Writer      io.Writer
	Service     string
	Version     string
	Environment string
	SetDefault  bool
}

func New(cfg Config) (*slog.Logger, error) {
	lvl, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	hopts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			key := a.Key
			if key == slog.LevelKey {
				if v, ok := a.Value.Any().(slog.Level); ok {
					return slog.Attr{Key: "level", Value: slog.StringValue(v.String())}
				}
				return slog.Attr{Key: "level", Value: slog.StringValue(strings.ToLower(a.Value.String()))}
			}
			if isSensitiveKey(key) {
				return slog.Attr{Key: key, Value: slog.StringValue("[REDACTED]")}
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(w, hopts)
	} else {
		handler = slog.NewTextHandler(w, hopts)
	}

	base := slog.New(handler)
	if cfg.Service != "" || cfg.Version != "" || cfg.Environment != "" {
		base = base.With(
			slog.String("service", cfg.Service),
			slog.String("version", cfg.Version),
			slog.String("env", cfg.Environment),
		)
	}

	if cfg.SetDefault {
		slog.SetDefault(base)
	}
	return base, nil
}

func parseLevel(s string) (slog.Leveler, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "err", "error":
		return slog.LevelError, nil
	default:
		return nil, errors.New("unknown log level: " + s)
	}
}

func isSensitiveKey(k string) bool {
	k = strings.ToLower(k)
	suspects := []string{"password", "passwd", "secret", "token", "authorization", "api_key", "apikey", "cookie"}
	for _, s := range suspects {
		if k == s || strings.HasSuffix(k, "_"+s) {
			return true
		}
	}
	return false
}

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
