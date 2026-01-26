package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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

		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}
