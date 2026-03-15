package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
	"github.com/yourname/pocket-api/internal/money"
)

func CreateBudget(db *sql.DB, userID, category, period string, limit float64) (models.Budget, error) {
	var b models.Budget
	err := db.QueryRow(
		`INSERT INTO budgets (user_id, category, amount_limit, period)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, category, amount_limit, period, created_at, updated_at`,
		userID, category, limit, period,
	).Scan(&b.ID, &b.UserID, &b.Category, &b.AmountLimit, &b.Period, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func GetBudgetsByUserID(db *sql.DB, userID string) ([]models.BudgetWithSpent, error) {
	rows, err := db.Query(
		`SELECT b.id, b.user_id, b.category, b.amount_limit, b.period, b.created_at, b.updated_at,
		   COALESCE(SUM(t.amount), 0) as spent
		 FROM budgets b
		 LEFT JOIN transactions t ON t.user_id = b.user_id
		   AND t.category = b.category
		   AND t.pending = false
		   AND t.amount > 0
		   AND DATE_TRUNC('month', t.date) = DATE_TRUNC('month', NOW())
		 WHERE b.user_id = $1
		 GROUP BY b.id
		 ORDER BY b.created_at ASC, b.id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []models.BudgetWithSpent
	for rows.Next() {
		var bws models.BudgetWithSpent
		if err := rows.Scan(
			&bws.ID, &bws.UserID, &bws.Category, &bws.AmountLimit, &bws.Period,
			&bws.CreatedAt, &bws.UpdatedAt, &bws.Spent,
		); err != nil {
			return nil, err
		}
		bws.Spent = money.Round2(bws.Spent)
		bws.Remaining = money.Round2(bws.AmountLimit - bws.Spent)
		budgets = append(budgets, bws)
	}
	return budgets, rows.Err()
}

func UpdateBudget(db *sql.DB, budgetID, userID string, limit float64) (models.Budget, error) {
	var b models.Budget
	err := db.QueryRow(
		`UPDATE budgets SET amount_limit = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3
		 RETURNING id, user_id, category, amount_limit, period, created_at, updated_at`,
		limit, budgetID, userID,
	).Scan(&b.ID, &b.UserID, &b.Category, &b.AmountLimit, &b.Period, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func DeleteBudget(db *sql.DB, budgetID, userID string) error {
	_, err := db.Exec(`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, budgetID, userID)
	return err
}
