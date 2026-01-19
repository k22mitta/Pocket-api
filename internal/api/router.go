package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourname/pocket-api/internal/api/handlers"
	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/config"
)

type router struct {
	db  *sql.DB
	cfg config.Config
}

func NewRouter(cfg config.Config, db *sql.DB) http.Handler {
	r := &router{db: db, cfg: cfg}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", r.handleHealth)
	mux.HandleFunc("POST /auth/register", handlers.Register(r.db))
	mux.HandleFunc("POST /auth/login", handlers.Login(r.db, r.cfg.JWTSecret))
	mux.Handle("GET /auth/me", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(r.handleMe)))

	return mux
}

func (r *router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (r *router) handleMe(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":    middleware.UserIDFromContext(req.Context()),
		"email": middleware.UserEmailFromContext(req.Context()),
	})
}
