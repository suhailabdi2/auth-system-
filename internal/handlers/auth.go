package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

func RegisterHandler(conn *pgxpool.Pool) http.HandlerFunc {
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

func LoginHandler(conn *pgxpool.Pool) http.HandlerFunc {
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

func RefreshTokensHandler(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refReq RefreshRequest
		var res LoginResponse
		if err := json.NewDecoder(r.Body).Decode(&refReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Missing Refresh Token")
			return
		}
		AccessToken, RefreshToken, err := services.RefreshToken(r.Context(), conn, refReq.RefreshToken)
		if err != nil {
			if err == services.ErrTokenExpired {
				WriteError(w, http.StatusForbidden, "token expired")
				return
			}
			if err == services.ErrTokenReuse {
				WriteError(w, http.StatusForbidden, "token already used")
				return
			}
			WriteError(w, http.StatusForbidden, "token expired")
			return
		}
		res.AccessToken = AccessToken
		res.RefreshToken = RefreshToken
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func LogoutHandler(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refReq RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&refReq); err != nil {
			WriteError(w, http.StatusBadRequest, "Missing refresh token")
			return
		}
		if err := repository.RevokeRefreshToken(r.Context(), conn, refReq.RefreshToken); err != nil {
			WriteError(w, http.StatusInternalServerError, "Error revoking token")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func MeHandler(conn *pgxpool.Pool) http.HandlerFunc {
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
	client := OAuthGoogle()
	log.Print(client.ClientID)
	log.Println(client.ClientSecret)
	url := client.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusFound)

}
func CallbackHandler(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// existing codec
		client := OAuthGoogle()
		code := r.URL.Query().Get("code")
		var user GoogleUser

		if code == "" {
			WriteError(w, http.StatusBadRequest, "Missing code")
			return
		}
		token, err := client.Exchange(r.Context(), code)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Error generating token")
			return
		}
		httpClient := client.Client(r.Context(), token)
		resp, err := httpClient.Get("https://www.googleapis.com/oauth2/v2/userinfo")

		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Error getting response from google")
			return
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusInternalServerError, "Error decoding")
			return
		}
		existingUser, err := repository.GetUserByEmail(r.Context(), conn, user.Email)
		if err != nil {
			if err == repository.ErrEmailDoesnTExist {
				if err := repository.CreateUser(r.Context(), conn, user.Email, "", "google"); err != nil {
					WriteError(w, http.StatusInternalServerError, "Error creating new user")
					return
				}
				newUser, err := repository.GetUserByEmail(r.Context(), conn, user.Email)
				if err != nil {
					WriteError(w, http.StatusInternalServerError, "Error getting new user")
					return
				}
				issueTokens(w, r, conn, newUser.UserID, newUser.Email, http.StatusCreated)
				return
			} else {
				WriteError(w, http.StatusInternalServerError, "Error getting user")
				return
			}
		}
		issueTokens(w, r, conn, existingUser.UserID, existingUser.Email, http.StatusOK)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
func issueTokens(w http.ResponseWriter, r *http.Request, conn *pgxpool.Pool, userID, email string, status int) {
	refreshToken := services.GenerateRefreshToken()
	if err := repository.StoreRefreshToken(r.Context(), conn, refreshToken, userID, time.Now().Add(7*24*time.Hour)); err != nil {
		WriteError(w, http.StatusInternalServerError, "Error storing refresh token")
		return
	}
	accessToken, err := services.GenerateAccessToken(userID, email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error generating token")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}
