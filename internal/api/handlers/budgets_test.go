package handlers

import "testing"

// TestIsValidBudgetCategoryRejectsNonCanonical asserts a budget can only be
// created against a category transactions actually use — e.g. "Gas" (which
// CSV/Plaid categorization maps to "Transport") must be rejected, or a
// budget created against it would sit at zero spend forever.
func TestIsValidBudgetCategoryRejectsNonCanonical(t *testing.T) {
	cases := map[string]bool{
		"Groceries":     true,
		"Dining":        true,
		"Shopping":      true,
		"Transport":     true,
		"Travel":        true,
		"Subscriptions": true,
		"Entertainment": true,
		"Health":        true,
		"Housing":       true,
		"Income":        true,
		"Transfers":     true,
		"Other":         true,
		"Gas":           false,
		"gas":           false,
		"groceries":     false,
		"":              false,
	}
	for category, want := range cases {
		if got := isValidBudgetCategory(category); got != want {
			t.Errorf("isValidBudgetCategory(%q) = %v, want %v", category, got, want)
		}
	}
}
