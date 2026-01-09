package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/tiongMax/gostocks/internal/processor"
)

func main() {
	// Configure JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Load Environment Variables
	if err := godotenv.Load(".env"); err != nil {
		slog.Info("No .env file found, using system environment variables")
	}

	// 2. Configure Kafka Brokers
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// 3. Initialize Consumer
	slog.Info("Starting Processor Service...")
	consumer := processor.NewConsumer(brokers, "market_ticks")

	// 4. Handle Shutdown Signals
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("Shutdown signal received", "signal", sig)
		cancel()
	}()

	// 5. Start Processing
	if err := consumer.Start(ctx); err != nil {
		slog.Error("Processor failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Processor Service stopped successfully")
}
