package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
)

func ResetData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		if err := queries.DeleteAllUserData(db, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset data"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}

func LoadDemoData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		if err := queries.DeleteAllUserData(db, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset data"})
			return
		}

		accountID, err := queries.CreateDemoAccount(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create demo account"})
			return
		}

		now := time.Now()

		type demoTx struct {
			name     string
			merchant string
			amount   float64
			category string
			daysAgo  int
		}

		txDefs := []demoTx{
			{"Uber", "Uber", 14.50, "Transport", 2},
			{"Netflix", "Netflix", 15.99, "Entertainment", 1},
			{"Starbucks", "Starbucks", 6.75, "Food & Drink", 3},
			{"Amazon.com", "Amazon", 45.99, "Shopping", 5},
			{"CVS Pharmacy", "CVS", 23.47, "Health", 6},
			{"Chipotle Mexican Grill", "Chipotle", 12.50, "Food & Drink", 7},
			{"Shell Gas Station", "Shell", 52.30, "Transport", 8},
			{"Whole Foods Market", "Whole Foods", 87.34, "Food & Drink", 10},
			{"Target", "Target", 112.43, "Shopping", 12},
			{"Uber Eats", "Uber Eats", 34.20, "Food & Drink", 14},
			{"Lyft", "Lyft", 11.75, "Transport", 15},
			{"Payroll Deposit", "Employer", -2800.00, "Income", 15},
			{"AMC Theatres", "AMC", 32.50, "Entertainment", 16},
			{"Starbucks", "Starbucks", 5.25, "Food & Drink", 18},
			{"Nike.com", "Nike", 89.00, "Shopping", 20},
			{"CVS Pharmacy", "CVS", 31.20, "Health", 20},
			{"Chipotle Mexican Grill", "Chipotle", 11.85, "Food & Drink", 22},
			{"Uber", "Uber", 18.90, "Transport", 25},
			{"Whole Foods Market", "Whole Foods", 64.12, "Food & Drink", 28},
			{"Rent Payment", "Property Management", 1500.00, "Housing", 30},
			{"Gym Membership", "LA Fitness", 40.00, "Health", 30},
			{"Spotify", "Spotify", 9.99, "Entertainment", 31},
			{"Amazon.com", "Amazon", 23.45, "Shopping", 32},
			{"Uber Eats", "Uber Eats", 28.90, "Food & Drink", 35},
			{"Shell Gas Station", "Shell", 48.60, "Transport", 38},
			{"Starbucks", "Starbucks", 7.50, "Food & Drink", 42},
			{"CVS Pharmacy", "CVS", 18.90, "Health", 44},
			{"Payroll Deposit", "Employer", -2800.00, "Income", 45},
			{"Target", "Target", 67.21, "Shopping", 45},
			{"Target", "Target", 89.55, "Shopping", 48},
			{"Lyft", "Lyft", 9.50, "Transport", 50},
			{"Amazon.com", "Amazon", 156.78, "Shopping", 52},
			{"Starbucks", "Starbucks", 8.25, "Food & Drink", 54},
			{"Uber", "Uber", 22.30, "Transport", 55},
			{"Chipotle Mexican Grill", "Chipotle", 13.75, "Food & Drink", 56},
			{"Shell Gas Station", "Shell", 55.10, "Transport", 57},
			{"Whole Foods Market", "Whole Foods", 93.40, "Food & Drink", 58},
			{"Netflix", "Netflix", 15.99, "Entertainment", 60},
			{"Rent Payment", "Property Management", 1500.00, "Housing", 60},
			{"Gym Membership", "LA Fitness", 40.00, "Health", 60},
		}

		created := 0
		for i, td := range txDefs {
			merchant := td.merchant
			plaidTxID := "demo-" + userID + "-" + string(rune('a'+i))
			t := models.Transaction{
				UserID:             userID,
				AccountID:          accountID,
				PlaidTransactionID: plaidTxID,
				MerchantName:       &merchant,
				Name:               td.name,
				Amount:             td.amount,
				Category:           td.category,
				Date:               now.AddDate(0, 0, -td.daysAgo),
				Pending:            false,
			}
			if err := queries.UpsertTransaction(db, t); err == nil {
				created++
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success":              true,
			"transactions_created": created,
		})
	}
}
