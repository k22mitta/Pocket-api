package api

import (
	"encoding/json"
	"net/http"

	"github.com/yourname/pocket-api/internal/config"
)

func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	return mux
}
