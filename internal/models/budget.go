package models

import "time"

type Budget struct {
	ID          string
	UserID      string
	Category    string
	AmountLimit float64
	Period      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BudgetWithSpent struct {
	Budget
	Spent     float64
	Remaining float64
}
