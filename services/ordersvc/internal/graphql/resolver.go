package graphql

import (
	"rxw1/ordersvc/internal/cache"
	"rxw1/ordersvc/internal/flags"
	"rxw1/ordersvc/internal/mongo"

	"github.com/nats-io/nats.go"
)

// This file will not be regenerated automatically. It serves as dependency
// injection for your app, add any dependencies you require here.

type Resolver struct {
	MO *mongo.Store
	NC *nats.Conn
	RC *cache.Cache
	FF *flags.Flags
}
