package graphql

import (
	"rxw1/productsvc/internal/cache"
	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/flags"
	"rxw1/productsvc/internal/natsx"
)

// This file will not be regenerated automatically. It serves as dependency
// injection for your app, add any dependencies you require here.

type Resolver struct {
	PG *db.PG
	NC natsx.Client
	RC *cache.Cache
	FF *flags.Flags
}
