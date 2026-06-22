package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEmailAlreadyExists = errors.New("Email already exists")
var ErrEmailDoesnTExist = errors.New("no user under that email")

type UserDetails struct {
	UserID             string
	Email              string
	HashedPassword     string
	IsActive           bool
	VerificationStatus bool
}

func CreateUser(ctx context.Context, conn *pgx.Conn, email, hashedPassword string, method string) error {

	_, err := conn.Exec(ctx, "insert into users (email, hashed_password,registration_method) values ($1, $2,$3);", email, hashedPassword, method)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrEmailAlreadyExists
		}

		return err
	}
	return nil
}

func GetUserByEmail(ctx context.Context, conn *pgx.Conn, email string) (*UserDetails, error) {
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
