package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/tiongMax/gostocks/internal/gateway"
)

func main() {
	// Configure JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
	slog.Info("Connecting to Redis...")
	redisClient, err := gateway.NewRedisClient(redisAddr)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("Connected to Redis")

	// 3. Connect to Alert Service (gRPC)
	slog.Info("Connecting to Alert Service...")
	alertClient, err := gateway.NewAlertClient(alertServiceAddr)
	if err != nil {
		slog.Error("Failed to connect to Alert Service", "error", err)
		os.Exit(1)
	}
	defer alertClient.Close()
	slog.Info("Connected to Alert Service")

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
		slog.Info("API Gateway listening", "port", port)
		if err := router.Run(":" + port); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	slog.Info("Shutdown signal received", "signal", sig)
	slog.Info("Shutting down API Gateway...")
}
