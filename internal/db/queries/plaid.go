package queries

import (
	"database/sql"

	"github.com/k22mitta/pocket-api/internal/models"
)

func CreatePlaidItem(db *sql.DB, userID, accessToken, itemID, institutionID, institutionName string) error {
	_, err := db.Exec(
		`INSERT INTO plaid_items (user_id, access_token, item_id, institution_id, institution_name)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, accessToken, itemID, institutionID, institutionName,
	)
	return err
}

func GetPlaidItemByItemID(db *sql.DB, itemID string) (models.PlaidItem, error) {
	var item models.PlaidItem
	err := db.QueryRow(
		`SELECT id, user_id, access_token, item_id, institution_id, institution_name, last_synced_at, sync_error, created_at
		 FROM plaid_items WHERE item_id = $1`,
		itemID,
	).Scan(
		&item.ID, &item.UserID, &item.AccessToken, &item.ItemID,
		&item.InstitutionID, &item.InstitutionName,
		&item.LastSyncedAt, &item.SyncError, &item.CreatedAt,
	)
	return item, err
}

func UpdateLastSynced(db *sql.DB, itemID string) error {
	_, err := db.Exec(`UPDATE plaid_items SET last_synced_at = NOW() WHERE id = $1`, itemID)
	return err
}

func GetPlaidItemByID(db *sql.DB, itemID, userID string) (models.PlaidItem, error) {
	var item models.PlaidItem
	err := db.QueryRow(
		`SELECT id, user_id, access_token, item_id, institution_id, institution_name, last_synced_at, sync_error, created_at
		 FROM plaid_items WHERE id = $1 AND user_id = $2`,
		itemID, userID,
	).Scan(
		&item.ID, &item.UserID, &item.AccessToken, &item.ItemID,
		&item.InstitutionID, &item.InstitutionName,
		&item.LastSyncedAt, &item.SyncError, &item.CreatedAt,
	)
	return item, err
}

func DeletePlaidItem(db *sql.DB, itemID, userID string) error {
	_, err := db.Exec(`DELETE FROM plaid_items WHERE id = $1 AND user_id = $2`, itemID, userID)
	return err
}

func GetAllPlaidItems(db *sql.DB) ([]models.PlaidItem, error) {
	rows, err := db.Query(
		`SELECT id, user_id, access_token, item_id, institution_id, institution_name, last_synced_at, sync_error, created_at
		 FROM plaid_items`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PlaidItem
	for rows.Next() {
		var item models.PlaidItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.AccessToken, &item.ItemID,
			&item.InstitutionID, &item.InstitutionName,
			&item.LastSyncedAt, &item.SyncError, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetPlaidItemsByUserID(db *sql.DB, userID string) ([]models.PlaidItem, error) {
	rows, err := db.Query(
		`SELECT id, user_id, access_token, item_id, institution_id, institution_name, last_synced_at, sync_error, created_at
		 FROM plaid_items WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PlaidItem
	for rows.Next() {
		var item models.PlaidItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.AccessToken, &item.ItemID,
			&item.InstitutionID, &item.InstitutionName,
			&item.LastSyncedAt, &item.SyncError, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
