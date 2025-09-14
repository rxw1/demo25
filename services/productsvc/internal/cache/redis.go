package cache

import (
	"context"
	"time"

	"rxw1/productsvc/internal/logging"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	R *redis.Client
}

// addr example: "localhost:6379"
func New(addr string) *Cache {
	return &Cache{
		R: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// returns redis.Nil error if key does not exist
func (c *Cache) Get(ctx context.Context, k string) (string, error) {
	ctx2 := logging.With(ctx, "k", k)
	logging.From(ctx2).Debug("cache get")
	r, err := c.R.Get(ctx, k).Result()
	logging.From(ctx2).Debug("cache get result", "result", r, "error", err)
	return r, err
}

// ttl 0 means no expiration
func (c *Cache) Set(ctx context.Context, k, v string, ttl time.Duration) error {
	ctx2 := logging.With(ctx, "k", k, "ttl", ttl)
	logging.From(ctx2).Debug("cache set", "value", v)
	return c.R.Set(ctx, k, v, ttl).Err()
}
