package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourname/pocket-api/internal/ai"
	"github.com/yourname/pocket-api/internal/api/handlers"
	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/config"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
)

type router struct {
	db          *sql.DB
	cfg         config.Config
	plaidClient *plaidclient.Client
	aiClient    *ai.Client
}

func NewRouter(cfg config.Config, db *sql.DB, plaidClient *plaidclient.Client, aiClient *ai.Client) http.Handler {
	r := &router{db: db, cfg: cfg, plaidClient: plaidClient, aiClient: aiClient}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", r.handleHealth)
	mux.HandleFunc("POST /auth/register", handlers.Register(r.db))
	mux.HandleFunc("POST /auth/login", handlers.Login(r.db, r.cfg.JWTSecret))
	mux.Handle("GET /auth/me", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(r.handleMe)))
	mux.Handle("POST /plaid/link-token", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.CreateLinkToken(r.plaidClient, r.db))))
	mux.Handle("POST /plaid/exchange", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.ExchangeToken(r.plaidClient, r.db))))
	mux.Handle("POST /plaid/sync", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.SyncAll(r.plaidClient, r.db))))
	mux.Handle("GET /accounts", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetAccounts(r.db))))
	mux.Handle("DELETE /accounts/{id}", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.DeleteAccount(r.plaidClient, r.db))))
	mux.Handle("GET /transactions", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetTransactions(r.db))))
	mux.Handle("POST /budgets", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.CreateBudget(r.db))))
	mux.Handle("GET /budgets", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetBudgets(r.db))))
	mux.Handle("PUT /budgets/{id}", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.UpdateBudget(r.db))))
	mux.Handle("DELETE /budgets/{id}", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.DeleteBudget(r.db))))
	mux.Handle("GET /summary/spending", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetSpendingSummary(r.db))))
	mux.Handle("GET /summary/cashflow", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetCashFlow(r.db))))
	mux.Handle("GET /summary/balance", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetBalance(r.db))))
	mux.Handle("POST /chat", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.Chat(r.aiClient, r.db))))
	mux.Handle("GET /chat/history", middleware.RequireAuth(r.cfg.JWTSecret)(http.HandlerFunc(handlers.GetChatHistory(r.db))))

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
