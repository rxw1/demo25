package graph

import (
	"rxw1/product-svc/internal/cache"
	"rxw1/product-svc/internal/db"
	"rxw1/product-svc/internal/flags"

	"github.com/nats-io/nats.go"
)

// This file will not be regenerated automatically. It serves as dependency
// injection for your app, add any dependencies you require here.

type Resolver struct {
	PG *db.PG
	NC *nats.Conn
	RC *cache.Cache
	FF *flags.Flags
}
