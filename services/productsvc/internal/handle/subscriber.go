package handle

import (
	"context"
	"encoding/json"

	"rxw1/logging"
	"rxw1/productsvc/internal/db"

	"github.com/nats-io/nats.go"
)

func AllProducts(ctx context.Context, nc *nats.Conn, db *db.PG) (*nats.Subscription, error) {
	ctx = logging.With(ctx, "fn", "AllProducts", "pkg", "NATS")
	sub, err := nc.Subscribe("products.all", func(m *nats.Msg) {
		res, err := db.GetProducts(ctx)
		if err != nil {
			logging.From(ctx).Error("failed to get all products", "error", err)
			return
		}

		b, err := json.Marshal(res)
		if err != nil {
			logging.From(ctx).Error("failed to marshal products", "error", err)
			return
		}

		logging.From(ctx).Info("responding to products.all", "count", len(res))

		if err := m.Respond(b); err != nil {
			logging.From(ctx).Error("failed to respond to products.all", "error", err)
			return
		}
	})
	return sub, err
}
