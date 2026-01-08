package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tiongMax/gostocks/internal/alert"
	pb "github.com/tiongMax/gostocks/proto/alert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Database Connection String
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=user password=password dbname=gostocks sslmode=disable host=127.0.0.1 port=5433"
	}

	// 2. gRPC Port
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// 3. Kafka Configuration
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "market_ticks"
	}

	// 4. Connect to Database
	log.Println("Connecting to database...")
	store, err := alert.NewStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// 5. Auto-migrate schema
	log.Println("Auto-migrating schema...")
	if err := store.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate schema: %v", err)
	}
	log.Println("Schema migrated successfully.")

	// 6. Start Kafka Consumer (Trigger Logic)
	consumer := alert.NewConsumer(brokers, kafkaTopic, store)
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Stop()

	// 7. Create gRPC Server
	grpcServer := grpc.NewServer()
	alertServer := alert.NewServer(store)
	pb.RegisterAlertServiceServer(grpcServer, alertServer)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	// 8. Start gRPC Listener
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// 9. Run gRPC Server in goroutine
	go func() {
		log.Printf("Alert Service gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Println("Alert Service is fully running (gRPC + Kafka Consumer)")

	// 10. Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Alert Service...")
	grpcServer.GracefulStop()
	log.Println("Alert Service stopped.")
}
