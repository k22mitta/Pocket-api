package models

import "time"

type Account struct {
	ID               string
	PlaidItemID      string
	UserID           string
	PlaidAccountID   string
	Name             string
	OfficialName     *string
	Type             string
	Subtype          *string
	CurrentBalance   *float64
	AvailableBalance *float64
	CurrencyCode     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
