package order

import (
	"context"

	of "github.com/open-feature/go-sdk/openfeature"
)

type Flags struct {
	client *of.Client
}

func New() *Flags {
	return &Flags{
		client: of.NewClient("ordersvc"),
	}
}

func (f *Flags) ThrottleEnabled(ctx context.Context) bool {
	val, err := f.client.BooleanValue(ctx, "throttleEnabled", false, of.EvaluationContext{})
	return err == nil && val
}
