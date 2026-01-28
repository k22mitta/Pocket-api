package plaidclient

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/plaid/plaid-go/v29/plaid"
	"github.com/yourname/pocket-api/internal/db/queries"
	"github.com/yourname/pocket-api/internal/models"
)

func SyncAccounts(ctx context.Context, client *Client, db *sql.DB, userID string, item models.PlaidItem) error {
	resp, _, err := client.api.PlaidApi.AccountsGet(ctx).
		AccountsGetRequest(*plaid.NewAccountsGetRequest(item.AccessToken)).
		Execute()
	if err != nil {
		return fmt.Errorf("fetching accounts: %w", err)
	}

	for _, pa := range resp.GetAccounts() {
		balances := pa.GetBalances()

		a := models.Account{
			PlaidItemID:    item.ID,
			UserID:         userID,
			PlaidAccountID: pa.GetAccountId(),
			Name:           pa.GetName(),
			Type:           string(pa.GetType()),
			CurrencyCode:   "USD",
		}

		if name, ok := pa.GetOfficialNameOk(); ok && name != nil {
			s := *name
			a.OfficialName = &s
		}

		if subtype, ok := pa.GetSubtypeOk(); ok && subtype != nil {
			s := string(*subtype)
			a.Subtype = &s
		}

		if current, ok := balances.GetCurrentOk(); ok && current != nil {
			v := *current
			a.CurrentBalance = &v
		}

		if available, ok := balances.GetAvailableOk(); ok && available != nil {
			v := *available
			a.AvailableBalance = &v
		}

		if iso, ok := balances.GetIsoCurrencyCodeOk(); ok && iso != nil && *iso != "" {
			a.CurrencyCode = *iso
		}

		if err := queries.UpsertAccount(db, a); err != nil {
			return fmt.Errorf("upserting account %s: %w", a.PlaidAccountID, err)
		}
	}

	return queries.UpdateLastSynced(db, item.ID)
}
