package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
)

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

		if txns == nil {
			txns = []models.Transaction{}
		}

		writeJSON(w, http.StatusOK, txns)
	}
}
