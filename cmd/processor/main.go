package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tiongMax/gostocks/internal/processor"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	log.Println("Starting Processor Service...")
	consumer := processor.NewConsumer(brokers, "market_ticks")

	if err := consumer.Start(); err != nil {
		log.Fatal("Failed to start processor:", err)
	}
}

