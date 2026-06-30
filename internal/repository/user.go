package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("Email already exists")
var ErrEmailDoesnTExist = errors.New("no user under that email")
var UserDoesntExist = errors.New("User not found")

type UserDetails struct {
	UserID             string
	Email              string
	HashedPassword     string
	IsActive           bool
	VerificationStatus bool
}
type UserResponse struct {
	UserID             string `json:"user_id"`
	Email              string `json:"email"`
	IsActive           bool   `json:"is_active"`
	VerificationStatus bool   `json:"verification_status"`
}

func CreateUser(ctx context.Context, conn *pgxpool.Pool, email, hashedPassword string, method string) error {

	_, err := conn.Exec(ctx, "insert into users (email, hashed_password,registration_method) values ($1, $2,$3);", email, hashedPassword, method)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrEmailAlreadyExists
		}

		return err
	}
	return nil
}

func GetUserByEmail(ctx context.Context, conn *pgxpool.Pool, email string) (*UserDetails, error) {
	var user UserDetails
	row := conn.QueryRow(ctx, "select user_id,email,hashed_password,is_active,verification_status from users where email = $1;", email)
	err := row.Scan(&user.UserID, &user.Email, &user.HashedPassword, &user.IsActive, &user.VerificationStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrEmailDoesnTExist
		}
		return nil, err
	}
	return &user, nil
}
func GetUserByID(ctx context.Context, conn *pgxpool.Pool, userID string) (*UserResponse, error) {
	var user UserResponse
	row := conn.QueryRow(ctx, "select user_id,email,is_active,verification_status from users where user_id = $1;", userID)
	err := row.Scan(&user.UserID, &user.Email, &user.IsActive, &user.VerificationStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, UserDoesntExist
		}
		return nil, err
	}
	return &user, nil
}
