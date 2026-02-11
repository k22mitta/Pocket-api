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

func CreateBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		var body struct {
			Category    string  `json:"category"`
			AmountLimit float64 `json:"amount_limit"`
			Period      string  `json:"period"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if body.Category == "" || body.AmountLimit == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category and amount_limit are required"})
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

		writeJSON(w, http.StatusCreated, budget)
	}
}

func GetBudgets(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		budgets, err := queries.GetBudgetsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if budgets == nil {
			budgets = []models.BudgetWithSpent{}
		}

		writeJSON(w, http.StatusOK, budgets)
	}
}

func UpdateBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		budgetID := r.PathValue("id")

		var body struct {
			AmountLimit float64 `json:"amount_limit"`
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

		writeJSON(w, http.StatusOK, budget)
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
