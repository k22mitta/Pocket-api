package ai

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/k22mitta/pocket-api/internal/db/queries"
)

func BuildFinancialContext(db *sql.DB, userID string) (string, error) {
	balance, err := queries.GetTotalBalance(db, userID)
	if err != nil {
		return "", fmt.Errorf("getting balance: %w", err)
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	spending, err := queries.GetSpendingByCategory(db, userID, startOfMonth.Format("2006-01-02"), endOfMonth.Format("2006-01-02"))
	if err != nil {
		return "", fmt.Errorf("getting spending: %w", err)
	}

	cashflow, err := queries.GetMonthlyTotals(db, userID, 3)
	if err != nil {
		return "", fmt.Errorf("getting cashflow: %w", err)
	}

	budgets, err := queries.GetBudgetsByUserID(db, userID)
	if err != nil {
		return "", fmt.Errorf("getting budgets: %w", err)
	}

	txns, err := queries.GetTransactionsByUserID(db, userID, queries.TransactionFilters{Limit: 10, Offset: 0})
	if err != nil {
		return "", fmt.Errorf("getting transactions: %w", err)
	}

	var sb strings.Builder

	sb.WriteString("You are Pocket, a personal finance assistant. You have access to the user's real financial data shown below.\n\n")

	sb.WriteString(fmt.Sprintf("CURRENT TOTAL BALANCE: $%.2f\n\n", balance))

	sb.WriteString("SPENDING BY CATEGORY THIS MONTH:\n")
	if len(spending) == 0 {
		sb.WriteString("  No spending recorded this month.\n")
	} else {
		for category, amount := range spending {
			sb.WriteString(fmt.Sprintf("  %s: $%.2f\n", category, amount))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("CASH FLOW (LAST 3 MONTHS):\n")
	if len(cashflow) == 0 {
		sb.WriteString("  No data available.\n")
	} else {
		for _, m := range cashflow {
			sb.WriteString(fmt.Sprintf("  %s — Income: $%.2f, Expenses: $%.2f\n", m.Month, m.Income, m.Expenses))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("ACTIVE BUDGETS:\n")
	if len(budgets) == 0 {
		sb.WriteString("  No budgets set.\n")
	} else {
		for _, b := range budgets {
			sb.WriteString(fmt.Sprintf("  %s: limit $%.2f, spent $%.2f, remaining $%.2f\n", b.Category, b.AmountLimit, b.Spent, b.Remaining))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("RECENT TRANSACTIONS (LAST 10):\n")
	if len(txns) == 0 {
		sb.WriteString("  No recent transactions.\n")
	} else {
		for _, t := range txns {
			sb.WriteString(fmt.Sprintf("  %s — %s: $%.2f\n", t.Date.Format("2006-01-02"), t.Name, t.Amount))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("INSTRUCTIONS:\n")
	sb.WriteString("- Give concise 2-4 sentence answers in plain English.\n")
	sb.WriteString("- Always reference specific numbers from the user's data above.\n")
	sb.WriteString("- Be direct about whether the user can afford something based on their real balance and spending patterns.\n")

	return sb.String(), nil
}
