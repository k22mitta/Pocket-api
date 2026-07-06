package queries

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/k22mitta/pocket-api/internal/models"
)

// TestDeletePlaidItemCascadesAccountsAndTransactions asserts that removing a
// plaid_item removes every account and transaction under it, and leaves
// everything else on the user's other connections (e.g. a CSV import) alone
// — this is the invariant that makes "unlink this bank" safe to expose in
// the UI: it can't partially remove data or reach into an unrelated item.
func TestDeletePlaidItemCascadesAccountsAndTransactions(t *testing.T) {
	db := testDB(t)

	email := fmt.Sprintf("unlink-test-%d@example.com", time.Now().UnixNano())
	user, err := CreateUser(db, email, "not-a-real-hash", "Unlink Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() {
		DeleteAllUserData(db, user.ID)
		db.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	// A real Plaid connection (real-looking access token) with one account
	// and one transaction — this is what gets unlinked.
	plaidItemID := fmt.Sprintf("unlink-item-%d", time.Now().UnixNano())
	if err := CreatePlaidItem(db, user.ID, "access-sandbox-fake-token", plaidItemID, "ins_1", "Test Bank"); err != nil {
		t.Fatalf("creating plaid item: %v", err)
	}
	item, err := GetPlaidItemByItemID(db, plaidItemID)
	if err != nil {
		t.Fatalf("fetching created plaid item: %v", err)
	}

	balance := 500.0
	account := models.Account{
		PlaidItemID:    item.ID,
		UserID:         user.ID,
		PlaidAccountID: "unlink-account-1",
		Name:           "Test Checking",
		Type:           "depository",
		CurrentBalance: &balance,
		CurrencyCode:   "USD",
	}
	if err := UpsertAccount(db, account); err != nil {
		t.Fatalf("creating account: %v", err)
	}
	accounts, err := GetAccountsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("fetching accounts: %v", err)
	}
	var accountID string
	for _, a := range accounts {
		if a.PlaidAccountID == "unlink-account-1" {
			accountID = a.ID
		}
	}
	if accountID == "" {
		t.Fatal("test account was not created")
	}

	tx := models.Transaction{
		UserID:             user.ID,
		AccountID:          accountID,
		PlaidTransactionID: fmt.Sprintf("unlink-tx-%d", time.Now().UnixNano()),
		Name:               "Test Purchase",
		Amount:             25.00,
		Category:           "Shopping",
		Date:               time.Now(),
	}
	if _, err := UpsertTransaction(db, tx); err != nil {
		t.Fatalf("creating transaction: %v", err)
	}

	// A second, unrelated item (a CSV import) that must survive the unlink.
	csvAccountID, err := CreateSeedAccount(db, user.ID, "unlink-csv", "Imported", "CSV Import", "depository", "checking", 100.00)
	if err != nil {
		t.Fatalf("creating csv seed account: %v", err)
	}

	if err := DeletePlaidItem(db, item.ID, user.ID); err != nil {
		t.Fatalf("DeletePlaidItem: %v", err)
	}

	if _, err := GetPlaidItemByID(db, item.ID, user.ID); err != sql.ErrNoRows {
		t.Errorf("plaid_item still exists after delete, err = %v, want sql.ErrNoRows", err)
	}

	remaining, err := GetAccountsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("fetching remaining accounts: %v", err)
	}
	for _, a := range remaining {
		if a.ID == accountID {
			t.Error("unlinked item's account still exists")
		}
	}
	found := false
	for _, a := range remaining {
		if a.ID == csvAccountID {
			found = true
		}
	}
	if !found {
		t.Error("unrelated CSV-import account was removed by unlinking a different item")
	}

	var txCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE account_id = $1`, accountID).Scan(&txCount); err != nil {
		t.Fatalf("counting transactions: %v", err)
	}
	if txCount != 0 {
		t.Errorf("transactions under unlinked account = %d, want 0", txCount)
	}
}
