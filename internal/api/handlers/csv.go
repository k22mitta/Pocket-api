package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/k22mitta/pocket-api/internal/api/middleware"
	"github.com/k22mitta/pocket-api/internal/db/queries"
	"github.com/k22mitta/pocket-api/internal/models"
	"github.com/k22mitta/pocket-api/internal/money"
	plaidclient "github.com/k22mitta/pocket-api/internal/plaid"
)

var columnAliases = map[string][]string{
	"date":        {"date", "transaction date", "posted date", "posting date", "trans date"},
	"description": {"description", "details", "memo", "payee", "narrative", "name", "transaction", "reference", "merchant", "vendor"},
	"amount":      {"amount", "value", "transaction amount"},
	"debit":       {"debit", "withdrawal", "money out", "paid out"},
	"credit":      {"credit", "deposit", "money in", "paid in"},
	"category":    {"category", "type"},
}

// columnOrder fixes the priority used to resolve a header to a column.
// Iterating a map (as this used to) randomizes order in Go, which made
// column resolution nondeterministic for any header matching more than one
// alias — e.g. "Transaction Date" contains both the "date" alias
// "transaction date" and the "description" alias "transaction", so which
// column won depended on map iteration order and could differ between runs.
// Checking exact matches (in this fixed order) before any substring match
// eliminates that: "Transaction Date" normalizes to an exact match on
// "date" before "description"'s substring check ever runs.
var columnOrder = []string{"date", "description", "amount", "debit", "credit", "category"}

func matchColumn(normalized string) string {
	for _, col := range columnOrder {
		for _, alias := range columnAliases[col] {
			if normalized == alias {
				return col
			}
		}
	}
	for _, col := range columnOrder {
		for _, alias := range columnAliases[col] {
			if strings.Contains(normalized, alias) {
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

		file, fileHeader, err := r.FormFile("file")
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

		type parsedRow struct {
			date     time.Time
			desc     string
			category string
			amount   float64
		}

		var parsed []parsedRow
		var errors []string

		for i, row := range records[1:] {
			// Reported as "entry N" (1-indexed among data rows, not counting the
			// header), so "entry 10" means the user's 10th transaction — matching
			// how someone reviewing their own statement would count, rather than
			// the file's raw line number (which would be 11, since line 1 is the
			// header, and reads as an off-by-one bug to anyone not counting it).
			entryNum := i + 1

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
				errors = append(errors, fmt.Sprintf("entry %d: invalid date %q", entryNum, dateStr))
				continue
			}

			// Internal ledger convention is positive = expense, negative = income
			// (matches Plaid's own sign convention). A single "amount" column in a
			// user-uploaded CSV follows the opposite, standard bank-statement
			// convention (negative = debit, positive = deposit), so it's inverted
			// here. Debit/credit columns are unambiguous and already normalized below.
			var amount float64
			if hasAmount {
				amount, err = parseAmount(safeGet(colIdx["amount"]))
				if err != nil {
					errors = append(errors, fmt.Sprintf("entry %d: invalid amount %q", entryNum, safeGet(colIdx["amount"])))
					continue
				}
				amount = -amount
			} else {
				debitStr := safeGet(colIdx["debit"])
				creditStr := safeGet(colIdx["credit"])
				if debitStr != "" {
					amount, err = parseAmount(debitStr)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid debit %q", entryNum, debitStr))
						continue
					}
				} else if creditStr != "" {
					amount, err = parseAmount(creditStr)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid credit %q", entryNum, creditStr))
						continue
					}
					amount = -amount
				} else {
					errors = append(errors, fmt.Sprintf("entry %d: no amount, debit, or credit value", entryNum))
					continue
				}
			}

			parsed = append(parsed, parsedRow{date: date, desc: desc, category: category, amount: amount})
		}

		var accountID string
		isNewAccount := len(accounts) == 0
		if isNewAccount {
			accountID, err = queries.CreateSeedAccount(
				db, userID, "csv-import", "Imported",
				importedAccountName(fileHeader.Filename), "depository", "checking", 0,
			)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
				return
			}
		} else {
			accountID = accounts[0].ID
		}

		// plaid_transaction_id is derived from row content (account + date +
		// description + amount + how many times this exact combo has already
		// been seen in this file) rather than a wall-clock timestamp. Re-uploading
		// the identical CSV then produces the identical IDs, so the existing
		// ON CONFLICT DO NOTHING dedupes it instead of silently doubling every
		// transaction (and the account balance) on every re-import. Genuine
		// same-day duplicate purchases (two identical $6 coffees) still get
		// distinct IDs because their occurrence counter differs.
		created := 0
		netChange := 0.0
		occurrences := map[string]int{}

		for i, p := range parsed {
			key := fmt.Sprintf("%s|%s|%.2f", p.date.Format("2006-01-02"), p.desc, p.amount)
			occurrence := occurrences[key]
			occurrences[key] = occurrence + 1

			plaidTxID := fmt.Sprintf("csv-%s-%s-%d", accountID, key, occurrence)
			merchant := CleanMerchantName(p.desc)
			t := models.Transaction{
				UserID:             userID,
				AccountID:          accountID,
				PlaidTransactionID: plaidTxID,
				Name:               p.desc,
				MerchantName:       &merchant,
				Amount:             p.amount,
				Category:           p.category,
				Date:               p.date,
				Pending:            false,
			}

			inserted, err := queries.UpsertTransaction(db, t)
			if err != nil {
				errors = append(errors, fmt.Sprintf("entry %d: failed to insert", i+1))
				continue
			}
			if inserted {
				created++
				netChange -= p.amount
			}
		}
		netChange = money.Round2(netChange)

		if netChange != 0 {
			if err := queries.AdjustAccountBalance(db, accountID, netChange); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update account balance"})
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success":              true,
			"transactions_created": created,
			"errors":               errors,
		})
	}
}

// importedAccountName derives a human-friendly account name from the
// uploaded file name, e.g. "chase_checking_export.csv" -> "Chase Checking Export".
func importedAccountName(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	base = strings.NewReplacer("_", " ", "-", " ").Replace(base)
	base = strings.TrimSpace(base)
	if base == "" {
		return "Imported Statement"
	}
	words := strings.Fields(base)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
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
	// Matches a trailing US state / Canadian province code exactly (e.g. " WA",
	// " ON"). Must be exactly 2 letters, not 2-or-more: since CleanMerchantName
	// upper-cases the whole string first, "{2,}" would match the last word of
	// ANY merchant name (e.g. "APARTMENTS", "CORP") and silently truncate it.
	reTrailingCityProvince = regexp.MustCompile(`(?i)\s+[A-Z]{2}\s*$`)
	reSymbolsAndNumbers    = regexp.MustCompile(`[#,\-]|\b\d+\b`)
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

// normalizeCategory maps a raw CSV category string onto Pocket's canonical
// taxonomy (Groceries, Dining, Shopping, Transport, Travel, Subscriptions,
// Entertainment, Health, Housing, Income, Transfers, Other) so budgets and
// filters line up across CSV import, Plaid sync, and demo data.
// Groceries/Dining and Subscriptions/Entertainment are each kept separate
// rather than merged: recurring auto-pay (Netflix, Spotify) is exactly the
// spend people most want to audit on its own, and a "Groceries" budget must
// not silently include restaurant/coffee spend. Transfers (moving money
// between the user's own accounts, paying down a loan/credit card) is kept
// out of Income and out of spending totals, since it's neither earned money
// nor discretionary spend. Unrecognized categories keep the user's own label
// (title-cased) instead of collapsing to "Other", since that's a more useful
// signal than a generic bucket.
func normalizeCategory(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "grocer") || strings.Contains(lower, "supermarket"):
		return "Groceries"
	case strings.Contains(lower, "food") || strings.Contains(lower, "dining") || strings.Contains(lower, "drink") ||
		strings.Contains(lower, "restaurant") || strings.Contains(lower, "coffee") || strings.Contains(lower, "cafe"):
		return "Dining"
	case strings.Contains(lower, "shop") || strings.Contains(lower, "retail") || strings.Contains(lower, "merchandise"):
		return "Shopping"
	case strings.Contains(lower, "travel") || strings.Contains(lower, "airline") || strings.Contains(lower, "flight") ||
		strings.Contains(lower, "hotel") || strings.Contains(lower, "lodging"):
		return "Travel"
	case strings.Contains(lower, "transport") || strings.Contains(lower, "gas") ||
		strings.Contains(lower, "uber") || strings.Contains(lower, "lyft") || strings.Contains(lower, "transit") ||
		strings.Contains(lower, "parking") || strings.Contains(lower, "fuel"):
		return "Transport"
	case strings.Contains(lower, "subscription") || strings.Contains(lower, "streaming"):
		return "Subscriptions"
	case strings.Contains(lower, "entertainment") || strings.Contains(lower, "fun") || strings.Contains(lower, "movie") ||
		strings.Contains(lower, "music"):
		return "Entertainment"
	case strings.Contains(lower, "health") || strings.Contains(lower, "medical") || strings.Contains(lower, "pharmacy") ||
		strings.Contains(lower, "gym") || strings.Contains(lower, "fitness"):
		return "Health"
	case strings.Contains(lower, "housing") || strings.Contains(lower, "rent") || strings.Contains(lower, "mortgage") ||
		strings.Contains(lower, "utilit"):
		return "Housing"
	case strings.Contains(lower, "income") || strings.Contains(lower, "payroll") || strings.Contains(lower, "salary") ||
		strings.Contains(lower, "deposit"):
		return "Income"
	case strings.Contains(lower, "transfer") || strings.Contains(lower, "payment"):
		// Matches the Plaid sync mapping: internal account transfers and
		// debt payments aren't income or discretionary spending — see
		// plaidclient.NormalizeCategory for the full reasoning.
		return "Transfers"
	default:
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return "Other"
		}
		words := strings.Fields(strings.ToLower(trimmed))
		for i, w := range words {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
		return strings.Join(words, " ")
	}
}
