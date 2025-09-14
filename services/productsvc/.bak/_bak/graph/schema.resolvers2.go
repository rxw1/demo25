package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rxw1/product-svc/graph/model"
	"rxw1/product-svc/internal/cache"
	"rxw1/product-svc/internal/db"
	"rxw1/product-svc/internal/flags"
	"rxw1/product-svc/internal/logging"

	"github.com/nats-io/nats.go"
)

type Resolver struct {
	PG *db.PG
	NC *nats.Conn
	RC *cache.Cache
	FF *flags.Flags
}

func (r *Resolver) ProductByID(ctx context.Context, id string) (*model.Product, error) {
	ctx2 := logging.With(ctx, "productID", id)
	logging.From(ctx2).Info("fetch product by ID")

	// Check Redis cache
	useCache := r.FF.RedisEnabled(ctx)
	if useCache {
		if s, err := r.RC.Get(ctx, "product:"+id); err == nil {
			var p model.Product
			if json.Unmarshal([]byte(s), &p) == nil {
				return &p, nil
			}
		}
	}

	// Fetch from Postgres
	pid, name, price, err := r.PG.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	p := &model.Product{
		ID:    pid,
		Name:  name,
		Price: int32(price),
	}

	// Store in Redis cache
	if useCache {
		b, _ := json.Marshal(p)
		_ = r.RC.Set(ctx, "product:"+id, string(b), 5*time.Minute)
	}

	return p, nil
}

func (r *Resolver) CreateOrder(ctx context.Context, productID string, qty int) (*model.Order, error) {
	ctx2 := logging.With(ctx, "productID", productID, "qty", qty)
	logging.From(ctx2).Info("create order")

	// TODO Validate input
	// In real life, you'd want to check productID exists, stock levels, etc.

	// Publish event to NATS
	event := map[string]any{
		"id":        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"productID": productID,
		"qty":       qty,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}
	logging.From(ctx2).Info("publishing event", "event", event)

	b, err := json.Marshal(event)
	if err != nil {
		logging.From(ctx2).Error("failed to marshal event", "error", err)
		return nil, err
	}

	if err := r.NC.Publish("orders.created", b); err != nil {
		logging.From(ctx2).Error("failed to publish event", "error", err)
		return nil, err
	}

	// TODO
	// Flush Redis cache for this product (to update stock levels, etc.)
	// if r.FF.RedisEnabled(ctx) {
	// 	_ = r.RC.Del(ctx, "product:"+productID)
	// }

	// TODO
	// Return order (in real life, you'd want to store this in Postgres)
	order := &model.Order{
		ID:        event["id"].(string),
		ProductID: productID,
		Qty:       int32(qty),
		CreatedAt: event["createdAt"].(string),
	}

	logging.From(ctx2).Info("order created", "order", order)
	return order, nil
}
