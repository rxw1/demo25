//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"rxw1/logging"
	"rxw1/productsvc/internal/cache"
	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/flags"
	"rxw1/productsvc/internal/graphql"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger, err := logging.NewTint(logging.Config{
		Level:       getenv("LOG_LEVEL", "debug"),
		JSON:        getenv("LOG_FORMAT", "json") == "false",
		AddSource:   getenv("LOG_SOURCE", "true") == "false",
		Service:     "productsvc",
		Version:     buildVersion,
		Environment: getenv("ENV", "dev"),
		SetDefault:  true,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := logging.Into(context.Background(), logger)

	logger.Info("boot", "pid", os.Getpid())

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

	// Jetstream
	_, err = jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	// Redis
	rc := cache.New(os.Getenv("REDIS_ADDR"))
	_ = redis.NewClient // keep import
	ff := flags.New()

	// GraphQL
	res := &graphql.Resolver{PG: pg, NC: nc, RC: rc, FF: ff}
	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: res}))

	// Websockets
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" { // no origin header, likely same-origin or non-browser
					return true
				}

				// Allowed origins from env (comma-separated). Defaults cover local dev.
				allowedFromEnv := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS"))
				var allowedOrigins []string
				if allowedFromEnv != "" {
					parts := strings.Split(allowedFromEnv, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							allowedOrigins = append(allowedOrigins, p)
						}
					}
				} else {
					allowedOrigins = []string{
						"http://localhost:3000",
						"http://localhost:3001",
					}
				}

				// Always allow if the Origin host matches the request host (same host/port).
				if u, err := url.Parse(origin); err == nil {
					if u.Host == r.Host {
						fmt.Printf("Allowed same-host origin: %s\n", origin)
						return true
					}
				}

				if slices.Contains(allowedOrigins, origin) {
					fmt.Printf("Allowed origin: %s\n", origin)
					return true
				}
				fmt.Printf("Blocked origin: %s (set WS_ALLOWED_ORIGINS to override)\n", origin)
				return false
			},
		},
	})

	srv.AddTransport(transport.Options{}) // For the playground
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{}) // Must be after the WebSocket transport
	srv.Use(extension.Introspection{}) // For running gqlgen

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

	log.Println("productsvc up on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
