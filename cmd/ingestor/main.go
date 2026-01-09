package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tiongMax/gostocks/internal/ingestor"
)

func main() {
	// Configure JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Load Environment Variables
	if err := godotenv.Load(".env"); err != nil {
		slog.Info("No .env file found, using system environment variables")
	}

	// 2. Configuration
	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		slog.Error("FINNHUB_API_KEY environment variable is not set")
		os.Exit(1)
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// 3. Initialize Client
	symbols := []string{"AAPL", "BINANCE:BTCUSDT", "IC MARKETS:1"}
	client, err := ingestor.NewClient(apiKey, symbols, brokers)
	if err != nil {
		slog.Error("Failed to create ingestor client", "error", err)
		os.Exit(1)
	}

	// 4. Start Client
	if err := client.Start(); err != nil {
		slog.Error("Failed to start ingestor", "error", err)
		os.Exit(1)
	}

	slog.Info("Ingestor service started", "symbols", symbols)

	// 5. Wait for Shutdown Signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	slog.Info("Shutdown signal received", "signal", sig)

	// 6. Graceful Shutdown
	if err := client.Close(); err != nil {
		slog.Error("Error closing client", "error", err)
	}

	time.Sleep(500 * time.Millisecond)
	slog.Info("Ingestor service stopped")
}
