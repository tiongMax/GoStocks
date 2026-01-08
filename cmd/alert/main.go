package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tiongMax/gostocks/internal/alert"
	pb "github.com/tiongMax/gostocks/proto/alert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Connection String (Default to localhost for local dev)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=user password=password dbname=gostocks sslmode=disable host=127.0.0.1 port=5433"
	}

	// 2. gRPC Port
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	log.Println("Connecting to database...")
	store, err := alert.NewStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// 3. Auto-migrate schema
	log.Println("Auto-migrating schema...")
	if err := store.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate schema: %v", err)
	}
	log.Println("Schema migrated successfully.")

	// 4. Create gRPC server
	grpcServer := grpc.NewServer()
	alertServer := alert.NewServer(store)
	pb.RegisterAlertServiceServer(grpcServer, alertServer)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	// 5. Start listening
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// 6. Run server in goroutine
	go func() {
		log.Printf("Alert Service gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 7. Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Alert Service...")
	grpcServer.GracefulStop()
	log.Println("Alert Service stopped.")
}
