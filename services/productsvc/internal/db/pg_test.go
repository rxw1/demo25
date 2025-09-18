package db_test

import (
	"context"
	"testing"

	"rxw1/productsvc/internal/db"
	"rxw1/productsvc/internal/model"
)

func TestPG_GetProduct(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		url string
		// Named input parameters for target function.
		id      string
		want    *model.Product
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := db.Connect(context.Background(), tt.url)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := p.GetProduct(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProduct() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProduct() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPG_GetProducts(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		url     string
		want    []*model.Product
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := db.Connect(context.Background(), tt.url)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := p.GetProducts(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProducts() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProducts() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}
