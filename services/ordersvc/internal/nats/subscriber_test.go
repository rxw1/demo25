package nats_test

import (
	"context"
	"testing"

	"rxw1/ordersvc/internal/mongo"
	mynats "rxw1/ordersvc/internal/nats"

	"github.com/nats-io/nats.go"
)

func TestSubscribeToOrdersCreated(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nc      *nats.Conn
		store   *mongo.Store
		want    *nats.Subscription
		wantErr bool
	}{
		{
			name:    "nil nc and store",
			nc:      nil,
			store:   nil,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := mynats.SubscribeToOrdersCreated(context.Background(), tt.nc, tt.store)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SubscribeToOrdersCreated() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("SubscribeToOrdersCreated() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("SubscribeToOrdersCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscribeToOrdersRequested(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nc      *nats.Conn
		store   *mongo.Store
		want    *nats.Subscription
		wantErr bool
	}{
		{
			name:    "nil nc and store",
			nc:      nil,
			store:   nil,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := mynats.SubscribeToOrdersRequested(context.Background(), tt.nc, tt.store)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SubscribeToOrdersRequested() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("SubscribeToOrdersRequested() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("SubscribeToOrdersRequested() = %v, want %v", got, tt.want)
			}
		})
	}
}
