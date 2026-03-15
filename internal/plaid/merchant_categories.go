package plaidclient

import "strings"

// CategorizeByMerchant infers a canonical category from a merchant/description
// string when a CSV row has no category column. Groceries/Dining and
// Subscriptions/Entertainment are each kept distinct (see NormalizeCategory)
// so a "Groceries" budget never silently absorbs restaurant spend, and a
// "Subscriptions" budget tracks recurring auto-pay separately from one-off
// entertainment purchases.
func CategorizeByMerchant(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "whole foods") || strings.Contains(lower, "trader joe") ||
		strings.Contains(lower, "safeway") || strings.Contains(lower, "kroger") ||
		strings.Contains(lower, "grocery") || strings.Contains(lower, "aldi") ||
		strings.Contains(lower, "costco") || strings.Contains(lower, "supermarket"):
		return "Groceries"
	case strings.Contains(lower, "starbucks") || strings.Contains(lower, "chipotle") ||
		strings.Contains(lower, "mcdonald") || strings.Contains(lower, "uber eats") ||
		strings.Contains(lower, "doordash") || strings.Contains(lower, "grubhub") ||
		strings.Contains(lower, "restaurant") || strings.Contains(lower, "cafe") ||
		strings.Contains(lower, "coffee") || strings.Contains(lower, "pizza") ||
		strings.Contains(lower, "taco") || strings.Contains(lower, "burger") ||
		strings.Contains(lower, "kitchen"):
		return "Dining"
	case strings.Contains(lower, "amazon") || strings.Contains(lower, "target") ||
		strings.Contains(lower, "walmart") || strings.Contains(lower, "nike") ||
		strings.Contains(lower, "best buy") || strings.Contains(lower, "ebay") ||
		strings.Contains(lower, "etsy") || strings.Contains(lower, "store") ||
		strings.Contains(lower, "shop") || strings.Contains(lower, "mall") ||
		strings.Contains(lower, "clothing") || strings.Contains(lower, "apple store"):
		return "Shopping"
	case strings.Contains(lower, "united") || strings.Contains(lower, "delta") ||
		strings.Contains(lower, "american airlines") || strings.Contains(lower, "airline") ||
		strings.Contains(lower, "airways") || strings.Contains(lower, "hotel") ||
		strings.Contains(lower, "airbnb") || strings.Contains(lower, "expedia"):
		return "Travel"
	case strings.Contains(lower, "uber") || strings.Contains(lower, "lyft") ||
		strings.Contains(lower, "shell") || strings.Contains(lower, "chevron") ||
		strings.Contains(lower, "exxon") || strings.Contains(lower, "gas") ||
		strings.Contains(lower, "transit") || strings.Contains(lower, "parking") ||
		strings.Contains(lower, "taxi") || strings.Contains(lower, "fuel"):
		return "Transport"
	case strings.Contains(lower, "netflix") || strings.Contains(lower, "spotify") ||
		strings.Contains(lower, "hulu") || strings.Contains(lower, "disney+") ||
		strings.Contains(lower, "disney plus") || strings.Contains(lower, "apple music") ||
		strings.Contains(lower, "apple tv") || strings.Contains(lower, "youtube premium") ||
		strings.Contains(lower, "hbo") || strings.Contains(lower, "paramount+") ||
		strings.Contains(lower, "subscription"):
		return "Subscriptions"
	case strings.Contains(lower, "disney") || strings.Contains(lower, "amc") || strings.Contains(lower, "cinema") ||
		strings.Contains(lower, "theater") || strings.Contains(lower, "youtube") ||
		strings.Contains(lower, "twitch") || strings.Contains(lower, "steam") ||
		strings.Contains(lower, "playstation") || strings.Contains(lower, "xbox") ||
		strings.Contains(lower, "game"):
		return "Entertainment"
	case strings.Contains(lower, "cvs") || strings.Contains(lower, "walgreens") ||
		strings.Contains(lower, "pharmacy") || strings.Contains(lower, "gym") ||
		strings.Contains(lower, "fitness") || strings.Contains(lower, "doctor") ||
		strings.Contains(lower, "medical") || strings.Contains(lower, "health") ||
		strings.Contains(lower, "dental") || strings.Contains(lower, "hospital") ||
		strings.Contains(lower, "clinic"):
		return "Health"
	case strings.Contains(lower, "rent") || strings.Contains(lower, "mortgage") ||
		strings.Contains(lower, "landlord") || strings.Contains(lower, "property") ||
		strings.Contains(lower, "lease") || strings.Contains(lower, "apartment") ||
		strings.Contains(lower, "hoa"):
		return "Housing"
	case strings.Contains(lower, "payroll") || strings.Contains(lower, "deposit") ||
		strings.Contains(lower, "salary") || strings.Contains(lower, "direct dep") ||
		strings.Contains(lower, "refund") || strings.Contains(lower, "paycheck"):
		return "Income"
	default:
		return "Other"
	}
}
