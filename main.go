package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetrics(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	var cfg apiConfig
	var server = http.NewServeMux()
	server.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	server.Handle("./app/assets/logo.png", http.StripPrefix("/app/", http.FileServer(http.Dir("./assets/logo.png"))))
	server.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		hits := cfg.fileserverHits.Load()
		w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hits)))
	})
	server.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
	})
	server.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "Something went wrong"}
			jsonResp, _ := json.Marshal(resp)
			w.Write(jsonResp)
			return
		}
		if len(params.Body) > 140 {
			w.WriteHeader(400)
			resp := map[string]string{"error": "Chirp is too long"}
			jsonResp, _ := json.Marshal(resp)
			w.Write(jsonResp)
			return
		}

		w.WriteHeader(200)
		resp := map[string]bool{"valid": true}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
	})
	var serverStruct = http.Server{
		Handler: server,
		Addr:    ":8080",
	}

	serverStruct.ListenAndServe()

}
