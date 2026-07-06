package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/k22mitta/pocket-api/internal/api/middleware"
	"github.com/k22mitta/pocket-api/internal/db/queries"
)

func GetSpendingSummary(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)

		startDate := r.URL.Query().Get("start")
		endDate := r.URL.Query().Get("end")
		if startDate == "" {
			startDate = start.Format("2006-01-02")
		}
		if endDate == "" {
			endDate = end.Format("2006-01-02")
		}

		spending, err := queries.GetSpendingByCategory(db, userID, startDate, endDate)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"spending": spending})
	}
}

func GetCashFlow(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		months := 6
		if v := r.URL.Query().Get("months"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				months = n
			}
		}

		totals, err := queries.GetMonthlyTotals(db, userID, months)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if totals == nil {
			totals = []queries.MonthlyTotal{}
		}

		writeJSON(w, http.StatusOK, map[string]any{"cashflow": totals})
	}
}

func GetBalance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		balance, err := queries.GetTotalBalance(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]float64{"balance": balance})
	}
}
