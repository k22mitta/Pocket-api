package queries

import (
	"database/sql"
	"fmt"

	"github.com/yourname/pocket-api/internal/models"
)

type TransactionFilters struct {
	Search    string
	Category  string
	AccountID string
	StartDate string
	EndDate   string
	Limit     int
	Offset    int
}

// filterWhereClause builds the shared "WHERE user_id = $1 AND pending = false
// [AND ...]" clause and its args, so the list query and the count query can
// never drift apart on which rows they consider a match.
func filterWhereClause(userID string, filters TransactionFilters) (string, []interface{}) {
	args := []interface{}{userID}
	n := 2

	clause := `WHERE user_id = $1 AND pending = false`

	if filters.Search != "" {
		clause += fmt.Sprintf(` AND (LOWER(name) LIKE LOWER($%d) OR LOWER(merchant_name) LIKE LOWER($%d))`, n, n)
		args = append(args, "%"+filters.Search+"%")
		n++
	}
	if filters.Category != "" {
		clause += fmt.Sprintf(` AND category = $%d`, n)
		args = append(args, filters.Category)
		n++
	}
	if filters.AccountID != "" {
		clause += fmt.Sprintf(` AND account_id = $%d`, n)
		args = append(args, filters.AccountID)
		n++
	}
	if filters.StartDate != "" {
		clause += fmt.Sprintf(` AND date >= $%d`, n)
		args = append(args, filters.StartDate)
		n++
	}
	if filters.EndDate != "" {
		clause += fmt.Sprintf(` AND date <= $%d`, n)
		args = append(args, filters.EndDate)
		n++
	}

	return clause, args
}

func GetTransactionsByUserID(db *sql.DB, userID string, filters TransactionFilters) ([]models.Transaction, error) {
	where, args := filterWhereClause(userID, filters)
	n := len(args) + 1

	query := `SELECT id, user_id, account_id, plaid_transaction_id, merchant_name, name, amount, category, plaid_category, date, pending, created_at
	          FROM transactions ` + where

	query += fmt.Sprintf(` ORDER BY date DESC, id DESC LIMIT $%d OFFSET $%d`, n, n+1)
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

// CountTransactionsByUserID reports the total number of transactions matching
// filters, ignoring Limit/Offset, so callers can paginate and show a true
// "N of M" count instead of silently capping at the page size.
func CountTransactionsByUserID(db *sql.DB, userID string, filters TransactionFilters) (int, error) {
	where, args := filterWhereClause(userID, filters)
	query := `SELECT COUNT(*) FROM transactions ` + where

	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UpsertTransaction inserts a transaction, skipping it if a row with the same
// plaid_transaction_id already exists. Returns whether a new row was actually
// inserted, so callers can tell a genuine addition from a deduped re-import.
func UpsertTransaction(db *sql.DB, t models.Transaction) (bool, error) {
	res, err := db.Exec(
		`INSERT INTO transactions (user_id, account_id, plaid_transaction_id, merchant_name, name, amount, category, plaid_category, date, pending)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (plaid_transaction_id) DO NOTHING`,
		t.UserID, t.AccountID, t.PlaidTransactionID, t.MerchantName,
		t.Name, t.Amount, t.Category, t.PlaidCategory, t.Date, t.Pending,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
