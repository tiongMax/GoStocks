package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tiongMax/gostocks/internal/alert"
)

func main() {
	// 1. Connection String (Default to localhost for local dev)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Try DSN format
		connStr = "user=user password=password dbname=gostocks sslmode=disable host=127.0.0.1 port=5433"
	}

	log.Println("Connecting to database...")
	store, err := alert.NewStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// 2. Auto-migrate schema
	log.Println("Auto-migrating schema...")
	if err := store.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate schema: %v", err)
	}
	log.Println("Schema migrated successfully.")

	// 3. Wait for shutdown signal
	log.Println("Alert Service is running...")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Alert Service...")
}
