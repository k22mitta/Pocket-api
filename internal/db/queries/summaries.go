package queries

import (
	"database/sql"
)

type MonthlyTotal struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
}

func GetSpendingByCategory(db *sql.DB, userID, startDate, endDate string) (map[string]float64, error) {
	rows, err := db.Query(
		`SELECT category, COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE user_id = $1
		   AND date >= $2
		   AND date <= $3
		   AND amount > 0
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
		result[category] = total
	}
	return result, rows.Err()
}

func GetMonthlyTotals(db *sql.DB, userID string, months int) ([]MonthlyTotal, error) {
	rows, err := db.Query(
		`SELECT
		   TO_CHAR(date, 'YYYY-MM') as month,
		   COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as income,
		   COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as expenses
		 FROM transactions
		 WHERE user_id = $1
		   AND date >= NOW() - ($2 || ' months')::interval
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
		totals = append(totals, mt)
	}
	return totals, rows.Err()
}

func GetTotalBalance(db *sql.DB, userID string) (float64, error) {
	var balance float64
	err := db.QueryRow(
		`SELECT COALESCE(SUM(current_balance), 0) FROM accounts WHERE user_id = $1`,
		userID,
	).Scan(&balance)
	return balance, err
}
