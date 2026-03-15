package handlers

import (
	"testing"

	"github.com/yourname/pocket-api/internal/models"
)

func strPtr(s string) *string   { return &s }
func f64Ptr(f float64) *float64 { return &f }

// TestToAccountResponseLiabilityTypes asserts Plaid's "loan" type (student
// loans, mortgages) maps to its own "loan" frontend type and is negated —
// never left on the default "checking" case, which would add debt to net
// position as if it were an asset.
func TestToAccountResponseLiabilityTypes(t *testing.T) {
	cases := []struct {
		name        string
		plaidType   string
		subtype     string
		rawBalance  float64
		wantFEType  string
		wantBalance float64
	}{
		{"checking", "depository", "checking", 110, "checking", 110},
		{"savings", "depository", "savings", 210, "savings", 210},
		{"cd", "depository", "cd", 1000, "savings", 1000},
		{"money market", "depository", "money market", 43200, "savings", 43200},
		{"credit card", "credit", "credit card", 410, "credit", -410},
		{"student loan", "loan", "student", 65262, "loan", -65262},
		{"mortgage", "loan", "mortgage", 56302.06, "loan", -56302.06},
		{"investment (brokerage type)", "brokerage", "brokerage", 320.76, "investment", 320.76},
		{"investment (investment subtype)", "depository", "investment", 23631.98, "investment", 23631.98},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := models.Account{
				Type:           c.plaidType,
				Subtype:        strPtr(c.subtype),
				CurrentBalance: f64Ptr(c.rawBalance),
			}
			got := toAccountResponse(a)
			if got.Type != c.wantFEType {
				t.Errorf("Type = %q, want %q", got.Type, c.wantFEType)
			}
			if got.Balance != c.wantBalance {
				t.Errorf("Balance = %v, want %v (raw Plaid balance %v, isLiability=%v)",
					got.Balance, c.wantBalance, c.rawBalance, isLiabilityType(c.plaidType))
			}
		})
	}
}

// TestMixedAssetLiabilityNetPositionToTheCent asserts a mix of depository,
// investment, credit, and loan accounts nets to the cent once liability
// balances are correctly negated.
func TestMixedAssetLiabilityNetPositionToTheCent(t *testing.T) {
	accounts := []models.Account{
		{Type: "depository", Subtype: strPtr("checking"), CurrentBalance: f64Ptr(5845.01)},
		{Type: "depository", Subtype: strPtr("checking"), CurrentBalance: f64Ptr(110)},
		{Type: "depository", Subtype: strPtr("savings"), CurrentBalance: f64Ptr(210)},
		{Type: "depository", Subtype: strPtr("cd"), CurrentBalance: f64Ptr(1000)},
		{Type: "credit", Subtype: strPtr("credit card"), CurrentBalance: f64Ptr(410)},
		{Type: "depository", Subtype: strPtr("money market"), CurrentBalance: f64Ptr(43200)},
		{Type: "depository", Subtype: strPtr("investment"), CurrentBalance: f64Ptr(320.76)},
		{Type: "depository", Subtype: strPtr("investment"), CurrentBalance: f64Ptr(23631.98)},
		{Type: "loan", Subtype: strPtr("student"), CurrentBalance: f64Ptr(65262)},
		{Type: "loan", Subtype: strPtr("mortgage"), CurrentBalance: f64Ptr(56302.06)},
		{Type: "depository", Subtype: strPtr("checking"), CurrentBalance: f64Ptr(6009)},
		{Type: "depository", Subtype: strPtr("checking"), CurrentBalance: f64Ptr(12060)},
		{Type: "credit", Subtype: strPtr("credit card"), CurrentBalance: f64Ptr(5020)},
	}

	total := 0.0
	for _, a := range accounts {
		total += toAccountResponse(a).Balance
	}

	// Hand-computed: assets (5845.01+110+210+1000+43200+320.76+23631.98+6009+12060)
	// = 92386.75; liabilities (410+65262+56302.06+5020) = 126994.06;
	// net = 92386.75 - 126994.06 = -34607.31.
	want := -34607.31
	if total != want {
		t.Errorf("net position = %v, want %v", total, want)
	}
}
