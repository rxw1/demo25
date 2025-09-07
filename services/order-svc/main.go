package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"rxw1/order-svc/internal/logging"
	"rxw1/order-svc/internal/mongo"
	nsub "rxw1/order-svc/internal/nats"

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

	buildVersion := "dev" // set via -ldflags "-X main.buildVersion=..."

	cfg := logging.Config{
		Level:       getenv("LOG_LEVEL", "info"),
		JSON:        getenv("LOG_FORMAT", "json") == "json",
		AddSource:   getenv("LOG_SOURCE", "false") == "true",
		Service:     "order-svc",
		Version:     buildVersion,
		Environment: getenv("ENV", "dev"),
		SetDefault:  true, // so slog.Default() works across the app
	}

	logger, err := logging.New(cfg)
	if err != nil {
		panic(err)
	}

	logger.Info("boot", "pid", os.Getpid())

	// Later in handlers:
	// func h(w http.ResponseWriter, r *http.Request) {
	// ctx := logging.With(r.Context(), "user_id", uid)
	// logging.From(ctx).Error("db query failed", "err", err)
	// }

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

	// NATS Subscriber
	if err := nsub.Start(ctx, nc, store); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Use(Logging)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logging.With(r.Context(), "user_id", 789234)
			logging.From(ctx).Error("db query failed", "err", err)

			// Simple health check for demo; use real one in prod.
			if r.URL.Path == "/healthz" {
				w.WriteHeader(200)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	// r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	// r.Get("/livez", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	// r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
	// 	// FIXME

	// 	ctx := logging.With(r.Context(), "productId", 789234)
	// 	// logging.From(ctx).Error("create order", "err", err)

	// 	res, err := store.CreateOrder(r.Context())
	// 	if err != nil {
	// 		logging.From(r.Context()).Error("create order failed", "err", err)
	// 		w.WriteHeader(500)
	// 		return
	// 	}

	// 	logging.From(ctx).Info("order created", "res", res)

	// 	w.WriteHeader(201)
	// })

	r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
		orders, err := store.GetAllOrders(r.Context())
		if err != nil {
			logging.From(r.Context()).Error("get orders failed", "err", err)
			w.WriteHeader(500)
			return
		}

		logging.From(r.Context()).Info("orders fetched", "count", len(orders))

		if err := json.NewEncoder(w).Encode(orders); err != nil {
			logging.From(r.Context()).Error("encode orders failed", "err", err)
			w.WriteHeader(500)
			return
		}
	})

	// Start server
	log.Println("order-svc up on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
