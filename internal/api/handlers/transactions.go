package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
	"github.com/yourname/pocket-api/internal/money"
)

// transactionResponse is the camelCase wire contract the frontend consumes.
// Amount is flipped from the internal "positive = expense" ledger convention
// to the standard "negative = debit" convention the UI displays everywhere.
type transactionResponse struct {
	ID        string  `json:"id"`
	AccountID string  `json:"accountId"`
	Date      string  `json:"date"`
	Merchant  string  `json:"merchant"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Notes     string  `json:"notes,omitempty"`
}

// transactionsListResponse wraps the page of transactions with the true
// total count, so the frontend can paginate instead of silently truncating
// at the page size with no indication more rows exist.
type transactionsListResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

func toTransactionResponse(t models.Transaction) transactionResponse {
	merchant := t.Name
	if t.MerchantName != nil && *t.MerchantName != "" {
		merchant = *t.MerchantName
	}

	notes := ""
	if merchant != t.Name {
		notes = t.Name
	}

	return transactionResponse{
		ID:        t.ID,
		AccountID: t.AccountID,
		Date:      t.Date.Format("2006-01-02"),
		Merchant:  merchant,
		Category:  t.Category,
		Amount:    money.Round2(-t.Amount),
		Notes:     notes,
	}
}

func GetTransactions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		filters := queries.TransactionFilters{
			Limit:  50,
			Offset: 0,
		}

		q := r.URL.Query()
		filters.Search = q.Get("search")
		filters.Category = q.Get("category")
		filters.StartDate = q.Get("start")
		filters.EndDate = q.Get("end")
		if v := q.Get("accountId"); v != "" {
			filters.AccountID = v
		}

		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				filters.Limit = n
			}
		}
		if v := q.Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				filters.Offset = n
			}
		}

		txns, err := queries.GetTransactionsByUserID(db, userID, filters)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		total, err := queries.CountTransactionsByUserID(db, userID, filters)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		resp := make([]transactionResponse, len(txns))
		for i, t := range txns {
			resp[i] = toTransactionResponse(t)
		}

		writeJSON(w, http.StatusOK, transactionsListResponse{
			Transactions: resp,
			Total:        total,
			Limit:        filters.Limit,
			Offset:       filters.Offset,
		})
	}
}
