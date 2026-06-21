package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateUser(ctx context.Context, conn *pgx.Conn, email, hashedPassword string, method string) error {
	_, err := conn.Exec(ctx, "insert into users (email, hashed_password,registration_method) values ($1, $2,$3);", email, hashedPassword, method)
	if err != nil {

		return err
	}
	return nil
}
