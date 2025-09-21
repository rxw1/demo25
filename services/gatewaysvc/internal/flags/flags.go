package flags

import (
	"context"

	"rxw1/logging"

	of "github.com/open-feature/go-sdk/openfeature"
)

type Flags struct {
	client *of.Client
}

func New() *Flags {
	return &Flags{
		client: of.NewClient("productsvc"),
	}
}

func (f *Flags) RedisEnabled(ctx context.Context) bool {
	ctx2 := logging.With(ctx, "flag", "redisCacheEnabled")
	val, err := f.client.BooleanValue(ctx, "redisCacheEnabled", false, of.EvaluationContext{})
	logging.From(ctx2).Debug("flag redisCacheEnabled", "value", val, "error", err)
	return err == nil && val
}
