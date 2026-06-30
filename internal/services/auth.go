package services

import (
	"context"
	"crypto/rand"
	"errors"
	"net/mail"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suhailabdi2/auth-system-/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var WrongPassword = errors.New("Passwords don't match")
var InActiveUser = errors.New("User is inactive")
var ErrTokenReuse = errors.New("Token already revoked")
var ErrTokenExpired = errors.New("Token expired")

func validatePassword(password string) error {
	match1, err := regexp.MatchString(`^.{8,}$`, password)
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

func Register(ctx context.Context, conn *pgxpool.Pool, email, password string) error {
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
func Login(ctx context.Context, conn *pgxpool.Pool, email, password string) (string, string, error) {
	user, err := repository.GetUserByEmail(ctx, conn, email)
	if err != nil {
		return "", "", err
	}
	if !user.IsActive {
		return "", "", InActiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", "", WrongPassword
	}
	// if tokenString,err := GenerateAccessToken(user.UserID,email); err != nil{
	// 	return "","",err
	// }
	tokenString, err := GenerateAccessToken(user.UserID, email)
	if err != nil {
		return "", "", err
	}
	RefreshToken := GenerateRefreshToken()
	if err := repository.StoreRefreshToken(ctx, conn, RefreshToken, user.UserID, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}
	return tokenString, RefreshToken, nil
}
func GenerateAccessToken(userID, email string) (string, error) {
	var tokenSecret = os.Getenv("JWT_SECRET_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", errors.New("error generating a new token")
	}
	return tokenString, nil

}
func GenerateRefreshToken() string {
	RefreshToken := rand.Text()
	return RefreshToken
}
func RefreshToken(ctx context.Context, conn *pgxpool.Pool, tokenStr string) (string, string, error) {
	OldToken, err := repository.GetRefreshToken(ctx, conn, tokenStr)

	if err != nil {
		return "", "", err
	}
	UserDetails, err := repository.GetUserByID(ctx, conn, OldToken.UserID)
	if err != nil {
		return "", "", err
	}
	if OldToken.RevokedAt != nil {
		if err := repository.RevokeRefreshTokensByUser(ctx, conn, OldToken.UserID); err != nil {
			return "", "", ErrTokenReuse
		}
		return "", "", ErrTokenReuse
	}
	if time.Now().After(OldToken.ExpiryDate) {
		return "", "", ErrTokenExpired
	}
	NewToken := GenerateRefreshToken()
	if err := repository.StoreRefreshToken(ctx, conn, NewToken, OldToken.UserID, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}
	AccessToken, err := GenerateAccessToken(UserDetails.UserID, UserDetails.Email)
	if err != nil {
		return "", "", err
	}
	return NewToken, AccessToken, nil
}
