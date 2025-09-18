package logging

import (
	"log/slog"
	"os"
	"strings"
)

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
