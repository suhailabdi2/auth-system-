package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(db string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), db)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	fmt.Print("Database connected!")
	return pool, nil
}
