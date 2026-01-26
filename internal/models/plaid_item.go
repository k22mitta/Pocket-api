package models

import "time"

type PlaidItem struct {
	ID              string
	UserID          string
	AccessToken     string
	ItemID          string
	InstitutionID   string
	InstitutionName string
	LastSyncedAt    *time.Time
	SyncError       *string
	CreatedAt       time.Time
}
