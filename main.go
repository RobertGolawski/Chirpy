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
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Sprintf("An error popped up: %v", err)
		return
	}

	var cfg apiConfig
	dbQueries := database.New(db)
	cfg.queries = dbQueries
	var server = http.NewServeMux()
	server.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	server.Handle("./app/assets/logo.png", http.StripPrefix("/app/", http.FileServer(http.Dir("./assets/logo.png"))))
	server.HandleFunc("GET /api/healthz", cfg.getHealthz)
	server.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	server.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	server.HandleFunc("POST /api/validate_chirp", cfg.validate_chirp)
	var serverStruct = http.Server{
		Handler: server,
		Addr:    ":8080",
	}

	serverStruct.ListenAndServe()

}
