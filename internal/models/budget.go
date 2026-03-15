package models

import "time"

// CanonicalCategories is the fixed set of categories transactions can be
// assigned (see plaidclient.NormalizeCategory and the CSV import's
// normalizeCategory). Budgets may only target one of these — a budget
// created against any other string can never match a transaction and would
// sit at zero spend forever.
var CanonicalCategories = []string{
	"Groceries",
	"Dining",
	"Shopping",
	"Transport",
	"Travel",
	"Subscriptions",
	"Entertainment",
	"Health",
	"Housing",
	"Income",
	"Transfers",
	"Other",
}

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
