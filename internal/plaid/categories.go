package plaidclient

import "strings"

// NormalizeCategory maps Plaid's category taxonomy onto Pocket's canonical
// budget categories. Groceries and Dining are kept separate (not merged into
// one "Food & Drink" bucket) because users create budgets against these
// exact category names, and a merged bucket makes a "Groceries" budget
// silently include restaurant spend, which is a real correctness bug, not
// just a display preference.
func NormalizeCategory(plaidCategories []string) string {
	if len(plaidCategories) == 0 {
		return "Other"
	}

	combined := strings.ToLower(strings.Join(plaidCategories, " "))

	switch {
	case strings.Contains(combined, "grocery") || strings.Contains(combined, "groceries") ||
		strings.Contains(combined, "supermarket"):
		return "Groceries"
	case strings.Contains(combined, "restaurant") || strings.Contains(combined, "coffee") ||
		strings.Contains(combined, "dining") || strings.Contains(combined, "food"):
		return "Dining"
	case strings.Contains(combined, "shop") || strings.Contains(combined, "retail") ||
		strings.Contains(combined, "store") || strings.Contains(combined, "merchandise"):
		return "Shopping"
	case strings.Contains(combined, "travel") || strings.Contains(combined, "airline") ||
		strings.Contains(combined, "hotel") || strings.Contains(combined, "lodging") ||
		strings.Contains(combined, "flight"):
		return "Travel"
	case strings.Contains(combined, "transport") || strings.Contains(combined, "taxi") ||
		strings.Contains(combined, "gas station") || strings.Contains(combined, "fuel") ||
		strings.Contains(combined, "parking") || strings.Contains(combined, "toll") ||
		strings.Contains(combined, "rideshare"):
		return "Transport"
	case strings.Contains(combined, "housing") || strings.Contains(combined, "rent") ||
		strings.Contains(combined, "utilities") || strings.Contains(combined, "mortgage"):
		return "Housing"
	case strings.Contains(combined, "subscription") || strings.Contains(combined, "streaming"):
		return "Subscriptions"
	case strings.Contains(combined, "entertain") || strings.Contains(combined, "recreation") ||
		strings.Contains(combined, "sport") || strings.Contains(combined, "music") ||
		strings.Contains(combined, "movie") || strings.Contains(combined, "game"):
		return "Entertainment"
	case strings.Contains(combined, "health") || strings.Contains(combined, "medical") ||
		strings.Contains(combined, "pharmacy") || strings.Contains(combined, "gym") ||
		strings.Contains(combined, "fitness") || strings.Contains(combined, "doctor") ||
		strings.Contains(combined, "dental"):
		return "Health"
	case strings.Contains(combined, "income") || strings.Contains(combined, "payroll") ||
		strings.Contains(combined, "deposit"):
		return "Income"
	case strings.Contains(combined, "transfer") || strings.Contains(combined, "payment"):
		// Plaid's own category is too coarse to tell an internal transfer
		// between the user's own accounts (or a debt payment) from real
		// income or real discretionary spending, and it's used for both
		// inflows (e.g. moving money into a CD) and outflows (e.g. an
		// automatic loan payment). "transfer" used to be grouped under
		// Income, which mislabeled outflow-direction transfers as paychecks;
		// giving it its own category — excluded from spending totals — is
		// more honest than guessing either way.
		return "Transfers"
	default:
		return "Other"
	}
}

// isTransferLike reports whether Plaid's raw category mentions transfer or
// payment activity anywhere in it, independent of NormalizeCategory's
// first-match-wins switch. Used when an income-shaped category (e.g.
// "Transfer, Deposit") fails the inflow sign check: since we already know
// from the raw category that this was transfer-shaped, preferring Transfers
// over a generic merchant-based guess (or "Other") is more accurate — a CD
// funding transfer categorized as "Other" would still count as spending,
// which is exactly what Transfers exists to avoid.
func isTransferLike(plaidCategories []string) bool {
	combined := strings.ToLower(strings.Join(plaidCategories, " "))
	return strings.Contains(combined, "transfer") || strings.Contains(combined, "payment")
}
