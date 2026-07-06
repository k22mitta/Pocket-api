package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/k22mitta/pocket-api/internal/ai"
	"github.com/k22mitta/pocket-api/internal/api/middleware"
	"github.com/k22mitta/pocket-api/internal/db/queries"
)

type cachedInsight struct {
	insight string
	at      time.Time
}

var (
	insightCache   = map[string]cachedInsight{}
	insightCacheMu sync.RWMutex
)

func GetSpendingInsight(aiClient *ai.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		forceRefresh := r.URL.Query().Get("refresh") == "true"

		if !forceRefresh {
			insightCacheMu.RLock()
			cached, ok := insightCache[userID]
			insightCacheMu.RUnlock()
			if ok && time.Since(cached.at) < 6*time.Hour {
				writeJSON(w, http.StatusOK, map[string]string{"insight": cached.insight})
				return
			}
		}

		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)

		spending, err := queries.GetSpendingByCategory(db, userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		balance, err := queries.GetTotalBalance(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Current balance: $%.2f\n", balance))
		sb.WriteString(fmt.Sprintf("Spending this month (%s) by category:\n", now.Format("January 2006")))
		total := 0.0
		for cat, amt := range spending {
			sb.WriteString(fmt.Sprintf("  %s: $%.2f\n", cat, amt))
			total += amt
		}
		sb.WriteString(fmt.Sprintf("Total: $%.2f\n", total))
		sb.WriteString("\nGive 2-3 specific, actionable insights about their spending. Be concise and direct.")

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		insight, err := aiClient.Chat(
			ctx,
			"You are Pocket, a personal finance assistant. Analyze the user's spending data and provide clear, helpful insights.",
			sb.String(),
			nil,
		)
		if err != nil {
			log.Printf("insight generation failed: %v", err)
			writeJSON(w, http.StatusOK, map[string]string{"insight": ""})
			return
		}

		insightCacheMu.Lock()
		insightCache[userID] = cachedInsight{insight: insight, at: time.Now()}
		insightCacheMu.Unlock()

		writeJSON(w, http.StatusOK, map[string]string{"insight": insight})
	}
}

func GetOverdraftRisk(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		balance, err := queries.GetTotalBalance(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		now := time.Now()
		spending, err := queries.GetSpendingByCategory(db, userID, now.AddDate(0, 0, -30).Format("2006-01-02"), now.Format("2006-01-02"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		totalSpent := 0.0
		for _, amt := range spending {
			totalSpent += amt
		}

		dailyAvg := totalSpent / 30.0
		daysUntilOverdraft := 0.0
		risk := "low"

		if dailyAvg > 0 {
			daysUntilOverdraft = balance / dailyAvg
			switch {
			case daysUntilOverdraft < 7:
				risk = "high"
			case daysUntilOverdraft < 30:
				risk = "medium"
			default:
				risk = "low"
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"risk":                 risk,
			"balance":              balance,
			"daily_avg_spend":      dailyAvg,
			"days_until_overdraft": daysUntilOverdraft,
		})
	}
}
