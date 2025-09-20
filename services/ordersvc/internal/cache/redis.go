package cache

import (
	"context"
	"time"

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
	r, err := c.R.Get(ctx, k).Result()
	return r, err
}

// ttl 0 means no expiration
func (c *Cache) Set(ctx context.Context, k, v string, ttl time.Duration) error {
	return c.R.Set(ctx, k, v, ttl).Err()
}
