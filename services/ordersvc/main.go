package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"rxw1/flags"
	"rxw1/logging"
	"rxw1/ordersvc/internal/db"
	"rxw1/ordersvc/internal/handle"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/nats-io/nats.go"
)

const (
	port = 8082
	name = "ordersvc"
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

	// Flags
	ff := flags.New("ordersvc")

	// MongoDB
	mo, err := db.Connect(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		logging.From(ctx).Error("", "error", err.Error())
		os.Exit(1)
	}

	// NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logging.From(ctx).Error("", "error", err.Error())
		os.Exit(1)
	}
	defer nc.Drain()

	// Subscribers
	sub, err := handle.SubscribeToOrdersCreated(ctx, nc, mo, ff)
	if err != nil {
		logging.From(ctx).Error("", "error", err.Error())
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	sub2, err := handle.SubscribeToOrdersRequested(ctx, nc, mo, ff)
	if err != nil {
		logging.From(ctx).Error("", "error", err.Error())
		os.Exit(1)
	}
	defer sub2.Unsubscribe()

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
