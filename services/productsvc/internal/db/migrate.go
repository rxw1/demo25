// https://github.com/golang-migrate/migrate

package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"rxw1/productsvc/internal/logging"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

// Migrate runs all up migrations from the embedded migrations directory.
// Call this before creating your pgx pool (see: services/productsvc/main.go).
func Migrate(ctx context.Context, databaseURL string, migrationsFS embed.FS) error {
	ctx2 := logging.With(ctx)
	logging.From(ctx2).Info("migrate", "databaseURL", databaseURL)

	// open a database/sql DB (required by golang-migrate database driver)
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()

	drv, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", drv)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
