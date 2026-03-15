package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
	"github.com/yourname/pocket-api/internal/money"
)

// accountResponse is the camelCase wire contract the frontend consumes.
// PlaidItemID and IsPlaidLinked let the UI offer an "unlink" action only on
// accounts backed by a real Plaid connection (not CSV imports or the demo
// seed, which share synthetic plaid_items with no access token to revoke).
type accountResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Institution   string  `json:"institution"`
	Balance       float64 `json:"balance"`
	LastSynced    string  `json:"lastSynced,omitempty"`
	PlaidItemID   string  `json:"plaidItemId"`
	IsPlaidLinked bool    `json:"isPlaidLinked"`
}

// isLiabilityType reports whether Plaid's account type represents money owed
// (credit cards, loans, mortgages, student loans) rather than money held.
// Plaid reports the balance on these as a positive "amount owed", but they
// are negative equity and must subtract from net position everywhere net
// position is computed — this same check is mirrored in
// queries.GetTotalBalance's SQL so /accounts and /summary/balance never
// disagree about which accounts are liabilities.
func isLiabilityType(plaidType string) bool {
	return plaidType == "credit" || plaidType == "loan"
}

// toAccountResponse derives the frontend's checking/savings/credit/loan/
// investment enum from Plaid's type+subtype, and flips liability balances
// negative since Plaid reports the amount owed as positive but the UI treats
// debt as negative equity.
func toAccountResponse(a models.Account) accountResponse {
	balance := 0.0
	if a.CurrentBalance != nil {
		balance = *a.CurrentBalance
	}

	subtype := ""
	if a.Subtype != nil {
		subtype = *a.Subtype
	}

	feType := "checking"
	switch {
	case a.Type == "loan":
		feType = "loan"
	case a.Type == "credit":
		feType = "credit"
	case a.Type == "investment" || a.Type == "brokerage" || subtype == "investment" || subtype == "brokerage":
		feType = "investment"
	case subtype == "savings" || subtype == "cd" || subtype == "money market":
		feType = "savings"
	default:
		feType = "checking"
	}

	if isLiabilityType(a.Type) {
		balance = -balance
	}

	institution := "Pocket"
	if a.InstitutionName != nil && *a.InstitutionName != "" {
		institution = *a.InstitutionName
	}

	return accountResponse{
		ID:            a.ID,
		Name:          a.Name,
		Type:          feType,
		Institution:   institution,
		Balance:       money.Round2(balance),
		LastSynced:    a.UpdatedAt.UTC().Format(time.RFC3339),
		PlaidItemID:   a.PlaidItemID,
		IsPlaidLinked: a.IsPlaidLinked,
	}
}

func GetAccounts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		accounts, err := queries.GetAccountsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		resp := make([]accountResponse, len(accounts))
		for i, a := range accounts {
			resp[i] = toAccountResponse(a)
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
