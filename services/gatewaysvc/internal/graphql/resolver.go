package graphql

import (
	"rxw1/gatewaysvc/internal/cache"
	"rxw1/gatewaysvc/internal/flags"

	"github.com/nats-io/nats.go"
)

// This file will not be regenerated automatically. It serves as dependency
// injection for your app, add any dependencies you require here.

type Resolver struct {
	NC *nats.Conn
	RC *cache.Cache
	FF *flags.Flags
}
