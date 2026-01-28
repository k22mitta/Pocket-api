package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
)

func UpsertAccount(db *sql.DB, a models.Account) error {
	_, err := db.Exec(
		`INSERT INTO accounts (plaid_item_id, user_id, plaid_account_id, name, official_name, type, subtype, current_balance, available_balance, currency_code)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (plaid_account_id) DO UPDATE SET
		   name              = EXCLUDED.name,
		   current_balance   = EXCLUDED.current_balance,
		   available_balance = EXCLUDED.available_balance,
		   updated_at        = NOW()`,
		a.PlaidItemID, a.UserID, a.PlaidAccountID, a.Name, a.OfficialName,
		a.Type, a.Subtype, a.CurrentBalance, a.AvailableBalance, a.CurrencyCode,
	)
	return err
}

func GetAccountsByUserID(db *sql.DB, userID string) ([]models.Account, error) {
	rows, err := db.Query(
		`SELECT id, plaid_item_id, user_id, plaid_account_id, name, official_name, type, subtype,
		        current_balance, available_balance, currency_code, created_at, updated_at
		 FROM accounts WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(
			&a.ID, &a.PlaidItemID, &a.UserID, &a.PlaidAccountID, &a.Name, &a.OfficialName,
			&a.Type, &a.Subtype, &a.CurrentBalance, &a.AvailableBalance,
			&a.CurrencyCode, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
