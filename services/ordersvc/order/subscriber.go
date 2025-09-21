package order

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"time"

	"rxw1/logging"

	"github.com/nats-io/nats.go"
)

type Event struct {
	ID        string
	ProductID string
	CreatedAt string
	Qty       int
}

func SubscribeToOrdersCreated(ctx context.Context, nc *nats.Conn, mo *Store, ff *Flags) (*nats.Subscription, error) {
	// ctx = logging.With(ctx, "fn", "SubscribeToOrdersCreated", "package", "nats")

	sub, err := nc.Subscribe("order.created", func(m *nats.Msg) {
		// ctx = logging.With(ctx, "fn", "Subscribe", "package", "nats")
		var e Event
		if json.Unmarshal(m.Data, &e) != nil {
			logging.From(ctx).Error("failed to unmarshal event", "data", string(m.Data))
			return // return nothing = skip message
		}

		ts, err := time.Parse(time.RFC3339, e.CreatedAt)
		if err != nil {
			logging.From(ctx).Error("failed to parse time", "error", err, "time", e.CreatedAt)
			return
		}

		logging.From(ctx).Info("event", "eventId", e.ID, "productId", e.ProductID, "qty", e.Qty, "createdAt", e.CreatedAt)

		if ff.ThrottleEnabled(ctx) {
			t := time.Duration(rand.IntN(500)) * time.Millisecond
			logging.From(ctx).Info("throttling enabled, sleeping", "t", t)
			time.Sleep(t)
		}

		err = mo.AddOrder(ctx, e.ID, e.ProductID, e.Qty, ts)
		if err != nil {
			logging.From(ctx).Error("failed to add order to mongodb", "error", err)
			return
		}

		logging.From(ctx).Info("order created", "event", e, "error", err)
	})
	return sub, err
}

func SubscribeToOrdersRequested(ctx context.Context, nc *nats.Conn, mo *Store, ff *Flags) (*nats.Subscription, error) {
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
