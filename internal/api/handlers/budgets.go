package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
)

// budgetResponse is the camelCase wire contract the frontend consumes —
// matching the treatment accounts/transactions already get. Drops user_id
// (never needed by the client, and not something to leak) and created_at/
// updated_at (unused by the frontend).
type budgetResponse struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"`
	AmountLimit float64 `json:"amountLimit"`
	Period      string  `json:"period"`
	Spent       float64 `json:"spent"`
	Remaining   float64 `json:"remaining"`
}

func isValidBudgetCategory(category string) bool {
	for _, c := range models.CanonicalCategories {
		if c == category {
			return true
		}
	}
	return false
}

func toBudgetResponse(b models.BudgetWithSpent) budgetResponse {
	return budgetResponse{
		ID:          b.ID,
		Category:    b.Category,
		AmountLimit: b.AmountLimit,
		Period:      b.Period,
		Spent:       b.Spent,
		Remaining:   b.Remaining,
	}
}

func CreateBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		var body struct {
			Category    string  `json:"category"`
			AmountLimit float64 `json:"amountLimit"`
			Period      string  `json:"period"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if body.Category == "" || body.AmountLimit == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category and amountLimit are required"})
			return
		}
		if !isValidBudgetCategory(body.Category) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category must be one of the canonical categories transactions use"})
			return
		}
		if body.Period == "" {
			body.Period = "monthly"
		}

		budget, err := queries.CreateBudget(db, userID, body.Category, body.Period, body.AmountLimit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusCreated, toBudgetResponse(withSpent(db, userID, budget)))
	}
}

// withSpent re-derives spent/remaining for a single just-created-or-updated
// budget so a POST/PUT response never reports spent as 0 when the category
// already has real spend this month — the same numbers /budgets (GET) would
// compute, just for one row instead of the whole list.
func withSpent(db *sql.DB, userID string, b models.Budget) models.BudgetWithSpent {
	all, err := queries.GetBudgetsByUserID(db, userID)
	if err == nil {
		for _, bws := range all {
			if bws.ID == b.ID {
				return bws
			}
		}
	}
	return models.BudgetWithSpent{Budget: b}
}

func GetBudgets(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		budgets, err := queries.GetBudgetsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		resp := make([]budgetResponse, len(budgets))
		for i, b := range budgets {
			resp[i] = toBudgetResponse(b)
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func UpdateBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		budgetID := r.PathValue("id")

		var body struct {
			AmountLimit float64 `json:"amountLimit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		budget, err := queries.UpdateBudget(db, budgetID, userID, body.AmountLimit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, toBudgetResponse(withSpent(db, userID, budget)))
	}
}

func DeleteBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		budgetID := r.PathValue("id")

		if err := queries.DeleteBudget(db, budgetID, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}
