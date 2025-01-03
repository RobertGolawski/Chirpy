package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/RobertGolawski/Chirpy/internal/auth"
	"github.com/RobertGolawski/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func (cfg *apiConfig) get_chirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	if s != "" {
		cfg.get_chirps_for_user(w, r)
		return
	}
	order := "asc"
	o := r.URL.Query().Get("sort")
	if o == "desc" {
		order = "desc"
	}

	w.Header().Set("Content-Type", "application/json")
	cs, err := cfg.queries.GetChirps(r.Context(), order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during retrieval"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	var chirps []chirpResponse

	for _, c := range cs {
		chirps = append(chirps, chirpResponse{
			ID:        c.ID.String(),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID.UUID.String(),
		})
	}

	jsonResp, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during response generation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func (cfg *apiConfig) get_chirp_by_id(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("id")
	parsedPath, err := uuid.Parse(path)
	if err != nil {
		w.WriteHeader(404)
		resp := map[string]string{"error": "Something went wrong during parsing"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	c, err := cfg.queries.GetChirp(r.Context(), parsedPath)
	if err != nil {
		w.WriteHeader(404)
		resp := map[string]string{"error": "Something went wrong during querying"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	resp := chirpResponse{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID.UUID.String(),
	}

	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func (cfg *apiConfig) send_chirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string        `json:"body"`
		User_id uuid.NullUUID `json:"user_id"`
		Token   string        `json:"token"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong during decoding"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	params.Token, err = auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong during bearer token retrieval"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	validated, err := validate_chirp(params.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong during validation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	userID, err := auth.ValidateJWT(params.Token, cfg.secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("error during validation: %v", err)
		resp := map[string]string{"error": "Something went wrong jwt validation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	// if userID != params.User_id.UUID {
	// 	log.Printf("Error happened here with user id: %v but was expecting %v", userID.String(), params.User_id.UUID.String())
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	resp := map[string]string{"error": "Something went wrong jwt validation"}
	// 	jsonResp, _ := json.Marshal(resp)
	// 	w.Write(jsonResp)
	// 	return
	// }
	nullID := uuid.NullUUID{UUID: userID, Valid: true}
	c, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validated,
		UserID: nullID})
	//uuid.NullUUID{UUID: userID, Valid: true}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during chirp creation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	resp := chirpResponse{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID.UUID.String(),
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during marshalling"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResp)
}

func validate_chirp(s string) (string, error) {

	if len(s) > 140 {
		return " ", errors.New("chirp too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(s, " ")
	for i, word := range words {
		lowered := strings.ToLower(word)
		if _, ok := badWords[lowered]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " "), nil
}

func (cfg *apiConfig) deleteChirpByID(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("chirpID")
	parsedPath, err := uuid.Parse(path)
	if err != nil {
		w.WriteHeader(404)
		resp := map[string]string{"error": "Something went wrong during parsing"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "Something went wrong with getting bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		log.Printf("Error with validation of the JWT in user update: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "Something went wrong with validating the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	c, err := cfg.queries.GetChirp(r.Context(), parsedPath)
	if err != nil {
		w.WriteHeader(404)
		resp := map[string]string{"error": "Something went wrong during querying"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if userID != c.UserID.UUID {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{"error": "Forbidden"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	err = cfg.queries.DeleteChirp(r.Context(), parsedPath)
	if err != nil {
		log.Fatalf("Error during database operation to delete chirp: %v", err)
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{"error": "Error with deleting chirp"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) get_chirps_for_user(w http.ResponseWriter, r *http.Request) {
	order := "asc"
	o := r.URL.Query().Get("sort")
	if o == "desc" {
		order = "desc"
	}
	s := r.URL.Query().Get("author_id")
	parsedPath, err := uuid.Parse(s)
	if err != nil {
		w.WriteHeader(404)
		resp := map[string]string{"error": "Something went wrong during parsing"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	nullID := uuid.NullUUID{UUID: parsedPath, Valid: true}

	w.Header().Set("Content-Type", "application/json")
	cs, err := cfg.queries.GetUserChirps(r.Context(), database.GetUserChirpsParams{UserID: nullID, SortOrder: order})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during retrieval"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	var chirps []chirpResponse

	for _, c := range cs {
		chirps = append(chirps, chirpResponse{
			ID:        c.ID.String(),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID.UUID.String(),
		})
	}

	jsonResp, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong during response generation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
