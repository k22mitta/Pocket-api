package queries

import (
	"database/sql"

	"github.com/k22mitta/pocket-api/internal/models"
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

// AdjustAccountBalance adds delta to an account's current (and available)
// balance, for callers that only know the incremental change (e.g. CSV
// import adding N new transactions to an already-existing account) rather
// than an authoritative absolute balance from a source like Plaid.
func AdjustAccountBalance(db *sql.DB, accountID string, delta float64) error {
	_, err := db.Exec(
		`UPDATE accounts SET
		   current_balance   = COALESCE(current_balance, 0) + $1,
		   available_balance = COALESCE(available_balance, 0) + $1,
		   updated_at        = NOW()
		 WHERE id = $2`,
		delta, accountID,
	)
	return err
}

// GetAccountsByUserID reports IsPlaidLinked by checking the parent
// plaid_item's access_token against the "seed" sentinel that both /demo/load
// and CSV import use for their synthetic plaid_items (see CreateSeedAccount)
// — anything else is a real Plaid Link connection with a real access token,
// which is the only kind of account that can be unlinked.
func GetAccountsByUserID(db *sql.DB, userID string) ([]models.Account, error) {
	rows, err := db.Query(
		`SELECT a.id, a.plaid_item_id, a.user_id, a.plaid_account_id, a.name, a.official_name, a.type, a.subtype,
		        a.current_balance, a.available_balance, a.currency_code, a.created_at, a.updated_at, p.institution_name,
		        p.access_token != 'seed'
		 FROM accounts a
		 LEFT JOIN plaid_items p ON p.id = a.plaid_item_id
		 WHERE a.user_id = $1
		 ORDER BY a.created_at ASC, a.id ASC`,
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
			&a.CurrencyCode, &a.CreatedAt, &a.UpdatedAt, &a.InstitutionName, &a.IsPlaidLinked,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
