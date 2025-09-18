package logging

import (
	"errors"
	"log/slog"
	"strings"
)

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
