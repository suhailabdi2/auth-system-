package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/suhailabdi2/auth-system-/internal/repository"
	"github.com/suhailabdi2/auth-system-/internal/services"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RegisterHandler(conn *pgx.Conn) http.HandlerFunc {
	//todo
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handler reached")
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		if err := services.Register(ctx, conn, req.Email, req.Password); err != nil {
			if err == repository.ErrEmailAlreadyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func LoginHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		var res LoginResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		AccessToken, RefreshToken, err := services.Login(r.Context(), conn, req.Email, req.Password)
		if err != nil {
			if err == services.WrongPassword || err == services.InActiveUser {
				w.WriteHeader(http.StatusForbidden)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		res.AccessToken = AccessToken
		res.RefreshToken = RefreshToken
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

}
func MeHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDKey := r.Context().Value(UserIDKey)
		userID := userIDKey.(string)
		user, err := repository.GetUserByID(r.Context(), conn, userID)
		if err != nil {
			if err == repository.UserDoesntExist {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "User ID not found")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error getting user id")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}

}

func GoogleHandler(w http.ResponseWriter, r *http.Request) {

}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {

}
