package nats

import (
	"context"
	"encoding/json"
	"time"

	"rxw1/order-svc/internal/logging"
	"rxw1/order-svc/internal/mongo"

	"github.com/nats-io/nats.go"
)

type Event struct {
	ID, ProductID string
	Qty           int
	CreatedAt     string
}

func Start(ctx context.Context, nc *nats.Conn, store *mongo.Store) error {
	ctx2 := logging.With(ctx, "nats", "Start")
	logging.From(ctx2).Debug("starting nats subscriber")
	subscription, err := nc.Subscribe("orders.created", func(m *nats.Msg) {
		var e Event
		if json.Unmarshal(m.Data, &e) != nil {
			logging.From(ctx2).Error("failed to unmarshal event", "data", string(m.Data))
			return
		}
		ts, _ := time.Parse(time.RFC3339, e.CreatedAt)
		_ = store.UpsertOrder(ctx, e.ID, e.ProductID, e.Qty, ts)
	})
	logging.From(ctx2).Debug("nats subscriber started", "subscription", subscription, "error", err)
	return err
}
