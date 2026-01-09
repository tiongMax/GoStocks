package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/tiongMax/gostocks/internal/gateway"
)

func main() {
	// 1. Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	alertServiceAddr := os.Getenv("ALERT_SERVICE_ADDR")
	if alertServiceAddr == "" {
		alertServiceAddr = "localhost:50051"
	}

	// 2. Connect to Redis (Speed Layer)
	log.Println("Connecting to Redis...")
	redisClient, err := gateway.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis.")

	// 3. Connect to Alert Service (gRPC)
	log.Println("Connecting to Alert Service...")
	alertClient, err := gateway.NewAlertClient(alertServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Alert Service: %v", err)
	}
	defer alertClient.Close()
	log.Println("Connected to Alert Service.")

	// 4. Set up Gin router
	router := gin.Default()

	// Create handler with dependencies
	handler := gateway.NewHandler(redisClient, alertClient)

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Day 11: Price endpoint (Redis lookup)
	router.GET("/price/:symbol", handler.GetPrice)

	// Day 12: Alert endpoints (gRPC to Alert Service)
	router.POST("/alerts", handler.CreateAlert)
	router.GET("/alerts", handler.GetAlerts)

	// 5. Start server in goroutine
	go func() {
		log.Printf("API Gateway listening on :%s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 6. Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down API Gateway...")
}

