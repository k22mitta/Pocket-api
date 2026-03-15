package plaidclient

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func SyncTransactions(ctx context.Context, client *Client, db *sql.DB, userID string, item models.PlaidItem, accounts map[string]string) error {
	now := time.Now()
	startDate := now.AddDate(0, 0, -90).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	resp, _, err := client.api.PlaidApi.TransactionsGet(ctx).
		TransactionsGetRequest(*plaid.NewTransactionsGetRequest(item.AccessToken, startDate, endDate)).
		Execute()
	if err != nil {
		return fmt.Errorf("fetching transactions: %w", err)
	}

	for _, pt := range resp.GetTransactions() {
		accountID, ok := accounts[pt.GetAccountId()]
		if !ok {
			continue
		}

		date, err := time.Parse("2006-01-02", pt.GetDate())
		if err != nil {
			continue
		}

		cats, _ := pt.GetCategoryOk()
		var plaidCat *string
		var catSlice []string
		if cats != nil {
			catSlice = *cats
			if len(catSlice) > 0 {
				s := catSlice[0]
				plaidCat = &s
			}
		}

		amount := pt.GetAmount()
		category := categorizeTransaction(catSlice, amount, pt.GetName())

		t := models.Transaction{
			UserID:             userID,
			AccountID:          accountID,
			PlaidTransactionID: pt.GetTransactionId(),
			Name:               pt.GetName(),
			Amount:             amount,
			Category:           category,
			PlaidCategory:      plaidCat,
			Date:               date,
			Pending:            pt.GetPending(),
		}

		if name, ok := pt.GetMerchantNameOk(); ok && name != nil {
			t.MerchantName = name
		}

		if _, err := queries.UpsertTransaction(db, t); err != nil {
			return fmt.Errorf("upserting transaction %s: %w", t.PlaidTransactionID, err)
		}
	}

	return nil
}

// categorizeTransaction combines Plaid's raw category with a sign check:
// positive amount = money leaving the account (Plaid's own convention), so a
// transaction can only be Income if it's actually negative (an inflow).
// Without this, Plaid's coarse "payroll"/"deposit" keyword categories get
// assigned regardless of direction, and an outflow-direction transaction
// (e.g. an ACH debit that happens to be tagged similarly to a payroll
// deposit) renders as a paycheck the user never received.
//
// The merchant-name fallback (CategorizeByMerchant) has the exact same
// keyword-only blind spot — e.g. a merchant literally named "CD DEPOSIT
// .INITIAL." matches its "deposit" keyword and also returns Income with no
// sign check. So the guard is re-applied after the fallback too: no code
// path here can ever return Income for a non-negative amount.
func categorizeTransaction(plaidCategories []string, amount float64, name string) string {
	category := NormalizeCategory(plaidCategories)
	if category == "Income" && amount >= 0 {
		if isTransferLike(plaidCategories) {
			return "Transfers"
		}
		category = CategorizeByMerchant(name)
		if category == "Income" {
			category = "Other"
		}
	}
	return category
}
