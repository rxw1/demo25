package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"rxw1/ordersvc/internal/logging"
	"rxw1/ordersvc/internal/mongo"
	mynats "rxw1/ordersvc/internal/nats"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// In real code, wrap ResponseWriter to capture status/bytes.
		reqID := r.Header.Get("X-Request-ID")

		l := slog.Default().With(
			"trace_id", reqID,
			"remote_ip", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.Path,
		)

		// ctx := Into(r.Context(), l)
		ctx := logging.With(r.Context(), l)
		r = r.WithContext(ctx)

		// call next
		next.ServeHTTP(w, r)

		l.Info("request complete", "duration_ms", time.Since(start).Milliseconds())
	})
}

func main() {
	ctx := context.Background()

	// TODO
	// buildVersion := "dev" // set via -ldflags "-X main.buildVersion=..."

	// cfg := logging.Config{
	// 	Level:       getenv("LOG_LEVEL", "debug"),
	// 	JSON:        getenv("LOG_FORMAT", "json") == "json",
	// 	AddSource:   getenv("LOG_SOURCE", "true") == "true",
	// 	Service:     "ordersvc",
	// 	Version:     buildVersion,
	// 	Environment: getenv("ENV", "dev"),
	// 	SetDefault:  true, // so slog.Default() works across the app
	// }

	// logger, err := logging.New(cfg)
	// if err != nil {
	// 	panic(err)
	// }

	// logger.Info("boot", "pid", os.Getpid())

	// Mongo
	store, err := mongo.Connect(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatal(err)
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	sub, err := mynats.SubscribeToOrdersCreated(ctx, nc, store)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	sub2, err := mynats.SubscribeToOrdersRequested(ctx, nc, store)
	if err != nil {
		log.Fatal(err)
	}
	defer sub2.Unsubscribe()

	// Chi
	r := chi.NewRouter()
	r.Use(Logging)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logging.With(r.Context(), "user_id", 789234) // TODO
			logging.From(ctx).Error("db query failed", "err", err)

			// TODO
			if r.URL.Path == "/healthz" {
				w.WriteHeader(200)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	// Start server
	log.Println("ordersvc up on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

// func getenv(k, def string) string {
// 	if v := os.Getenv(k); v != "" {
// 		return v
// 	}
// 	return def
// }
