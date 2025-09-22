/// PostgreSQL database access using pgx
/// Migrations using golang-migrate with embedded migrations
/// Migrations are run automatically if AUTO_MIGRATE env is "true"

package db

import (
	"context"

	"rxw1/gatewaysvc/model"
	"rxw1/logging"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, url string) (*PG, error) {
	logging.From(ctx).Info("pg connect", "url", url)

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		logging.From(ctx).Error("pg connect", "err", err)
		return nil, err
	}

	logging.From(ctx).Info("pg connect", "status", "success")
	return &PG{Pool: pool}, nil
}

func (p *PG) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	logging.From(ctx).Info("pg get product", "id", id)
	row := p.Pool.QueryRow(ctx, `select id, name, price from products where id=$1`, id)

	var pid, name string
	var price int
	if err := row.Scan(&pid, &name, &price); err != nil {
		if err == pgx.ErrNoRows {
			logging.From(ctx).Warn("pg get product", "id", id, "status", "no rows")
			return nil, nil // return nil if not found
		}

		logging.From(ctx).Error("pg get product", "id", id, "err", err)
		return nil, err
	}

	logging.From(ctx).Info("pg get product", "id", pid, "name", name, "price", price)
	return &model.Product{
		ID:    pid,
		Name:  name,
		Price: int32(price),
	}, nil
}

func (p *PG) GetProducts(ctx context.Context) ([]*model.Product, error) {
	logging.From(ctx).Info("pg get products")

	rows, err := p.Pool.Query(ctx, `select id, name, price from products`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*model.Product, 0, 16) // preallocate

	for rows.Next() {
		product := &model.Product{}
		var price int
		if err := rows.Scan(&product.ID, &product.Name, &price); err != nil {
			logging.From(ctx).Error("pg get products", "err", err)
			return nil, err
		}
		product.Price = int32(price)
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	logging.From(ctx).Info("pg get products", "count", len(products))
	return products, nil
}
