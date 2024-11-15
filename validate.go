package main

// import (
// 	"encoding/json"
// 	"net/http"
// 	"strings"
// )

// func (cfg *apiConfig) validate_chirp(w http.ResponseWriter, r *http.Request) {
// 	type parameters struct {
// 		Body string `json:"body"`
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	decoder := json.NewDecoder(r.Body)
// 	params := parameters{}
// 	err := decoder.Decode(&params)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		resp := map[string]string{"error": "Something went wrong"}
// 		jsonResp, _ := json.Marshal(resp)
// 		w.Write(jsonResp)
// 		return
// 	}
// 	if len(params.Body) > 140 {
// 		w.WriteHeader(400)
// 		resp := map[string]string{"error": "Chirp is too long"}
// 		jsonResp, _ := json.Marshal(resp)
// 		w.Write(jsonResp)
// 		return
// 	}

// 	// words := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax", "KERFUFFLE", "SHARBERT", "FORNAX"}

// 	badWords := map[string]struct{}{
// 		"kerfuffle": {},
// 		"sharbert":  {},
// 		"fornax":    {},
// 	}

// 	words := strings.Split(params.Body, " ")
// 	for i, word := range words {
// 		lowered := strings.ToLower(word)
// 		if _, ok := badWords[lowered]; ok {
// 			words[i] = "****"
// 		}
// 	}

// 	// for _, word := range words {
// 	// 	if strings.Contains(params.Body, word) {
// 	// 		params.Body = strings.Replace(params.Body, word, "****", -1)
// 	// 	}
// 	// }

// 	w.WriteHeader(200)
// 	resp := map[string]string{"cleaned_body": params.Body}
// 	jsonResp, _ := json.Marshal(resp)
// 	w.Write(jsonResp)
// //}
