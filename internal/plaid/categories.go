package plaidclient

import "strings"

func NormalizeCategory(plaidCategories []string) string {
	if len(plaidCategories) == 0 {
		return "Other"
	}

	combined := strings.ToLower(strings.Join(plaidCategories, " "))

	switch {
	case strings.Contains(combined, "food") || strings.Contains(combined, "drink") ||
		strings.Contains(combined, "restaurant") || strings.Contains(combined, "coffee") ||
		strings.Contains(combined, "dining") || strings.Contains(combined, "grocery") ||
		strings.Contains(combined, "supermarket"):
		return "Food & Drink"
	case strings.Contains(combined, "shop") || strings.Contains(combined, "retail") ||
		strings.Contains(combined, "store") || strings.Contains(combined, "merchandise"):
		return "Shopping"
	case strings.Contains(combined, "transport") || strings.Contains(combined, "travel") ||
		strings.Contains(combined, "taxi") || strings.Contains(combined, "airline") ||
		strings.Contains(combined, "gas station") || strings.Contains(combined, "fuel") ||
		strings.Contains(combined, "parking") || strings.Contains(combined, "toll"):
		return "Transport"
	case strings.Contains(combined, "housing") || strings.Contains(combined, "rent") ||
		strings.Contains(combined, "utilities") || strings.Contains(combined, "mortgage"):
		return "Housing"
	case strings.Contains(combined, "entertain") || strings.Contains(combined, "recreation") ||
		strings.Contains(combined, "sport") || strings.Contains(combined, "music") ||
		strings.Contains(combined, "movie") || strings.Contains(combined, "game") ||
		strings.Contains(combined, "streaming"):
		return "Entertainment"
	case strings.Contains(combined, "health") || strings.Contains(combined, "medical") ||
		strings.Contains(combined, "pharmacy") || strings.Contains(combined, "gym") ||
		strings.Contains(combined, "fitness") || strings.Contains(combined, "doctor") ||
		strings.Contains(combined, "dental"):
		return "Health"
	case strings.Contains(combined, "income") || strings.Contains(combined, "payroll") ||
		strings.Contains(combined, "deposit") || strings.Contains(combined, "transfer"):
		return "Income"
	default:
		return "Other"
	}
}
