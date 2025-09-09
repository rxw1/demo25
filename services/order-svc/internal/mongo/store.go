package mongo

import (
	"context"
	"time"

	"rxw1/order-svc/internal/logging"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct{ C *mongo.Collection }

func Connect(ctx context.Context, uri string) (*Store, error) {
	ctx2 := logging.With(ctx, "mongo", "Connect")
	logging.From(ctx2).Debug("connecting to mongo", "uri", uri)
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	logging.From(ctx2).Debug("mongo connected")
	return &Store{C: cli.Database("app").Collection("orders")}, nil
}

func (s *Store) UpsertOrder(ctx context.Context, evtID, productID string, qty int, createdAt time.Time) error {
	ctx2 := logging.With(ctx, "mongo", "UpsertOrder")
	logging.From(ctx2).Debug("upserting order", "eventId", evtID, "productId", productID, "qty", qty, "createdAt", createdAt)
	res, err := s.C.UpdateOne(ctx,
		bson.M{"eventId": evtID},
		bson.M{
			"$setOnInsert": bson.M{
				"eventId":   evtID,
				"productId": productID,
				"qty":       qty,
				"createdAt": createdAt,
			},
		}, options.Update().SetUpsert(true))
	logging.From(ctx2).Debug("upsert result", "result", res, "error", err)
	return err
}

func (s *Store) GetAllOrders(ctx context.Context) ([]bson.M, error) {
	ctx2 := logging.With(ctx, "mongo", "GetAllOrders")
	logging.From(ctx2).Debug("getting all orders")
	cur, err := s.C.Find(ctx, bson.M{})
	if err != nil {
		logging.From(ctx2).Error("failed to find orders", "error", err)
		return nil, err
	}
	var orders []bson.M
	if err := cur.All(ctx, &orders); err != nil {
		logging.From(ctx2).Error("failed to decode orders", "error", err)
		return nil, err
	}
	logging.From(ctx2).Debug("fetched orders", "count", len(orders))
	return orders, nil
}
