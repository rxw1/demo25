package db

import (
	"context"
	"time"

	"rxw1/gatewaysvc/model"
	"rxw1/logging"

	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct{ C *mongo.Collection }

func Connect(ctx context.Context, uri string) (*Store, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Store{C: cli.Database("app").Collection("orders")}, nil
}

func (s *Store) AddOrder(ctx context.Context, eventID, productID string, qty int, createdAt time.Time) error {
	ctx = logging.With(ctx, "eventID", eventID, "productID", productID, "qty", qty, "createdAt", createdAt)

	logging.From(ctx).Debug("AddOrder")

	res, err := s.C.UpdateOne(ctx,
		bson.M{"eventId": eventID},
		bson.M{
			"$setOnInsert": bson.M{
				"id":        ulid.Make().String(),
				"eventId":   eventID,
				"productId": productID,
				"qty":       qty,
				"createdAt": createdAt,
			},
		}, options.Update().SetUpsert(true))
	logging.From(ctx).Debug("result", "res", res, "err", err)
	return err
}

func (s *Store) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	ctx = logging.With(ctx, "mongo", "GetAllOrders")
	cur, err := s.C.Find(ctx, bson.M{})
	if err != nil {
		logging.From(ctx).Error("DATABASE MONGO failed to find orders", "error", err)
		return nil, err
	}
	var orders []model.Order
	if err := cur.All(ctx, &orders); err != nil {
		logging.From(ctx).Error("DATABASE MONGO failed to decode orders", "error", err)
		return nil, err
	}
	return orders, nil
}
