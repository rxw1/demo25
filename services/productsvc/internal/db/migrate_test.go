package db_test

import (
	"context"
	"embed"
	"testing"

	"rxw1/productsvc/internal/db"
)

func TestMigrate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		databaseURL  string
		migrationsFS embed.FS
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := db.Migrate(context.Background(), tt.databaseURL, tt.migrationsFS)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Migrate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Migrate() succeeded unexpectedly")
			}
		})
	}
}
