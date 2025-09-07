package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct{ C *mongo.Collection }

func Connect(ctx context.Context, uri string) (*Store, error) {
	fmt.Printf("connecting to mongo %s\n", uri)
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	fmt.Printf("mongo connected\n")
	return &Store{C: cli.Database("app").Collection("orders")}, nil
}

func (s *Store) UpsertOrder(ctx context.Context, evtID, productID string, qty int, createdAt time.Time) error {
	fmt.Printf("upsert order event %s product %s qty %d at %s\n", evtID, productID, qty, createdAt)
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

	fmt.Printf("upsert result: %+v, %v\n", res, err)
	return err
}

func (s *Store) GetAllOrders(ctx context.Context) ([]bson.M, error) {
	cur, err := s.C.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var orders []bson.M
	if err := cur.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
