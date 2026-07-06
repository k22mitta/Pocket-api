package plaidclient

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/k22mitta/pocket-api/internal/db/queries"
)

func StartSyncScheduler(db *sql.DB, client *Client, interval time.Duration) func() {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				runSync(db, client)
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(done)
	}
}

func runSync(db *sql.DB, client *Client) {
	ctx := context.Background()

	items, err := queries.GetAllPlaidItems(db)
	if err != nil {
		log.Printf("scheduler: failed to fetch plaid items: %v", err)
		return
	}

	for _, item := range items {
		log.Printf("scheduler: syncing item %s for user %s", item.ItemID, item.UserID)

		if err := SyncAccounts(ctx, client, db, item.UserID, item); err != nil {
			log.Printf("scheduler: SyncAccounts failed for item %s: %v", item.ItemID, err)
			continue
		}

		accounts, err := queries.GetAccountsByUserID(db, item.UserID)
		if err != nil {
			log.Printf("scheduler: failed to build accounts map for item %s: %v", item.ItemID, err)
			continue
		}
		accountsMap := make(map[string]string, len(accounts))
		for _, a := range accounts {
			accountsMap[a.PlaidAccountID] = a.ID
		}

		if err := SyncTransactions(ctx, client, db, item.UserID, item, accountsMap); err != nil {
			log.Printf("scheduler: SyncTransactions failed for item %s: %v", item.ItemID, err)
			continue
		}

		log.Printf("scheduler: sync complete for item %s", item.ItemID)
	}
}
