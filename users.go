package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) createUserRequest(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type userResponse struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with parsing JSON"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	u, err := cfg.queries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating the user: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with user creation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := userResponse{
		ID:        u.ID.String(),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with response creation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	w.Write(jsonResp)

}
