package flags

import (
	"context"
	"fmt"

	of "github.com/open-feature/go-sdk/openfeature"
)

type Flags struct {
	client *of.Client
}

func New() *Flags {
	return &Flags{
		client: of.NewClient("product-svc"),
	}
}

func (f *Flags) RedisEnabled(ctx context.Context) bool {
	val, err := f.client.BooleanValue(ctx, "redisCacheEnabled", false, of.EvaluationContext{})
	fmt.Println("flag redisCacheEnabled", val, err)
	return err == nil && val
}
