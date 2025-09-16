package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rxw1/ordersvc/internal/logging"
	"rxw1/ordersvc/internal/mongo"

	"github.com/nats-io/nats.go"
)

type Event struct {
	ID, ProductID string
	Qty           int
	CreatedAt     string
}

func SubscribeToOrdersCreated(ctx context.Context, nc *nats.Conn, store *mongo.Store) (*nats.Subscription, error) {
	ctx2 := logging.With(ctx, "nats", "Start")
	sub, err := nc.Subscribe("orders.created", func(m *nats.Msg) {
		var e Event
		if json.Unmarshal(m.Data, &e) != nil {
			logging.From(ctx2).Error("failed to unmarshal event", "data", string(m.Data))
			return
		}
		ts, _ := time.Parse(time.RFC3339, e.CreatedAt)
		_ = store.AddOrder(ctx, e.ID, e.ProductID, e.Qty, ts)
	})
	return sub, err
}

func SubscribeToOrdersRequested(ctx context.Context, nc *nats.Conn, store *mongo.Store) (*nats.Subscription, error) {
	ctx2 := logging.With(ctx, "nats", "Start")
	sub, err := nc.Subscribe("orders.all", func(m *nats.Msg) {
		res, err := store.GetAllOrders(ctx)
		if err != nil {
			logging.From(ctx2).Error("failed to get all orders", "error", err)
			return
		}

		b, err := json.Marshal(res)
		if err != nil {
			logging.From(ctx2).Error("failed to marshal orders", "error", err)
			return
		}

		fmt.Printf("responding to orders.all: %s\n", string(b))

		if err := m.Respond(b); err != nil {
			logging.From(ctx2).Error("failed to respond to orders.all", "error", err)
			return
		}
	})
	return sub, err
}
