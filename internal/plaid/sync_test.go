package plaidclient

import "testing"

// TestCategorizeTransactionIncomeSignGuard asserts transactions Plaid tags
// with income-like categories ("Transfer", which used to map to Income) but
// that are actually outflows (positive amount, per Plaid's own sign
// convention) never render as paychecks the user never received. A genuine
// inflow with the same raw category must still resolve correctly.
func TestCategorizeTransactionIncomeSignGuard(t *testing.T) {
	cases := []struct {
		name       string
		categories []string
		amount     float64
		merchant   string
		want       string
	}{
		{"outflow tagged Transfer -> not Income", []string{"Transfer"}, 5850, "ACH Electronic CreditGUSTO PAY 123456", "Transfers"},
		// Real Plaid data for this one is a 2-element category array
		// ["Transfer", "Deposit"] — the "Deposit" element alone would match
		// the Income case (checked before Transfers in NormalizeCategory's
		// switch), AND the merchant-name fallback also independently matches
		// "deposit" in "CD DEPOSIT .INITIAL." — both paths must be
		// sign-guarded. Since the raw category is transfer-shaped, this
		// prefers Transfers over a generic "Other" (so it's excluded from
		// spending, not silently counted as an unlabeled expense).
		{"outflow tagged Transfer+Deposit (CD deposit) -> Transfers, not Income or Other", []string{"Transfer", "Deposit"}, 1000, "CD DEPOSIT .INITIAL.", "Transfers"},
		// A genuine inflow tagged Transfer still isn't real "earned" income,
		// but must never be miscategorized as an outflow-style expense either.
		{"inflow tagged Transfer -> Transfers, not Income or expense", []string{"Transfer"}, -4.22, "INTRST PYMNT", "Transfers"},
		// Genuine payroll: negative (inflow) with a Payroll-style category.
		{"inflow tagged Payroll -> Income", []string{"Payroll"}, -2800, "Employer Direct Deposit", "Income"},
		// A hypothetical outflow mistakenly tagged with payroll/deposit
		// keywords must not become Income either — the general sign guard
		// applies regardless of which income-like keyword matched.
		{"outflow tagged Deposit -> falls back to merchant categorization", []string{"Deposit"}, 42.00, "XYZ Unrecognized Merchant 12345", "Other"},
		// Unrelated categories are unaffected by the sign guard.
		{"unaffected: Food and Drink", []string{"Food and Drink", "Restaurants"}, 25.00, "Chipotle", "Dining"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := categorizeTransaction(c.categories, c.amount, c.merchant)
			if got != c.want {
				t.Errorf("categorizeTransaction(%v, %v, %q) = %q, want %q", c.categories, c.amount, c.merchant, got, c.want)
			}
		})
	}
}

// TestNormalizeCategoryTransfersNotIncome guards against "transfer" ever
// being re-added to the Income keyword list — it must resolve to its own
// Transfers category, not Income, regardless of amount (the sign guard in
// categorizeTransaction handles direction; NormalizeCategory itself is
// sign-unaware and must not conflate the two).
func TestNormalizeCategoryTransfersNotIncome(t *testing.T) {
	got := NormalizeCategory([]string{"Transfer"})
	if got != "Transfers" {
		t.Errorf(`NormalizeCategory(["Transfer"]) = %q, want "Transfers"`, got)
	}
	got = NormalizeCategory([]string{"Payment"})
	if got != "Transfers" {
		t.Errorf(`NormalizeCategory(["Payment"]) = %q, want "Transfers"`, got)
	}
}
