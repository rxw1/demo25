package mongo_test

import (
	"context"
	"testing"

	"rxw1/ordersvc/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

type Order struct {
	ID        string `bson:"id"`
	EventID   string `bson:"eventId"`
	ProductID string `bson:"productId"`
	Qty       int    `bson:"qty"`
	CreatedAt string `bson:"createdAt"`
}

func TestStore_GetAllOrders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		uri     string
		want    []bson.M
		wantErr bool
	}{
		{
			name:    "valid uri",
			uri:     "mongodb://localhost:27017",
			want:    []Order, // TODO: fill in expected result
			wantErr: false,
		},
		{
			name:    "invalid uri",
			uri:     "invalid_uri",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := mongo.Connect(context.Background(), tt.uri)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := s.GetAllOrders(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetAllOrders() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetAllOrders() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetAllOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}
