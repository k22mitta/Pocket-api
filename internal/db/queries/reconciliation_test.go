package queries

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/k22mitta/pocket-api/internal/models"
)

// testDB connects to the same database the app uses (DATABASE_URL), so this
// exercises the real Postgres month-boundary and aggregation logic rather
// than a mock. Skipped if DATABASE_URL isn't set (e.g. a CI run with no
// database available).
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping DB-backed reconciliation test")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// TestReconciliationInvariant is the invariant the whole app promises a
// finance user: sum(account balances) must equal the net-position endpoint,
// and each budget's "spent" must equal the sum of that category's
// current-month transactions — never last month's, never other categories'.
// It plants transactions straddling the current/previous calendar month
// boundary (using relative dates, not hardcoded ones, so it doesn't rot)
// and hand-computes the expected numbers independently of the SQL under test.
func TestReconciliationInvariant(t *testing.T) {
	db := testDB(t)

	email := fmt.Sprintf("reconcile-test-%d@example.com", time.Now().UnixNano())
	user, err := CreateUser(db, email, "not-a-real-hash", "Reconciliation Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() {
		DeleteAllUserData(db, user.ID)
		db.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	checkingID, err := CreateSeedAccount(db, user.ID, "recon-checking", "Test Bank", "Checking", "depository", "checking", 1000.00)
	if err != nil {
		t.Fatalf("creating checking account: %v", err)
	}
	if _, err := CreateSeedAccount(db, user.ID, "recon-credit", "Test Card Co", "Credit Card", "credit", "credit card", 200.00); err != nil {
		t.Fatalf("creating credit account: %v", err)
	}
	// Loan-type accounts (mortgages, student loans) must subtract from net
	// position the same as credit accounts, in both /accounts and
	// GetTotalBalance's SQL.
	if _, err := CreateSeedAccount(db, user.ID, "recon-loan", "Test Loan Co", "Student Loan", "loan", "student", 5000.00); err != nil {
		t.Fatalf("creating loan account: %v", err)
	}

	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)
	lastMonth := time.Date(now.Year(), now.Month()-1, 15, 0, 0, 0, 0, time.UTC)

	// Internal ledger convention: positive = expense, negative = income.
	txns := []models.Transaction{
		{UserID: user.ID, AccountID: checkingID, PlaidTransactionID: "recon-1", Name: "Groceries this month", Amount: 50.00, Category: "Groceries", Date: thisMonth},
		{UserID: user.ID, AccountID: checkingID, PlaidTransactionID: "recon-2", Name: "Groceries last month", Amount: 30.00, Category: "Groceries", Date: lastMonth},
		{UserID: user.ID, AccountID: checkingID, PlaidTransactionID: "recon-3", Name: "Dining this month", Amount: 20.00, Category: "Dining", Date: thisMonth},
		{UserID: user.ID, AccountID: checkingID, PlaidTransactionID: "recon-4", Name: "Payroll", Amount: -1500.00, Category: "Income", Date: thisMonth},
	}
	for _, tx := range txns {
		if _, err := UpsertTransaction(db, tx); err != nil {
			t.Fatalf("inserting transaction %s: %v", tx.PlaidTransactionID, err)
		}
	}

	if _, err := CreateBudget(db, user.ID, "Groceries", "monthly", 200.00); err != nil {
		t.Fatalf("creating Groceries budget: %v", err)
	}
	if _, err := CreateBudget(db, user.ID, "Dining", "monthly", 100.00); err != nil {
		t.Fatalf("creating Dining budget: %v", err)
	}

	// Invariant 1: net position == sum of account balances, applying the same
	// credit-is-negative-equity convention the /accounts endpoint uses.
	accounts, err := GetAccountsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("GetAccountsByUserID: %v", err)
	}
	wantBalance := 0.0
	for _, a := range accounts {
		bal := 0.0
		if a.CurrentBalance != nil {
			bal = *a.CurrentBalance
		}
		if a.Type == "credit" || a.Type == "loan" {
			bal = -bal
		}
		wantBalance += bal
	}
	gotBalance, err := GetTotalBalance(db, user.ID)
	if err != nil {
		t.Fatalf("GetTotalBalance: %v", err)
	}
	if gotBalance != wantBalance {
		t.Errorf("GetTotalBalance = %v, want %v (sum of displayed account balances: 1000 - 200 - 5000 = -4200)", gotBalance, wantBalance)
	}
	if gotBalance != -4200.00 {
		t.Errorf("GetTotalBalance = %v, want -4200 (1000 checking - 200 credit - 5000 student loan)", gotBalance)
	}

	// Invariant 2: each budget's spent == sum of THIS MONTH's transactions in
	// that exact category. Groceries must be 50 (this month only, excluding
	// the 30 from last month), Dining must be 20, and Income must never count
	// as spend anywhere.
	budgets, err := GetBudgetsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("GetBudgetsByUserID: %v", err)
	}
	spentByCategory := map[string]float64{}
	for _, b := range budgets {
		spentByCategory[b.Category] = b.Spent
	}
	if spentByCategory["Groceries"] != 50.00 {
		t.Errorf(`Groceries budget spent = %v, want 50 (this month's 50 only, not last month's 30)`, spentByCategory["Groceries"])
	}
	if spentByCategory["Dining"] != 20.00 {
		t.Errorf(`Dining budget spent = %v, want 20`, spentByCategory["Dining"])
	}

	// Invariant 3: the spending-by-category summary for the current month
	// must never include income, and must match the budgets endpoint exactly.
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	spending, err := GetSpendingByCategory(db, user.ID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("GetSpendingByCategory: %v", err)
	}
	if _, hasIncome := spending["Income"]; hasIncome {
		t.Error("GetSpendingByCategory included an Income category — income must never count as spending")
	}
	if spending["Groceries"] != spentByCategory["Groceries"] {
		t.Errorf("GetSpendingByCategory Groceries (%v) disagrees with budget spent (%v)", spending["Groceries"], spentByCategory["Groceries"])
	}
	if spending["Dining"] != spentByCategory["Dining"] {
		t.Errorf("GetSpendingByCategory Dining (%v) disagrees with budget spent (%v)", spending["Dining"], spentByCategory["Dining"])
	}
}

// TestReconciliationInvariantDemoSeedShape mirrors the same invariant but
// against data shaped like the demo seed (3 accounts including a credit
// card, budgets across multiple categories) to catch category- or
// account-count-specific regressions the single-account test above wouldn't.
func TestReconciliationInvariantDemoSeedShape(t *testing.T) {
	db := testDB(t)

	email := fmt.Sprintf("reconcile-demo-shape-%d@example.com", time.Now().UnixNano())
	user, err := CreateUser(db, email, "not-a-real-hash", "Reconciliation Demo Shape")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() {
		DeleteAllUserData(db, user.ID)
		db.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	checkingID, _ := CreateSeedAccount(db, user.ID, "recon2-checking", "Bank", "Checking", "depository", "checking", 3000.00)
	savingsID, _ := CreateSeedAccount(db, user.ID, "recon2-savings", "Bank", "Savings", "depository", "savings", 5000.00)
	creditID, _ := CreateSeedAccount(db, user.ID, "recon2-credit", "Card Co", "Credit Card", "credit", "credit card", 400.00)

	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 10, 0, 0, 0, 0, time.UTC)

	for i, tx := range []models.Transaction{
		{AccountID: checkingID, Category: "Housing", Amount: 1200.00},
		{AccountID: creditID, Category: "Shopping", Amount: 75.00},
		{AccountID: savingsID, Category: "Income", Amount: -10.00},
	} {
		tx.UserID = user.ID
		tx.PlaidTransactionID = fmt.Sprintf("recon2-%d", i)
		tx.Name = tx.Category
		tx.Date = thisMonth
		if _, err := UpsertTransaction(db, tx); err != nil {
			t.Fatalf("inserting transaction: %v", err)
		}
	}
	CreateBudget(db, user.ID, "Housing", "monthly", 1500.00)
	CreateBudget(db, user.ID, "Shopping", "monthly", 100.00)

	accounts, err := GetAccountsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("GetAccountsByUserID: %v", err)
	}
	if len(accounts) != 3 {
		t.Fatalf("expected 3 accounts, got %d", len(accounts))
	}
	wantBalance := 3000.00 + 5000.00 - 400.00
	gotBalance, err := GetTotalBalance(db, user.ID)
	if err != nil {
		t.Fatalf("GetTotalBalance: %v", err)
	}
	if gotBalance != wantBalance {
		t.Errorf("GetTotalBalance = %v, want %v", gotBalance, wantBalance)
	}

	budgets, err := GetBudgetsByUserID(db, user.ID)
	if err != nil {
		t.Fatalf("GetBudgetsByUserID: %v", err)
	}
	for _, b := range budgets {
		switch b.Category {
		case "Housing":
			if b.Spent != 1200.00 {
				t.Errorf("Housing spent = %v, want 1200", b.Spent)
			}
		case "Shopping":
			if b.Spent != 75.00 {
				t.Errorf("Shopping spent = %v, want 75", b.Spent)
			}
		}
	}
}
