package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/yourname/pocket-api/internal/config"
)

type router struct {
	db *sql.DB
}

func NewRouter(cfg config.Config, db *sql.DB) http.Handler {
	r := &router{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", r.handleHealth)

	return mux
}

func (r *router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
