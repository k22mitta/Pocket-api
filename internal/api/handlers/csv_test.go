package handlers

import "testing"

func TestParseDate(t *testing.T) {
	cases := []struct {
		in   string
		want string // formatted back as YYYY-MM-DD
	}{
		{"2026-07-04", "2026-07-04"},
		{"07/04/2026", "2026-07-04"},
		{"2026/07/04", "2026-07-04"},
		{"04 Jul 2026", "2026-07-04"},
		{"Jul 4, 2026", "2026-07-04"},
	}
	for _, c := range cases {
		got, err := parseDate(c.in)
		if err != nil {
			t.Errorf("parseDate(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.Format("2006-01-02") != c.want {
			t.Errorf("parseDate(%q) = %v, want %s", c.in, got, c.want)
		}
	}
}

func TestParseDateRejectsGarbage(t *testing.T) {
	if _, err := parseDate("not-a-date"); err == nil {
		t.Error("parseDate(\"not-a-date\") expected an error, got nil")
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"-1,234.56", -1234.56},
		{"$1,234.56", 1234.56},
		{"-$48.20", -48.20},
		{"  45.00  ", 45.00},
		{"€12.50", 12.50},
	}
	for _, c := range cases {
		got, err := parseAmount(c.in)
		if err != nil {
			t.Errorf("parseAmount(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseAmount(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestNormalizeCategoryGroceriesVsDining asserts Groceries and Dining
// resolve to different canonical categories — merging them would let a
// "Groceries" budget silently absorb restaurant spend.
func TestNormalizeCategoryGroceriesVsDining(t *testing.T) {
	groceries := normalizeCategory("Groceries")
	dining := normalizeCategory("Dining")
	if groceries != "Groceries" {
		t.Errorf(`normalizeCategory("Groceries") = %q, want "Groceries"`, groceries)
	}
	if dining != "Dining" {
		t.Errorf(`normalizeCategory("Dining") = %q, want "Dining"`, dining)
	}
	if groceries == dining {
		t.Fatal("Groceries and Dining must not normalize to the same category")
	}
}

// TestNormalizeCategorySubscriptionsVsEntertainment: Subscriptions (recurring
// auto-pay like Netflix/Spotify) is a distinct category from one-off
// Entertainment spend, so a "Subscriptions" budget can track recurring
// charges without merging in movie tickets, games, etc.
func TestNormalizeCategorySubscriptionsVsEntertainment(t *testing.T) {
	subscriptions := normalizeCategory("Subscriptions")
	entertainment := normalizeCategory("Entertainment")
	if subscriptions != "Subscriptions" {
		t.Errorf(`normalizeCategory("Subscriptions") = %q, want "Subscriptions"`, subscriptions)
	}
	if entertainment != "Entertainment" {
		t.Errorf(`normalizeCategory("Entertainment") = %q, want "Entertainment"`, entertainment)
	}
	if subscriptions == entertainment {
		t.Fatal("Subscriptions and Entertainment must not normalize to the same category")
	}
}

func TestNormalizeCategoryLowercaseAndVariants(t *testing.T) {
	cases := map[string]string{
		"groceries":     "Groceries",
		"dining out":    "Dining",
		"transport":     "Transport",
		"Gas":           "Transport",
		"Rent":          "Housing",
		"Subscriptions": "Subscriptions",
		"streaming":     "Subscriptions",
		"Travel":        "Travel",
		"Income":        "Income",
	}
	for in, want := range cases {
		if got := normalizeCategory(in); got != want {
			t.Errorf("normalizeCategory(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeCategoryUnknownKeepsUserLabel(t *testing.T) {
	got := normalizeCategory("Pet Supplies")
	if got != "Pet Supplies" {
		t.Errorf(`normalizeCategory("Pet Supplies") = %q, want "Pet Supplies" (unrecognized categories keep the user's label)`, got)
	}
}

// TestCleanMerchantNameNoTrailingJunk asserts "Direct Deposit - Acme Corp"
// doesn't become "Direct Deposit -", and "Trader Joes" doesn't lose its
// second word to the state-code stripper.
func TestCleanMerchantNameNoTrailingJunk(t *testing.T) {
	cases := map[string]string{
		"Direct Deposit - Acme Corp": "Direct Deposit Acme",
		"Trader Joes":                "Trader Joes",
		"Oak Street Apartments":      "Oak Street Apartments",
		"Whole Foods Market":         "Whole Foods Market",
		"Shell Gas Station":          "Shell", // known brand, short-circuits before word-limit trimming
	}
	for in, want := range cases {
		got := CleanMerchantName(in)
		if got != want {
			t.Errorf("CleanMerchantName(%q) = %q, want %q", in, got, want)
		}
		if len(got) > 0 && !isLetterOrDigit(rune(got[len(got)-1])) {
			t.Errorf("CleanMerchantName(%q) = %q ends in non-alphanumeric junk", in, got)
		}
	}
}

func isLetterOrDigit(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// TestMatchColumnIsDeterministic asserts matchColumn resolves a header
// matching aliases for two different columns (e.g. "Transaction Date"
// contains both the "date" alias "transaction date" and the "description"
// alias "transaction") the same way every run, not based on Go map iteration
// order.
func TestMatchColumnIsDeterministic(t *testing.T) {
	for i := 0; i < 100; i++ {
		if got := matchColumn("transaction date"); got != "date" {
			t.Fatalf(`matchColumn("transaction date") = %q, want "date" (run %d)`, got, i)
		}
	}
}

func TestMatchColumnRecognizesMerchantAsDescription(t *testing.T) {
	if got := matchColumn("merchant"); got != "description" {
		t.Errorf(`matchColumn("merchant") = %q, want "description"`, got)
	}
}
