package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/plaid/plaid-go/v29/plaid"
	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/db/queries"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
)

func CreateLinkToken(pc *plaidclient.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		req := plaid.NewLinkTokenCreateRequest(
			"Pocket",
			"en",
			[]plaid.CountryCode{plaid.COUNTRYCODE_US, plaid.COUNTRYCODE_CA},
			*plaid.NewLinkTokenCreateRequestUser(userID),
		)
		req.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

		resp, _, err := pc.API().PlaidApi.LinkTokenCreate(context.Background()).LinkTokenCreateRequest(*req).Execute()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create link token"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"link_token": resp.GetLinkToken()})
	}
}

func ExchangeToken(pc *plaidclient.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		var body struct {
			PublicToken     string `json:"public_token"`
			InstitutionID   string `json:"institution_id"`
			InstitutionName string `json:"institution_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if body.PublicToken == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "public_token is required"})
			return
		}

		resp, _, err := pc.API().PlaidApi.ItemPublicTokenExchange(context.Background()).
			ItemPublicTokenExchangeRequest(*plaid.NewItemPublicTokenExchangeRequest(body.PublicToken)).
			Execute()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if err := queries.CreatePlaidItem(db, userID, resp.GetAccessToken(), resp.GetItemId(), body.InstitutionID, body.InstitutionName); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		itemID := resp.GetItemId()
		go func() {
			ctx := context.Background()
			item, err := queries.GetPlaidItemByItemID(db, itemID)
			if err != nil {
				log.Printf("sync: failed to fetch item %s: %v", itemID, err)
				return
			}
			if err := plaidclient.SyncAccounts(ctx, pc, db, userID, item); err != nil {
				log.Printf("sync: SyncAccounts failed for item %s: %v", itemID, err)
				return
			}
			accountsMap, err := buildAccountsMap(db, userID)
			if err != nil {
				log.Printf("sync: failed to build accounts map for user %s: %v", userID, err)
				return
			}
			if err := plaidclient.SyncTransactions(ctx, pc, db, userID, item, accountsMap); err != nil {
				log.Printf("sync: SyncTransactions failed for item %s: %v", itemID, err)
			}
		}()

		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}

func SyncAll(pc *plaidclient.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		items, err := queries.GetPlaidItemsByUserID(db, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		ctx := context.Background()
		synced := 0

		for _, item := range items {
			if err := plaidclient.SyncAccounts(ctx, pc, db, userID, item); err != nil {
				log.Printf("sync: SyncAccounts failed for item %s: %v", item.ItemID, err)
				continue
			}
			accountsMap, err := buildAccountsMap(db, userID)
			if err != nil {
				log.Printf("sync: failed to build accounts map for user %s: %v", userID, err)
				continue
			}
			if err := plaidclient.SyncTransactions(ctx, pc, db, userID, item, accountsMap); err != nil {
				log.Printf("sync: SyncTransactions failed for item %s: %v", item.ItemID, err)
				continue
			}
			synced++
		}

		writeJSON(w, http.StatusOK, map[string]int{"synced": synced})
	}
}

// DeletePlaidItem unlinks a bank connection: it revokes the item with Plaid
// (best-effort — a revoke failure doesn't block the local delete, since the
// user should still be able to remove a stale or already-revoked connection)
// then deletes the plaid_items row, which cascades to every account and
// transaction under it. {id} is the plaid_items row ID, not an account ID —
// one connection can carry several accounts (e.g. checking + savings from
// the same bank), and unlinking removes all of them together.
func DeletePlaidItem(pc *plaidclient.Client, db *sql.DB) http.HandlerFunc {
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

func buildAccountsMap(db *sql.DB, userID string) (map[string]string, error) {
	accounts, err := queries.GetAccountsByUserID(db, userID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(accounts))
	for _, a := range accounts {
		m[a.PlaidAccountID] = a.ID
	}
	return m, nil
}
