// Package logging provides an opinionated, stdlib-first setup for slog.
// Goals:
//   - App decides formatting/output; libraries accept *slog.Logger or slog.Handler.
//   - Single initialization point, no hidden globals (optionally set slog default).
//   - Context-friendly helpers and request-scoped fields.
//   - ReplaceAttr hook for redaction and consistent field names.
//   - Reasonable dev (text) vs prod (JSON) defaults.
package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config controls logger construction. Keep it small and boring.
type Config struct {
	// Level: "debug", "info", "warn", "error" (case-insensitive).
	Level string
	// JSON forces JSON output; if false, a human-friendly text handler is used.
	JSON bool
	// AddSource includes file:line. Useful in dev; costly in hot paths.
	AddSource bool
	// Writer selects the destination (defaults to os.Stdout).
	Writer io.Writer
	// Service metadata you may want on every log line.
	Service     string
	Version     string
	Environment string // e.g., dev, staging, prod
	// SetDefault also sets slog.SetDefault(logger) for packages using slog.Default().
	SetDefault bool
}

// New constructs a *slog.Logger from Config. It never panics.
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
			// Normalize level key and redact known sensitive keys.
			// You can expand this list to fit your domain.
			key := a.Key
			if key == slog.LevelKey {
				// Emit level as lowercase string for consistency.
				if v, ok := a.Value.Any().(slog.Level); ok {
					return slog.Attr{Key: "level", Value: slog.StringValue(v.String())}
				}
				return slog.Attr{Key: "level", Value: slog.StringValue(strings.ToLower(a.Value.String()))}
			}
			// Redact secrets by key name heuristics.
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
	// Attach stable, always-on attributes.
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

// parseLevel accepts common strings; default is Info.
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
//
// Prefer passing *slog.Logger explicitly. When that’s noisy (eg HTTP handlers),
// use context to carry a request-scoped child logger. These helpers keep it tidy.

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

// Example: HTTP middleware sketch (framework-agnostic)
//
// func Logging(next http.Handler) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         start := time.Now()
//         // In real code, wrap ResponseWriter to capture status/bytes.
//         reqID := r.Header.Get("X-Request-ID")
//
//         l := slog.Default().With(
//             "trace_id", reqID,
//             "remote_ip", r.RemoteAddr,
//             "method", r.Method,
//             "path", r.URL.Path,
//         )
//         ctx := Into(r.Context(), l)
//         r = r.WithContext(ctx)
//
//         // call next
//         next.ServeHTTP(w, r)
//
//         l.Info("request complete", "duration_ms", time.Since(start).Milliseconds())
//     })
// }

// Example: timed helper for ad-hoc spans.
func Timed(ctx context.Context, msg string, attrs ...any) func(extra ...any) {
	l := From(ctx)
	start := time.Now()
	l.Debug("start: "+msg, attrs...)
	return func(extra ...any) {
		elapsed := time.Since(start)
		all := append(attrs, append(extra, "duration", elapsed)...) // safe: odd/even OK in slog
		l.Info(msg, all...)
	}
}

// ------- USAGE (in your app main) -------
//
// func main() {
//     cfg := logging.Config{
//         Level:       getenv("LOG_LEVEL", "info"),
//         JSON:        getenv("LOG_FORMAT", "json") == "json",
//         AddSource:   getenv("LOG_SOURCE", "false") == "true",
//         Service:     "payments-api",
//         Version:     buildVersion,
//         Environment: getenv("ENV", "dev"),
//         SetDefault:  true, // so slog.Default() works across the app
//     }
//     logger, err := logging.New(cfg)
//     if err != nil { panic(err) }
//
//     logger.Info("boot", "pid", os.Getpid())
//
//     // Later in handlers:
//     // func h(w http.ResponseWriter, r *http.Request) {
//     //     ctx := logging.With(r.Context(), "user_id", uid)
//     //     logging.From(ctx).Error("db query failed", "err", err)
//     // }
// }

// tiny getenv helper for the usage sketch
// func getenv(k, def string) string {
// 	if v := os.Getenv(k); v != "" {
// 		return v
// 	}
// 	return def
// }
