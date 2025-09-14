package graph

import (
	"rxw1/productsvc/internal/cache"
	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/flags"

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
