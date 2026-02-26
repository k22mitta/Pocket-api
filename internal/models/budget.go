package models

import "time"

type Budget struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Category    string    `json:"category"`
	AmountLimit float64   `json:"amount_limit"`
	Period      string    `json:"period"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BudgetWithSpent struct {
	Budget
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
}
