//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	"rxw1/product-svc/graph"
	"rxw1/product-svc/internal/cache"
	"rxw1/product-svc/internal/db"
	"rxw1/product-svc/internal/flags"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	ctx := context.Background()

	// Postgres
	pg, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Migrations
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.Migrate(ctx, os.Getenv("DATABASE_URL"), migrationsFS); err != nil {
			log.Fatal("migrate failed: ", err)
		}
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Redis
	rc := cache.New(os.Getenv("REDIS_ADDR"))
	_ = redis.NewClient // keep import
	ff := flags.New()

	// GraphQL
	res := &graph.Resolver{PG: pg, NC: nc, RC: rc, FF: ff}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: res}))

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

	r.Handle("/", playground.Handler("GraphQL", "/graphql"))
	r.Handle("/graphql", srv)

	log.Println("product-svc up on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
