package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func StoreRefreshToken(ctx context.Context, conn *pgx.Conn, token, user_id string, expiryDate time.Time) error {
	if _, err := conn.Exec(ctx, "insert into refresh_tokens (user_id,value,expiry_date) values ($1,$2,$3)", user_id, token, expiryDate); err != nil {
		return err
	}
	return nil
}
