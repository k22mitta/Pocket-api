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
	if _, err := db.Exec(`DELETE FROM budgets WHERE user_id = $1`, userID); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM plaid_items WHERE user_id = $1`, userID)
	return err
}

// CreateSeedAccount creates (or updates) a plaid_items + accounts pair that
// isn't backed by a real Plaid connection — used by both /demo/load (one call
// per seeded account) and CSV import (one call for the imported statement).
// slug must be unique per account within a user (e.g. "demo-checking", "csv-import").
func CreateSeedAccount(db *sql.DB, userID, slug, institutionName, accountName, acctType, subtype string, balance float64) (string, error) {
	var plaidItemID string
	err := db.QueryRow(
		`INSERT INTO plaid_items (user_id, access_token, item_id, institution_id, institution_name)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (item_id) DO UPDATE SET institution_name = EXCLUDED.institution_name
		 RETURNING id`,
		userID, "seed", slug+"-"+userID, "seed", institutionName,
	).Scan(&plaidItemID)
	if err != nil {
		return "", err
	}

	var accountID string
	err = db.QueryRow(
		`INSERT INTO accounts (plaid_item_id, user_id, plaid_account_id, name, type, subtype, current_balance, available_balance, currency_code)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (plaid_account_id) DO UPDATE SET
		   name              = EXCLUDED.name,
		   current_balance   = EXCLUDED.current_balance,
		   available_balance = EXCLUDED.available_balance,
		   updated_at        = NOW()
		 RETURNING id`,
		plaidItemID, userID, slug+"-"+userID, accountName, acctType, subtype, balance, balance, "USD",
	).Scan(&accountID)
	if err != nil {
		return "", err
	}

	return accountID, nil
}
