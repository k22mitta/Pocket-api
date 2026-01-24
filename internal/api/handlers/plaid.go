package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/plaid/plaid-go/v29/plaid"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
	"github.com/yourname/pocket-api/internal/api/middleware"
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
