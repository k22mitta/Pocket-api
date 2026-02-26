package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourname/pocket-api/internal/ai"
	"github.com/yourname/pocket-api/internal/api"
	"github.com/yourname/pocket-api/internal/api/middleware"
	"github.com/yourname/pocket-api/internal/config"
	"github.com/yourname/pocket-api/internal/db"
	plaidclient "github.com/yourname/pocket-api/internal/plaid"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Printf("database connection failed: %v", err)
		os.Exit(1)
	}

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	plaidClient := plaidclient.NewClient(cfg.PlaidClientID, cfg.PlaidSecret, cfg.PlaidEnv)

	ctx := context.Background()
	aiClient, err := ai.NewClient(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("ai client failed: %v", err)
	}
	defer aiClient.Close()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.Logger()(middleware.CORS()(api.NewRouter(cfg, database, plaidClient, aiClient))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	stopSync := plaidclient.StartSyncScheduler(database, plaidClient, 6*time.Hour)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	stopSync()
	database.Close()
	log.Println("server stopped")
}
