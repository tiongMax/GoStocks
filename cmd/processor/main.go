package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tiongMax/gostocks/internal/processor"
)

func main() {
	// 1. Load Environment Variables
	// Tries to load variables from a local .env file.
	// If missing, it falls back to system-level environment variables (useful for Docker/CI).
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 2. Configure Kafka Brokers
	// Fetches the comma-separated list of brokers (e.g., "host1:9092,host2:9092").
	// Defaults to "localhost:9092" for local development.
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// 3. Initialize Consumer
	// Creates a new consumer instance pointing to the 'market_ticks' topic.
	log.Println("Starting Processor Service...")
	consumer := processor.NewConsumer(brokers, "market_ticks")

	// 4. Start Processing
	// consumer.Start() is a blocking call that enters a loop to consume messages.
	// The program will exit only if Start() returns an error (e.g., connection failure).
	if err := consumer.Start(); err != nil {
		log.Fatal("Failed to start processor:", err)
	}
}
