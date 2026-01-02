package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tiongMax/gostocks/internal/ingestor"
)

func main() {
	// 0. Load Environment Variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 1. Get API Key
	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		log.Fatal("FINNHUB_API_KEY environment variable is not set")
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// 2. Initialize and Start Client
	symbols := []string{"AAPL", "BINANCE:BTCUSDT", "IC MARKETS:1"}
	client, err := ingestor.NewClient(apiKey, symbols, brokers)
	if err != nil {
		log.Fatal("Failed to create ingestor client:", err)
	}

	if err := client.Start(); err != nil {
		log.Fatal("Failed to start ingestor:", err)
	}

	// 3. Handle Graceful Shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("Interrupt received, shutting down...")

	if err := client.Close(); err != nil {
		log.Println("Error closing connection:", err)
	}

	// Give it a moment to close
	time.Sleep(1 * time.Second)
	log.Println("Shutdown complete")
}
