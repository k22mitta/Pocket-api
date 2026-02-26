package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
)

var columnAliases = map[string][]string{
	"date":        {"date", "transaction date", "posted date", "posting date", "trans date"},
	"description": {"description", "details", "memo", "payee", "narrative", "name", "transaction", "reference"},
	"amount":      {"amount", "value", "transaction amount"},
	"debit":       {"debit", "withdrawal", "money out", "paid out"},
	"credit":      {"credit", "deposit", "money in", "paid in"},
	"category":    {"category", "type"},
}

func matchColumn(normalized string) string {
	for col, aliases := range columnAliases {
		for _, alias := range aliases {
			if normalized == alias || strings.Contains(normalized, alias) {
				return col
			}
		}
	}
	return ""
}

func resolveColumns(header []string) map[string]int {
	idx := map[string]int{
		"date": -1, "description": -1, "amount": -1,
		"debit": -1, "credit": -1, "category": -1,
	}
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		if col := matchColumn(normalized); col != "" {
			if idx[col] == -1 {
				idx[col] = i
			}
		}
	}
	return idx
}

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer("$", "", "£", "", "€", "", ",", "", " ", "").Replace(s)
	return strconv.ParseFloat(s, 64)
}

func ImportCSV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file field"})
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to parse CSV"})
			return
		}

		if len(records) < 2 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "CSV has no data rows"})
			return
		}

		colIdx := resolveColumns(records[0])

		hasAmount := colIdx["amount"] != -1
		hasDebitCredit := colIdx["debit"] != -1 || colIdx["credit"] != -1
		if colIdx["date"] == -1 || colIdx["description"] == -1 || (!hasAmount && !hasDebitCredit) {
			detected := make([]string, len(records[0]))
			for i, h := range records[0] {
				detected[i] = strings.TrimSpace(h)
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("could not find required columns (date, description, amount or debit/credit) in detected headers: %s", strings.Join(detected, ", ")),
			})
			return
		}

		accounts, err := queries.GetAccountsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		var accountID string
		if len(accounts) == 0 {
			accountID, err = queries.CreateDemoAccount(db, userID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
				return
			}
		} else {
			accountID = accounts[0].ID
		}

		ts := fmt.Sprintf("%d", time.Now().UnixNano())
		created := 0
		var errors []string

		for i, row := range records[1:] {
			rowNum := i + 2

			safeGet := func(idx int) string {
				if idx == -1 || idx >= len(row) {
					return ""
				}
				return strings.TrimSpace(row[idx])
			}

			dateStr := safeGet(colIdx["date"])
			desc := safeGet(colIdx["description"])

			var category string
			if raw := safeGet(colIdx["category"]); raw != "" {
				category = normalizeCategory(raw)
			} else {
				category = plaidclient.CategorizeByMerchant(desc)
			}

			date, parseErr := parseDate(dateStr)
			if parseErr != nil {
				errors = append(errors, fmt.Sprintf("row %d: invalid date %q", rowNum, dateStr))
				continue
			}

			var amount float64
			if hasAmount {
				amount, err = parseAmount(safeGet(colIdx["amount"]))
				if err != nil {
					errors = append(errors, fmt.Sprintf("row %d: invalid amount %q", rowNum, safeGet(colIdx["amount"])))
					continue
				}
			} else {
				debitStr := safeGet(colIdx["debit"])
				creditStr := safeGet(colIdx["credit"])
				if debitStr != "" {
					amount, err = parseAmount(debitStr)
					if err != nil {
						errors = append(errors, fmt.Sprintf("row %d: invalid debit %q", rowNum, debitStr))
						continue
					}
				} else if creditStr != "" {
					amount, err = parseAmount(creditStr)
					if err != nil {
						errors = append(errors, fmt.Sprintf("row %d: invalid credit %q", rowNum, creditStr))
						continue
					}
					amount = -amount
				} else {
					errors = append(errors, fmt.Sprintf("row %d: no amount, debit, or credit value", rowNum))
					continue
				}
			}

			plaidTxID := fmt.Sprintf("csv-%s-%d-%s", userID, i, ts)
			merchant := CleanMerchantName(desc)
			t := models.Transaction{
				UserID:             userID,
				AccountID:          accountID,
				PlaidTransactionID: plaidTxID,
				Name:               desc,
				MerchantName:       &merchant,
				Amount:             amount,
				Category:           category,
				Date:               date,
				Pending:            false,
			}

			if err := queries.UpsertTransaction(db, t); err != nil {
				errors = append(errors, fmt.Sprintf("row %d: failed to insert", rowNum))
				continue
			}
			created++
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success":              true,
			"transactions_created": created,
			"errors":               errors,
		})
	}
}

var dateLayouts = []string{
	"2006-01-02",
	"01/02/2006",
	"02/01/2006",
	"02 Jan 2006",
	"2 Jan 2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"02-Jan-2006",
	"2006/01/02",
	"01-02-2006",
	"02.01.2006",
	"2006-01-02T15:04:05Z07:00",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %q", s)
}

var processorPrefixes = []string{
	"INTERAC ", "VISA ", "PRE-AUTH ", "PURCHASE ", "DEBIT ", "POS ",
	"PAYPAL ", "PP*", "TST* ", "TST*", "SQ ", "SQ", "DD ", "DD", "BT",
}

var knownBrands = []struct {
	substr string
	name   string
}{
	{"DOORDASH", "DoorDash"},
	{"UBER EATS", "Uber Eats"},
	{"UBER", "Uber"},
	{"LYFT", "Lyft"},
	{"AMAZON", "Amazon"},
	{"STARBUCKS", "Starbucks"},
	{"MCDONAL", "McDonald's"},
	{"DOMINOS", "Domino's"},
	{"CHIPOTLE", "Chipotle"},
	{"ALDI", "Aldi"},
	{"WALMART", "Walmart"},
	{"TARGET", "Target"},
	{"COSTCO", "Costco"},
	{"NETFLIX", "Netflix"},
	{"SPOTIFY", "Spotify"},
	{"APPLE", "Apple"},
	{"SHELL", "Shell"},
	{"ESSO", "Esso"},
	{"PETRO", "Petro-Canada"},
	{"TIM HORTONS", "Tim Hortons"},
	{"FAMOSO", "Famoso"},
	{"PRESTO", "Presto"},
	{"METROLINX", "Metrolinx"},
}

var (
	reTrailingCityProvince = regexp.MustCompile(`(?i)\s+[A-Z]{2,}\s*$`)
	reSymbolsAndNumbers    = regexp.MustCompile(`[#,]|\b\d+\b`)
	reExtraSpace           = regexp.MustCompile(`\s{2,}`)
)

func CleanMerchantName(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return raw
	}

	up := strings.ToUpper(s)

	for {
		stripped := false
		for _, p := range processorPrefixes {
			if strings.HasPrefix(up, p) {
				up = strings.TrimPrefix(up, p)
				stripped = true
				break
			}
		}
		if !stripped {
			break
		}
	}

	for _, b := range knownBrands {
		if strings.Contains(up, b.substr) {
			return b.name
		}
	}

	clean := reTrailingCityProvince.ReplaceAllString(up, "")
	clean = reSymbolsAndNumbers.ReplaceAllString(clean, " ")
	clean = reExtraSpace.ReplaceAllString(clean, " ")
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return strings.TrimSpace(raw)
	}

	words := strings.Fields(clean)
	if len(words) > 3 {
		words = words[:3]
	}

	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}

	result := strings.Join(words, " ")
	if result == "" {
		return strings.TrimSpace(raw)
	}
	return result
}

func normalizeCategory(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "food") || strings.Contains(lower, "dining") || strings.Contains(lower, "drink") || strings.Contains(lower, "restaurant"):
		return "Food & Drink"
	case strings.Contains(lower, "shop") || strings.Contains(lower, "retail") || strings.Contains(lower, "merchandise"):
		return "Shopping"
	case strings.Contains(lower, "transport") || strings.Contains(lower, "travel") || strings.Contains(lower, "gas") || strings.Contains(lower, "uber") || strings.Contains(lower, "lyft"):
		return "Transport"
	case strings.Contains(lower, "entertainment") || strings.Contains(lower, "fun") || strings.Contains(lower, "movie") || strings.Contains(lower, "music"):
		return "Entertainment"
	case strings.Contains(lower, "health") || strings.Contains(lower, "medical") || strings.Contains(lower, "pharmacy") || strings.Contains(lower, "gym"):
		return "Health"
	case strings.Contains(lower, "housing") || strings.Contains(lower, "rent") || strings.Contains(lower, "mortgage"):
		return "Housing"
	case strings.Contains(lower, "income") || strings.Contains(lower, "payroll") || strings.Contains(lower, "salary") || strings.Contains(lower, "deposit"):
		return "Income"
	default:
		return "Other"
	}
}
