package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
)

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

func GetTransactionsByUserID(db *sql.DB, userID string, limit, offset int) ([]models.Transaction, error) {
	rows, err := db.Query(
		`SELECT id, user_id, account_id, plaid_transaction_id, merchant_name, name, amount, category, plaid_category, date, pending, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY date DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
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
