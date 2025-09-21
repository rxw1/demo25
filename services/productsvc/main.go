//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"embed"
	"log"
	"os"
	"time"

	"rxw1/logging"
	"rxw1/productsvc/product"

	"github.com/nats-io/nats.go"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger, err := logging.NewTint(logging.Config{
		Level:       getenv("LOG_LEVEL", "debug"),
		JSON:        getenv("LOG_FORMAT", "json") == "json",
		AddSource:   getenv("LOG_SOURCE", "true") == "true",
		Service:     "productsvc",
		Version:     getenv("BUILD_VERSION", "dev"),
		Environment: getenv("ENV", "dev"),
		SetDefault:  true,
		TimeFormat:  time.Kitchen,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := logging.Into(context.Background(), logger)

	logger.Info("boot", "pid", os.Getpid())

	// Postgres
	pg, err := product.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Pool.Close()

	// Migrations
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := product.Migrate(ctx, os.Getenv("DATABASE_URL"), migrationsFS); err != nil {
			log.Fatal("migrate failed: ", err)
		}
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
