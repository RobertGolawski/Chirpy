package main

import (
	"net/http"
)

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}
