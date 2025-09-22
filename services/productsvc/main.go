//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"rxw1/logging"
	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/handle"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/nats-io/nats.go"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	port = 8080
	name = "productsvc"
)

func main() {
	logger, err := logging.NewTint(logging.Config{
		Service: name,
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := logging.Into(context.Background(), logger)
	logging.From(ctx).Info("boot", "pid", os.Getpid())

	// Postgres
	pg, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logging.From(ctx).Error("connect pg", "error", err)
		os.Exit(1)
	}
	defer pg.Pool.Close()

	// Migrations
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.Migrate(ctx, os.Getenv("DATABASE_URL"), migrationsFS); err != nil {
			logging.From(ctx).Error("postgres migration failed", "error", err)
			os.Exit(1)
		}
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logging.From(ctx).Error("nats connection failed", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	// Subscribers
	sub, err := handle.AllProducts(ctx, nc, pg)
	if err != nil {
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	// Chi
	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Start server
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		logging.From(ctx).Error("server startup failed", "port", port, "svc", name)
		os.Exit(1)
	}
	logging.From(ctx).Info("server ready", "port", port, "svc", name)
}
