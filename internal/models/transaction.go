package models

import "time"

type Transaction struct {
	ID                 string
	UserID             string
	AccountID          string
	PlaidTransactionID string
	MerchantName       *string
	Name               string
	Amount             float64
	Category           string
	PlaidCategory      *string
	Date               time.Time
	Pending            bool
	CreatedAt          time.Time
}
