package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	var database *sql.DB
	var err error

	for i := 1; i <= 5; i++ {
		database, err = sql.Open("postgres", databaseURL)
		if err == nil {
			err = database.Ping()
		}
		if err == nil {
			database.SetMaxOpenConns(25)
			database.SetMaxIdleConns(5)
			database.SetConnMaxLifetime(5 * time.Minute)
			return database, nil
		}
		log.Printf("db connection attempt %d/5 failed: %v", i, err)
		if i < 5 {
			time.Sleep(2 * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to connect after 5 attempts: %w", err)
}
