package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
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

func DeleteAccount(pc *plaidclient.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		itemID := r.PathValue("id")

		item, err := queries.GetPlaidItemByID(db, itemID, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if err := plaidclient.RemoveItem(context.Background(), pc, item.AccessToken); err != nil {
			log.Printf("plaid: failed to revoke item %s: %v", item.ItemID, err)
		}

		if err := queries.DeletePlaidItem(db, itemID, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}
