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

	"rxw1/logging"
	"rxw1/ordersvc/internal/flags"
	"rxw1/ordersvc/internal/graphql"
	"rxw1/ordersvc/internal/mongo"
	mynats "rxw1/ordersvc/internal/nats"

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

	// Flags
	ff := flags.New()

	// MongoDB
	mo, err := mongo.Connect(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatal(err)
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// Subscribers
	sub, err := mynats.SubscribeToOrdersCreated(ctx, nc, mo, ff)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	sub2, err := mynats.SubscribeToOrdersRequested(ctx, nc, mo, ff)
	if err != nil {
		log.Fatal(err)
	}
	defer sub2.Unsubscribe()

	res := &graphql.Resolver{MO: mo, NC: nc, FF: ff}
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

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.AddTransport(transport.Options{}) // For the playground
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{}) // Must be after the WebSocket transport

	srv.Use(extension.Introspection{}) // For running gqlgen
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

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

	r.Handle("/", playground.Handler("GraphQL", "/graphql"))
	r.Handle("/graphql", srv)

	// Start server
	log.Println("ordersvc up on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
