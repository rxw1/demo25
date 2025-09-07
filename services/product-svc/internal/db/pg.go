package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct{ Pool *pgxpool.Pool }

func Connect(ctx context.Context, url string) (*PG, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	fmt.Println("postgres connected")
	return &PG{Pool: pool}, nil
}

func (p *PG) GetProduct(ctx context.Context, id string) (string, string, int, error) {
	row := p.Pool.QueryRow(ctx, `select id, name, price from products where id=$1`, id)
	var pid, name string
	var price int
	fmt.Printf("db get product: %s, %s, %d\n", pid, name, price)
	return pid, name, price, row.Scan(&pid, &name, &price)
}
