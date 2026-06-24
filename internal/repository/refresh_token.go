package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var NoTokenFound = errors.New("No token found")

type TokenDetails struct {
	ID         string
	UserID     string
	Value      string
	RevokedAt  *time.Time
	ExpiryDate time.Time
}

func StoreRefreshToken(ctx context.Context, conn *pgx.Conn, token, user_id string, expiryDate time.Time) error {
	if _, err := conn.Exec(ctx, "insert into refresh_tokens (user_id,value,expiry_date) values ($1,$2,$3)", user_id, token, expiryDate); err != nil {
		return err
	}
	return nil
}
func GetRefreshToken(ctx context.Context, conn *pgx.Conn, tokenStr string) (*TokenDetails, error) {
	var token TokenDetails
	rows := conn.QueryRow(ctx, "SELECT id, user_id, value, revoked_at, expiry_date FROM refresh_tokens WHERE value = $1", tokenStr)
	err := rows.Scan(&token.ID, &token.UserID, &token.Value, &token.RevokedAt, &token.ExpiryDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, NoTokenFound
		}
		return nil, err
	}
	return &token, nil
}
func RevokeRefreshToken(ctx context.Context, conn *pgx.Conn, token string) error {
	_, err := conn.Exec(ctx, "UPDATE refresh_tokens set revoked_at = $1 where value = $2", time.Now(), token)
	if err != nil {
		return err
	}
	return nil
}

// revoking all tokens belonging to a user
func RevokeRefreshTokensByUser(ctx context.Context, conn *pgx.Conn, userID string) error {
	_, err := conn.Exec(ctx, "UPDATE refresh_tokens set revoked_at = $1 where user_id = $2", time.Now(), userID)
	if err != nil {
		return err
	}
	return nil
}
