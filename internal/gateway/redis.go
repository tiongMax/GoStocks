package gateway

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis connection for price lookups.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client connection.
func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password
		DB:       0,  // default DB
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// PriceData represents the price information for a symbol.
type PriceData struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// GetPrice retrieves the latest price for a symbol from Redis.
// Returns nil if the symbol is not found.
func (r *RedisClient) GetPrice(ctx context.Context, symbol string) (*PriceData, error) {
	key := fmt.Sprintf("price:%s", symbol)

	// Get price value
	priceStr, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price format: %w", err)
	}

	// Try to get timestamp (optional)
	timestampKey := fmt.Sprintf("price:%s:timestamp", symbol)
	timestamp := time.Now().Unix()
	if ts, err := r.client.Get(ctx, timestampKey).Int64(); err == nil {
		timestamp = ts
	}

	return &PriceData{
		Symbol:    symbol,
		Price:     price,
		Timestamp: timestamp,
	}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

