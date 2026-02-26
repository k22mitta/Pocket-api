package models

import "time"

type PlaidItem struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	AccessToken     string     `json:"-"`
	ItemID          string     `json:"item_id"`
	InstitutionID   string     `json:"institution_id"`
	InstitutionName string     `json:"institution_name"`
	LastSyncedAt    *time.Time `json:"last_synced_at"`
	SyncError       *string    `json:"sync_error"`
	CreatedAt       time.Time  `json:"created_at"`
}
