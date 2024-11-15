package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) resetAllUSers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{"Body": "Forbidden"}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			resp := map[string]string{"error": "Something went wrong with response creation"}
			jsonResp, _ := json.Marshal(resp)
			w.Write(jsonResp)
			return
		}
		w.Write(jsonResp)
	}

	cfg.queries.ResetUsers(r.Context())
	w.WriteHeader(http.StatusOK)
}
