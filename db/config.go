package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Connect(db string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), db)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	fmt.Print("Database connected!")
	return conn, nil
}
