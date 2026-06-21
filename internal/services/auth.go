package services

import (
	"context"
	"errors"
	"net/mail"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/suhailabdi2/auth-system-/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func validatePassword(password string) error {
	match1, err := regexp.MatchString(`^.{8,}$)`, password)
	if err != nil {
		return err
	}
	match2, err := regexp.MatchString(`\d`, password)
	if err != nil {
		return err
	}
	if (!match1) || (!match2) {
		return errors.New("password too weak")
	}
	return nil
}
func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}
	return nil
}

func Register(ctx context.Context, conn *pgx.Conn, email, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	if err := repository.CreateUser(ctx, conn, email, string(hashedPassword), "password"); err != nil {
		return err
	}
	return nil

}
