package main

import (
	"context"
	"log/slog"
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
	// Configure JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
	slog.Info("Connecting to database...")
	store, err := alert.NewStore(connStr)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// 5. Auto-migrate schema
	slog.Info("Auto-migrating schema...")
	if err := store.AutoMigrate(); err != nil {
		slog.Error("Failed to auto-migrate schema", "error", err)
		os.Exit(1)
	}
	slog.Info("Schema migrated successfully")

	// 6. Start Kafka Consumer (Trigger Logic)
	ctx, cancel := context.WithCancel(context.Background())
	consumer := alert.NewConsumer(brokers, kafkaTopic, store)

	go func() {
		slog.Info("Starting Alert Consumer")
		if err := consumer.Start(ctx); err != nil {
			slog.Error("Alert Consumer failed", "error", err)
			cancel() // Stop everything if consumer fails
		}
	}()

	// 7. Create gRPC Server
	grpcServer := grpc.NewServer()
	alertServer := alert.NewServer(store)
	pb.RegisterAlertServiceServer(grpcServer, alertServer)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	// 8. Start gRPC Listener
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("Failed to listen on port", "port", grpcPort, "error", err)
		os.Exit(1)
	}

	// 9. Run gRPC Server in goroutine
	go func() {
		slog.Info("Alert Service gRPC server listening", "port", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("Failed to serve gRPC", "error", err)
			cancel()
		}
	}()

	slog.Info("Alert Service is fully running (gRPC + Kafka Consumer)")

	// 10. Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		slog.Info("Shutdown signal received", "signal", sig)
	case <-ctx.Done():
		slog.Info("Context cancelled, shutting down")
	}

	// Graceful shutdown
	cancel()
	grpcServer.GracefulStop()
	slog.Info("Alert Service stopped")
}
