package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"rxw1/logging"
	"rxw1/ordersvc/internal/flags"
	"rxw1/ordersvc/internal/mongo"

	"github.com/nats-io/nats.go"
)

type Event struct {
	ID        string
	ProductID string
	CreatedAt string
	Qty       int
}

type Order struct {
	ID        string
	EventID   string
	ProductID string
	Qty       int32
	CreatedAt string
}

func SubscribeToOrdersCreated(ctx context.Context, nc *nats.Conn, mo *mongo.Store, ff *flags.Flags) (*nats.Subscription, error) {
	ctx = logging.With(ctx, "fn", "SubscribeToOrdersCreated", "pkg", "NATS")
	sub, err := nc.Subscribe("order.created", func(m *nats.Msg) {
		var e Event
		if json.Unmarshal(m.Data, &e) != nil {
			logging.From(ctx).Error("EVENT failed to unmarshal event", "data", string(m.Data))
			return
		}

		ts, err := time.Parse(time.RFC3339, e.CreatedAt)
		if err != nil {
			logging.From(ctx).Error("EVENT failed to parse time", "error", err, "createdAt", e.CreatedAt)
			return
		}

		logging.From(ctx).Info("EVENT received order created", "eventId", e.ID, "productId", e.ProductID, "qty", e.Qty, "createdAt", e.CreatedAt)

		if ff.ThrottleEnabled(ctx) {
			logging.From(ctx).Info("throttling enabled, sleeping")
			time.Sleep(time.Duration(rand.IntN(500)) * time.Millisecond)
		}

		err = mo.AddOrder(ctx, e.ID, e.ProductID, e.Qty, ts)
		if err != nil {
			logging.From(ctx).Error("EVENT failed to add order", "error", err)
			return
		}

		fmt.Printf("order created: %+v\n", e)
	})
	return sub, err
}

func SubscribeToOrdersRequested(ctx context.Context, nc *nats.Conn, mo *mongo.Store, ff *flags.Flags) (*nats.Subscription, error) {
	ctx = logging.With(ctx, "fn", "SubscribeToOrdersRequested", "pkg", "NATS")
	sub, err := nc.Subscribe("orders.all", func(m *nats.Msg) {
		res, err := mo.GetAllOrders(ctx)
		if err != nil {
			logging.From(ctx).Error("failed to get all orders", "error", err)
			return
		}

		b, err := json.Marshal(res)
		if err != nil {
			logging.From(ctx).Error("failed to marshal orders", "error", err)
			return
		}

		logging.From(ctx).Info("responding to orders.all", "count", len(res))

		if err := m.Respond(b); err != nil {
			logging.From(ctx).Error("failed to respond to orders.all", "error", err)
			return
		}
	})
	return sub, err
}
