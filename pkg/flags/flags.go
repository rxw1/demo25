package flags

import (
	"context"
	"log/slog"

	"rxw1/logging"

	of "github.com/open-feature/go-sdk/openfeature"
)

type Flags struct {
	client *of.Client
}

func New(clientName string) *Flags {
	return &Flags{
		client: of.NewClient(clientName),
	}
}

func (f *Flags) RedisEnabled(ctx context.Context) bool {
	val, err := f.client.BooleanValue(ctx, "redisCacheEnabled", false, of.EvaluationContext{})
	logging.From(ctx).Debug("flag",
		slog.String("name", "redisCacheEnabled"),
		slog.Bool("value", val),
		slog.Any("error", err),
	)
	return err == nil && val
}

func (f *Flags) ThrottleEnabled(ctx context.Context) bool {
	val, err := f.client.BooleanValue(ctx, "throttleEnabled", false, of.EvaluationContext{})
	logging.From(ctx).Debug("flag",
		slog.String("name", "throttleEnabled"),
		slog.Bool("value", val),
		slog.Any("error", err),
	)
	return err == nil && val
}
