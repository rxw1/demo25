//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"time"

	"rxw1/logging"
	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/handle"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	pg, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("connect pg", "error", err)
		panic(err)
	}
	defer pg.Pool.Close()

	// Migrations
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.Migrate(ctx, os.Getenv("DATABASE_URL"), migrationsFS); err != nil {
			logger.Error("migrate", "error", err)
			panic(err)
		}
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Error("nats connect", "error", err)
		panic(err)
	}
	defer nc.Drain()

	// Subscribers
	sub, err := handle.AllProducts(ctx, nc, pg)
	if err != nil {
		logging.From(ctx).Error("", "error", err.Error())
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

	port := getenv("PORT", ":8081")

	// Start server
	err = http.ListenAndServe(port, r)
	if err != nil {
		logging.From(ctx).Error("productsvc failed to start up", "port", port)
		os.Exit(1)
	}
	logging.From(ctx).Info("productsvc up", "port", port)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
