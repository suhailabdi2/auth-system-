package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/suhailabdi2/auth-system-/internal/services"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(conn *pgx.Conn) http.HandlerFunc {
	//todo
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		if err := services.Register(ctx, conn, req.Email, req.Password); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	//todo
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

}
func MeHandler(w http.ResponseWriter, r *http.Request) {

}

func GoogleHandler(w http.ResponseWriter, r *http.Request) {

}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {

}
