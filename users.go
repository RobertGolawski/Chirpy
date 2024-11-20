package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/RobertGolawski/Chirpy/internal/auth"
	"github.com/RobertGolawski/Chirpy/internal/database"
)

type userResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createUserRequest(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("Error creating the user: here %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with user creation"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	hp, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error creating the hash: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with hashing"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if err := cfg.queries.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		HashedPassword: hp,
		ID:             u.ID,
	}); err != nil {
		log.Printf("Error updating the user: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with updating password"}
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

func (cfg *apiConfig) logInRequest(w http.ResponseWriter, r *http.Request) {
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

	u, err := cfg.queries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "user not found"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	err = auth.CheckPasswordHash(params.Password, u.HashedPassword)
	if err != nil {
		log.Printf("wrong pass tbh: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "incorrect email or password"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
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
