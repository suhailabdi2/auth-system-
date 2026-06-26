package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func MeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Missing token")
			return
		}
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				w.WriteHeader(http.StatusForbidden)
				return nil, fmt.Errorf("unexpected signin method")
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "invalid")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Invalid token")
			return
		}
		userID := claims["user_id"].(string)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func RateLimiter(client *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ClientIp, _, err := net.SplitHostPort(r.RemoteAddr)
			key := "rate_limit:" + ClientIp
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "Error getting ip address")
				return
			}
			count, err := client.Incr(context.Background(), key).Result()
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "Error getting count")
				return
			}
			if count == 1 {
				client.Expire(context.Background(), key, time.Minute)
			}
			if count > 5 {
				WriteError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)

		})
	}
}
