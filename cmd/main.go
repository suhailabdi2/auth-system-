package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/suhailabdi2/auth-system-/db"
	"github.com/suhailabdi2/auth-system-/internal/handlers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading environment variables", err)
	}
	port := os.Getenv("PORT")
	dbString := os.Getenv("DATABASE_URL")
	database, err := db.Connect(dbString)
	if err != nil {
		log.Fatal("Error connecting to the database", err)
	}
	server := mux.NewRouter()
	server.HandleFunc("/auth/register", handlers.RegisterHandler).Methods("POST")
	server.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")
	server.HandleFunc("/auth/refresh", handlers.RefreshTokensHandler).Methods("POST")
	server.HandleFunc("/auth/logout", handlers.LogoutHandler).Methods("POST")
	server.HandleFunc("/auth/me", handlers.MeHandler).Methods("GET")
	server.HandleFunc("/auth/google", handlers.GoogleHandler).Methods("GET")
	server.HandleFunc("/auth/google/callback", handlers.CallbackHandler).Methods("GET")
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
