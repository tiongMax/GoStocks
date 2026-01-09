package gateway

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for HTTP handlers.
type Handler struct {
	redis       *RedisClient
	alertClient *AlertClient
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(redis *RedisClient, alertClient *AlertClient) *Handler {
	return &Handler{
		redis:       redis,
		alertClient: alertClient,
	}
}

// GetPrice handles GET /price/:symbol
// Returns the latest price for a stock symbol from Redis.
func (h *Handler) GetPrice(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	price, err := h.redis.GetPrice(c.Request.Context(), symbol)
	if err != nil {
		slog.Error("Redis lookup failed", "symbol", symbol, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if price == nil {
		slog.Debug("Symbol not found in Redis", "symbol", symbol)
		c.JSON(http.StatusNotFound, gin.H{"error": "symbol not found", "symbol": symbol})
		return
	}

	c.JSON(http.StatusOK, price)
}

// CreateAlert handles POST /alerts
// Creates a new price alert via the Alert Service.
func (h *Handler) CreateAlert(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid CreateAlert payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize symbol
	req.Symbol = strings.ToUpper(req.Symbol)
	req.Condition = strings.ToUpper(req.Condition)

	resp, err := h.alertClient.CreateAlert(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to create alert", "symbol", req.Symbol, "user_id", req.UserID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("Alert created", "alert_id", resp.AlertID, "symbol", req.Symbol)
	c.JSON(http.StatusCreated, resp)
}

// GetAlerts handles GET /alerts
// Query params: user_id (optional), active_only (optional, default: false)
func (h *Handler) GetAlerts(c *gin.Context) {
	// Parse query parameters
	userIDStr := c.Query("user_id")
	activeOnlyStr := c.Query("active_only")

	var userID int32 = 0
	if userIDStr != "" {
		id, err := strconv.ParseInt(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		userID = int32(id)
	}

	activeOnly := activeOnlyStr == "true" || activeOnlyStr == "1"

	alerts, err := h.alertClient.GetAlerts(c.Request.Context(), userID, activeOnly)
	if err != nil {
		slog.Error("Failed to fetch alerts", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "gateway",
	})
}
