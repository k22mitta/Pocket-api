package plaidclient

import "strings"

func NormalizeCategory(plaidCategories []string) string {
	if len(plaidCategories) == 0 {
		return "Other"
	}

	first := strings.ToLower(plaidCategories[0])

	switch {
	case strings.Contains(first, "food") || strings.Contains(first, "drink") || strings.Contains(first, "restaurant"):
		return "Food & Drink"
	case strings.Contains(first, "shop") || strings.Contains(first, "retail") || strings.Contains(first, "store"):
		return "Shopping"
	case strings.Contains(first, "transport") || strings.Contains(first, "travel") || strings.Contains(first, "taxi") || strings.Contains(first, "airline"):
		return "Transport"
	case strings.Contains(first, "housing") || strings.Contains(first, "rent") || strings.Contains(first, "utilities") || strings.Contains(first, "mortgage"):
		return "Housing"
	case strings.Contains(first, "entertain") || strings.Contains(first, "recreation") || strings.Contains(first, "sport"):
		return "Entertainment"
	case strings.Contains(first, "health") || strings.Contains(first, "medical") || strings.Contains(first, "pharmacy"):
		return "Health"
	case strings.Contains(first, "income") || strings.Contains(first, "payroll") || strings.Contains(first, "transfer"):
		return "Income"
	default:
		return "Other"
	}
}
