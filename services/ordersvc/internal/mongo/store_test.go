package mongo_test

import (
	"context"
	"testing"
	"time"

	"rxw1/ordersvc/internal/mongo"
)

type Order struct {
	ID        string `bson:"id"`
	EventID   string `bson:"eventId"`
	ProductID string `bson:"productId"`
	Qty       int    `bson:"qty"`
	CreatedAt string `bson:"createdAt"`
}

func TestStore_AddOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		uri string
		// Named input parameters for target function.
		eventID   string
		productID string
		qty       int
		createdAt time.Time
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := mongo.Connect(context.Background(), tt.uri)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := s.AddOrder(context.Background(), tt.eventID, tt.productID, tt.qty, tt.createdAt)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddOrder() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddOrder() succeeded unexpectedly")
			}
		})
	}
}

func TestStore_GetAllOrders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		uri     string
		want    []mongo.Order
		wantErr bool
	}{
		// TODO: Add test cases.
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
