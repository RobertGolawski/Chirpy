package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/RobertGolawski/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
	secret         string
	api            string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	p := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Sprintf("An error popped up: %v", err)
		return
	}
	s := os.Getenv("SECRET")
	a := os.Getenv("POLKA_KEY")
	var cfg apiConfig
	cfg.platform = p
	cfg.api = a
	dbQueries := database.New(db)
	cfg.queries = dbQueries
	cfg.secret = s
	var server = http.NewServeMux()
	server.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	server.Handle("./app/assets/logo.png", http.StripPrefix("/app/", http.FileServer(http.Dir("./assets/logo.png"))))
	server.HandleFunc("GET /api/healthz", cfg.getHealthz)
	server.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	server.HandleFunc("POST /admin/reset", cfg.resetAllUSers)
	// server.HandleFunc("POST /api/validate_chirp", cfg.validate_chirp)
	server.HandleFunc("POST /api/chirps", cfg.send_chirp)
	server.HandleFunc("GET /api/chirps", cfg.get_chirps)
	server.HandleFunc("GET /api/chirps/{id}", cfg.get_chirp_by_id)
	// server.HandleFunc("GET /api/chirps/{author_id}", cfg.get_chirps_for_user)
	server.HandleFunc("POST /api/users", cfg.createUserRequest)
	server.HandleFunc("POST /api/login", cfg.logInRequest)
	server.HandleFunc("POST /api/refresh", cfg.refreshJWT)
	server.HandleFunc("POST /api/revoke", cfg.revokeRefreshToken)
	server.HandleFunc("PUT /api/users", cfg.updateUserDetails)
	server.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpByID)
	server.HandleFunc("POST /api/polka/webhooks", cfg.upgradeUserToRed)
	var serverStruct = http.Server{
		Handler: server,
		Addr:    ":8080",
	}

	serverStruct.ListenAndServe()

}
