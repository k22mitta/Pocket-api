package handlers

import (
	"database/sql"
	"net/http"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
)

func GetAccounts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		accounts, err := queries.GetAccountsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if accounts == nil {
			accounts = []models.Account{}
		}

		writeJSON(w, http.StatusOK, accounts)
	}
}
