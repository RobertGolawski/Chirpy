package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/RobertGolawski/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) send_chirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string        `json:"body"`
		User_id uuid.NullUUID `json:"user_id"`
	}

	type chirpResponse struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    string    `json:"user_id"`
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

	validated, err := validate_chirp(params.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong during validation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	c, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validated,
		UserID: params.User_id})

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
