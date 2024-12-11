package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/RobertGolawski/Chirpy/internal/auth"
	"github.com/RobertGolawski/Chirpy/internal/database"
	"github.com/google/uuid"
)

type userResponse struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Expiration time.Duration `json:"expires_in_seconds"`
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

	tokenString, err := auth.MakeJWT(u.ID, cfg.secret, 3600*time.Second)
	if err != nil {
		log.Printf("Error making the JWT in user create: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong with making the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	checkID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		log.Printf("Error with validation of the JWT in user create: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong with validating the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	if checkID != u.ID {
		log.Printf("User ID did not match during validation.")
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong with user comparison"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error making a refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong refresh token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	nullID := uuid.NullUUID{UUID: checkID, Valid: true}
	err = cfg.queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    nullID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("Error inserting the refresh token into db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong refresh token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := userResponse{
		ID:           u.ID.String(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Email:        u.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
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

	tokenString, err := auth.MakeJWT(u.ID, cfg.secret, 3600*time.Second)
	if err != nil {
		log.Printf("Error making the JWT in user create: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with making the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	checkID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		log.Printf("Error with validation of the JWT in user create: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with validating the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	if checkID != u.ID {
		log.Printf("User ID did not match during validation.")
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with user comparison"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error making a refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong refresh token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	nullID := uuid.NullUUID{UUID: checkID, Valid: true}
	err = cfg.queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    nullID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("Error inserting the refresh token into db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Something went wrong refresh token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := userResponse{
		ID:           u.ID.String(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Email:        u.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
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

func (cfg *apiConfig) refreshJWT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with getting bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	tokenData, err := cfg.queries.GetTokenData(r.Context(), bearerToken)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "Invalid bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "expired bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if tokenData.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "revoked bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	tokenString, err := auth.MakeJWT(tokenData.UserID.UUID, cfg.secret, 3600*time.Second)
	if err != nil {
		log.Printf("Error making the JWT in user create: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with making the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	checkID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		log.Printf("Error with validation of the JWT in user create: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with validating the jwt"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	if checkID != tokenData.UserID.UUID {
		log.Printf("User ID did not match during validation.")
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with user comparison"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
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

func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "Something went wrong with getting bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	tokenData, err := cfg.queries.GetTokenData(r.Context(), bearerToken)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "Invalid bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "expired bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if tokenData.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]string{"error": "revoked bearer token"}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "Error revoking the "}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
