package models

import "time"

type Transaction struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	AccountID          string     `json:"account_id"`
	PlaidTransactionID string     `json:"plaid_transaction_id"`
	MerchantName       *string    `json:"merchant_name"`
	Name               string     `json:"name"`
	Amount             float64    `json:"amount"`
	Category           string     `json:"category"`
	PlaidCategory      *string    `json:"plaid_category"`
	Date               time.Time  `json:"date"`
	Pending            bool       `json:"pending"`
	CreatedAt          time.Time  `json:"created_at"`
}
