package handlers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/k22mitta/pocket-api/internal/api/middleware"
	"github.com/k22mitta/pocket-api/internal/db/queries"
	"github.com/k22mitta/pocket-api/internal/models"
	"github.com/k22mitta/pocket-api/internal/money"
)

// DemoEmail identifies the single shared demo account. /demo/load and
// /demo/reset are destructive (they wipe all of the caller's data) and must
// never be reachable by a real user's account, whether by a UI bug, a stale
// tab, or someone hitting the endpoint directly.
const DemoEmail = "demo@example.com"

func requireDemoUser(w http.ResponseWriter, r *http.Request) bool {
	if middleware.UserEmailFromContext(r.Context()) != DemoEmail {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "this endpoint is only available to the demo account"})
		return false
	}
	return true
}

func ResetData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireDemoUser(w, r) {
			return
		}
		userID := middleware.UserIDFromContext(r.Context())
		if err := queries.DeleteAllUserData(db, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset data"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}

const (
	acctChecking = 0
	acctSavings  = 1
	acctCredit   = 2
)

type demoMerchant struct {
	name     string
	merchant string
	category string
	min      float64
	max      float64
	account  int
}

// Recurring merchant pool the generator draws from for the day-by-day tail.
// Amounts are picked within [min, max] per occurrence. Groceries and Dining
// are separate categories (not merged) so their budgets track independently.
var demoMerchantPool = []demoMerchant{
	{"Whole Foods Market", "Whole Foods", "Groceries", 35.00, 110.00, acctChecking},
	{"Trader Joe's", "Trader Joe's", "Groceries", 28.00, 75.00, acctCredit},
	{"Starbucks", "Starbucks", "Dining", 4.50, 8.75, acctCredit},
	{"Chipotle Mexican Grill", "Chipotle", "Dining", 10.50, 15.25, acctCredit},
	{"Uber Eats", "Uber Eats", "Dining", 18.00, 42.00, acctCredit},
	{"Amazon.com", "Amazon", "Shopping", 15.00, 130.00, acctCredit},
	{"Target", "Target", "Shopping", 22.00, 95.00, acctCredit},
	{"Nike.com", "Nike", "Shopping", 45.00, 120.00, acctCredit},
	{"Best Buy", "Best Buy", "Shopping", 40.00, 220.00, acctCredit},
	{"Uber", "Uber", "Transport", 9.50, 28.00, acctCredit},
	{"Lyft", "Lyft", "Transport", 8.75, 24.00, acctCredit},
	{"Shell Gas Station", "Shell", "Transport", 38.00, 58.00, acctChecking},
	{"Metro Transit", "Metro Transit", "Transport", 2.75, 6.00, acctCredit},
	{"Netflix", "Netflix", "Subscriptions", 15.99, 15.99, acctChecking},
	{"Spotify", "Spotify", "Subscriptions", 9.99, 9.99, acctChecking},
	{"AMC Theatres", "AMC", "Entertainment", 18.00, 42.00, acctCredit},
	{"Steam", "Steam", "Entertainment", 12.00, 60.00, acctCredit},
	{"CVS Pharmacy", "CVS", "Health", 12.00, 45.00, acctCredit},
	{"LA Fitness", "LA Fitness", "Health", 40.00, 40.00, acctChecking},
	{"Walgreens", "Walgreens", "Health", 8.00, 30.00, acctCredit},
}

func LoadDemoData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireDemoUser(w, r) {
			return
		}
		userID := middleware.UserIDFromContext(r.Context())

		if err := queries.DeleteAllUserData(db, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reset data"})
			return
		}

		accountIDs := make([]string, 3)
		var err error
		accountIDs[acctChecking], err = queries.CreateSeedAccount(
			db, userID, "demo-checking", "Meridian Bank", "Everyday Checking", "depository", "checking", 3842.19,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create demo accounts"})
			return
		}
		accountIDs[acctSavings], err = queries.CreateSeedAccount(
			db, userID, "demo-savings", "Meridian Bank", "High-Yield Savings", "depository", "savings", 12480.55,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create demo accounts"})
			return
		}
		accountIDs[acctCredit], err = queries.CreateSeedAccount(
			db, userID, "demo-credit", "Horizon Card Services", "Rewards Credit Card", "credit", "credit card", 612.40,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create demo accounts"})
			return
		}

		now := time.Now()
		rng := rand.New(rand.NewSource(42))
		created := 0
		seq := 0

		insert := func(daysAgo int, name, merchant, category string, amount float64, account int) {
			plaidTxID := fmt.Sprintf("demo-%s-%d", userID, seq)
			seq++
			m := merchant
			t := models.Transaction{
				UserID:             userID,
				AccountID:          accountIDs[account],
				PlaidTransactionID: plaidTxID,
				MerchantName:       &m,
				Name:               name,
				Amount:             amount,
				Category:           category,
				Date:               now.AddDate(0, 0, -daysAgo),
				Pending:            false,
			}
			if inserted, err := queries.UpsertTransaction(db, t); err == nil && inserted {
				created++
			}
		}

		for daysAgo := 1; daysAgo <= 85; daysAgo += 14 {
			insert(daysAgo, "Payroll Deposit", "Employer", "Income", -2850.00, acctChecking)
		}
		for daysAgo := 2; daysAgo <= 85; daysAgo += 30 {
			insert(daysAgo, "Rent Payment", "Meridian Properties", "Housing", 1650.00, acctChecking)
		}
		for daysAgo := 5; daysAgo <= 85; daysAgo += 30 {
			insert(daysAgo, "Gym Membership", "LA Fitness", "Health", 40.00, acctChecking)
		}
		for daysAgo := 10; daysAgo <= 85; daysAgo += 30 {
			insert(daysAgo, "Interest Payment", "Meridian Bank", "Income", -18.42, acctSavings)
		}

		// Day-by-day recurring spend, dense near "today" so every budget
		// category has coverage in the current calendar month no matter what
		// day it is, and a long tail back to ~85 days for cash-flow history.
		// Guarantee at least one transaction per budgeted category today and
		// over the last few days regardless of RNG.
		guaranteedCategories := []string{"Groceries", "Dining", "Shopping", "Transport", "Subscriptions", "Entertainment", "Health", "Housing"}
		for _, cat := range guaranteedCategories {
			for _, m := range demoMerchantPool {
				if m.category == cat {
					amt := m.min + rng.Float64()*(m.max-m.min)
					// Dated "today" (daysAgo=0) so every budget shows partial
					// spend regardless of which day of the calendar month it is.
					insert(0, m.name, m.merchant, m.category, money.Round2(amt), m.account)
					break
				}
			}
		}

		for daysAgo := 0; daysAgo <= 85; daysAgo++ {
			hits := 1
			if rng.Float64() < 0.4 {
				hits = 2
			}
			if rng.Float64() < 0.1 {
				hits = 0
			}
			for h := 0; h < hits; h++ {
				m := demoMerchantPool[rng.Intn(len(demoMerchantPool))]
				amt := m.min + rng.Float64()*(m.max-m.min)
				insert(daysAgo, m.name, m.merchant, m.category, money.Round2(amt), m.account)
			}
		}

		budgetDefs := []struct {
			category string
			limit    float64
		}{
			{"Groceries", 350.00},
			{"Dining", 200.00},
			{"Shopping", 250.00},
			{"Transport", 150.00},
			{"Subscriptions", 30.00},
			{"Entertainment", 60.00},
			{"Health", 100.00},
			{"Housing", 1700.00},
		}
		for _, b := range budgetDefs {
			if _, err := queries.CreateBudget(db, userID, b.category, "monthly", b.limit); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create demo budgets"})
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success":              true,
			"transactions_created": created,
		})
	}
}
