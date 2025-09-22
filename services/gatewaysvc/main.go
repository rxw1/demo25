//go:generate go run github.com/99designs/gqlgen generate
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"rxw1/flags"
	"rxw1/gatewaysvc/internal/cache"
	"rxw1/gatewaysvc/internal/graphql"
	"rxw1/logging"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	port = 8080
	name = "gatewaysvc"
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

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// Redis
	rc := cache.New(os.Getenv("REDIS_ADDR"))

	// Flags
	ff := flags.New(name)

	// GraphQL
	res := &graphql.Resolver{NC: nc, RC: rc, FF: ff}
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
					parts := strings.SplitSeq(allowedFromEnv, ",")
					for p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							allowedOrigins = append(allowedOrigins, p)
						}
					}
				} else {
					allowedOrigins = []string{
						"http://localhost:8088",
					}
				}

				// Always allow if the Origin host matches the request host (same host/port).
				if u, err := url.Parse(origin); err == nil {
					if u.Host == r.Host {
						logging.From(ctx).Warn("Allowed same-host origin", "origin", origin)
						return true
					}
				}

				if slices.Contains(allowedOrigins, origin) {
					logging.From(ctx).Warn("Allowed WS origin", "origin", origin)
					return true
				}
				logging.From(ctx).Warn("Blocked WS origin", "origin", origin)
				return false
			},
		},
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.AddTransport(transport.Options{}) // For the playground
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{}) // Must be after the WebSocket transport

	srv.Use(extension.Introspection{}) // For running gqlgen
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100), // From default config
	})

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

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		logging.From(ctx).Error("server startup failed", "port", port, "svc", name)
		os.Exit(1)
	}
	logging.From(ctx).Info("server ready", "port", port, "svc", name)
}
