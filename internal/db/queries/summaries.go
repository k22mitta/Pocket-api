package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/money"
)

type MonthlyTotal struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
}

// GetSpendingByCategory excludes the "Transfers" category: moving money
// between the user's own accounts (or paying down a loan/credit card) isn't
// discretionary spending, and counting it inflates "spent this month" by
// the transfer amount even though the user's net worth didn't change.
func GetSpendingByCategory(db *sql.DB, userID, startDate, endDate string) (map[string]float64, error) {
	rows, err := db.Query(
		`SELECT category, COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE user_id = $1
		   AND date >= $2
		   AND date <= $3
		   AND amount > 0
		   AND category != 'Transfers'
		   AND pending = false
		 GROUP BY category`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			return nil, err
		}
		result[category] = money.Round2(total)
	}
	return result, rows.Err()
}

// GetMonthlyTotals excludes "Transfers" from both income and expenses for
// the same reason as GetSpendingByCategory: an internal transfer or debt
// payment isn't earned income or discretionary spending.
func GetMonthlyTotals(db *sql.DB, userID string, months int) ([]MonthlyTotal, error) {
	rows, err := db.Query(
		`SELECT
		   TO_CHAR(date, 'YYYY-MM') as month,
		   COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as income,
		   COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as expenses
		 FROM transactions
		 WHERE user_id = $1
		   AND date >= NOW() - ($2 || ' months')::interval
		   AND category != 'Transfers'
		   AND pending = false
		 GROUP BY month
		 ORDER BY month DESC`,
		userID, months,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []MonthlyTotal
	for rows.Next() {
		var mt MonthlyTotal
		if err := rows.Scan(&mt.Month, &mt.Income, &mt.Expenses); err != nil {
			return nil, err
		}
		mt.Income = money.Round2(mt.Income)
		mt.Expenses = money.Round2(mt.Expenses)
		totals = append(totals, mt)
	}
	return totals, rows.Err()
}

// GetTotalBalance sums account balances the same way the /accounts endpoint
// displays them: credit and loan accounts (credit cards, mortgages, student
// loans) store the amount owed as a positive number, but represent negative
// equity, so they're subtracted rather than added. Must stay in exact sync
// with handlers.isLiabilityType — this is the same liability check, just
// expressed in SQL instead of Go, so /accounts and /summary/balance (and the
// dashboard net position built from it) never disagree about which accounts
// are liabilities.
func GetTotalBalance(db *sql.DB, userID string) (float64, error) {
	var balance float64
	err := db.QueryRow(
		`SELECT COALESCE(SUM(
		   CASE WHEN type IN ('credit', 'loan') THEN -current_balance ELSE current_balance END
		 ), 0) FROM accounts WHERE user_id = $1`,
		userID,
	).Scan(&balance)
	return money.Round2(balance), err
}
