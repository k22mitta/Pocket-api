package models

import "time"

type Account struct {
	ID               string    `json:"id"`
	PlaidItemID      string    `json:"plaid_item_id"`
	UserID           string    `json:"user_id"`
	PlaidAccountID   string    `json:"plaid_account_id"`
	Name             string    `json:"name"`
	OfficialName     *string   `json:"official_name"`
	Type             string    `json:"type"`
	Subtype          *string   `json:"subtype"`
	CurrentBalance   *float64  `json:"current_balance"`
	AvailableBalance *float64  `json:"available_balance"`
	CurrencyCode     string    `json:"currency_code"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	InstitutionName  *string   `json:"-"`
	IsPlaidLinked    bool      `json:"-"`
}
