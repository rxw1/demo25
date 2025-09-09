package seed

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	_, err = pool.Exec(ctx, `insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Seed done")
}
