package queries

import (
	"database/sql"
	"fmt"

	"github.com/yourname/pocket-api/internal/models"
)

type TransactionFilters struct {
	Search    string
	Category  string
	StartDate string
	EndDate   string
	Limit     int
	Offset    int
}

func GetTransactionsByUserID(db *sql.DB, userID string, filters TransactionFilters) ([]models.Transaction, error) {
	args := []interface{}{userID}
	n := 2

	query := `SELECT id, user_id, account_id, plaid_transaction_id, merchant_name, name, amount, category, plaid_category, date, pending, created_at
	          FROM transactions WHERE user_id = $1 AND pending = false`

	if filters.Search != "" {
		query += fmt.Sprintf(` AND (LOWER(name) LIKE LOWER($%d) OR LOWER(merchant_name) LIKE LOWER($%d))`, n, n)
		args = append(args, "%"+filters.Search+"%")
		n++
	}
	if filters.Category != "" {
		query += fmt.Sprintf(` AND category = $%d`, n)
		args = append(args, filters.Category)
		n++
	}
	if filters.StartDate != "" {
		query += fmt.Sprintf(` AND date >= $%d`, n)
		args = append(args, filters.StartDate)
		n++
	}
	if filters.EndDate != "" {
		query += fmt.Sprintf(` AND date <= $%d`, n)
		args = append(args, filters.EndDate)
		n++
	}

	query += fmt.Sprintf(` ORDER BY date DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.PlaidTransactionID,
			&t.MerchantName, &t.Name, &t.Amount, &t.Category, &t.PlaidCategory,
			&t.Date, &t.Pending, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func UpsertTransaction(db *sql.DB, t models.Transaction) error {
	_, err := db.Exec(
		`INSERT INTO transactions (user_id, account_id, plaid_transaction_id, merchant_name, name, amount, category, plaid_category, date, pending)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (plaid_transaction_id) DO NOTHING`,
		t.UserID, t.AccountID, t.PlaidTransactionID, t.MerchantName,
		t.Name, t.Amount, t.Category, t.PlaidCategory, t.Date, t.Pending,
	)
	return err
}
