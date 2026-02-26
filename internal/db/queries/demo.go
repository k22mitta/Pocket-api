package queries

import (
	"database/sql"
)

func DeleteAllUserData(db *sql.DB, userID string) error {
	if _, err := db.Exec(`DELETE FROM transactions WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM accounts WHERE user_id = $1`, userID); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM plaid_items WHERE user_id = $1`, userID)
	return err
}

func CreateDemoAccount(db *sql.DB, userID string) (string, error) {
	var plaidItemID string
	err := db.QueryRow(
		`INSERT INTO plaid_items (user_id, access_token, item_id, institution_id, institution_name)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (item_id) DO UPDATE SET institution_name = EXCLUDED.institution_name
		 RETURNING id`,
		userID, "demo", "demo-"+userID, "demo", "Demo Bank",
	).Scan(&plaidItemID)
	if err != nil {
		return "", err
	}

	var accountID string
	err = db.QueryRow(
		`INSERT INTO accounts (plaid_item_id, user_id, plaid_account_id, name, type, subtype, current_balance, available_balance, currency_code)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (plaid_account_id) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		plaidItemID, userID, "demo-"+userID, "Demo Checking", "depository", "checking", 4827.50, 4827.50, "USD",
	).Scan(&accountID)
	if err != nil {
		return "", err
	}

	return accountID, nil
}
