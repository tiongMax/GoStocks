package alert

import (
	"context"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	stock "github.com/tiongMax/gostocks/proto/stock"
	"google.golang.org/protobuf/proto"
)

// Consumer listens to Kafka and triggers alerts based on price conditions.
type Consumer struct {
	brokers []string
	topic   string
	store   *Store
	groupID string
}

// NewConsumer creates a new Kafka consumer for the Alert Service.
func NewConsumer(brokers []string, topic string, store *Store) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
		store:   store,
		groupID: "alert-service-group",
	}
}

// Start begins consuming messages from Kafka and checking alerts.
func (c *Consumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(c.brokers, c.groupID, config)
	if err != nil {
		return err
	}
	defer group.Close()

	slog.Info("Connected to Kafka Consumer Group", "group", c.groupID, "topic", c.topic)

	handler := &AlertGroupHandler{
		store: c.store,
	}

	for {
		if err := group.Consume(ctx, []string{c.topic}, handler); err != nil {
			slog.Error("Error from consumer group", "error", err)
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

// AlertGroupHandler implements sarama.ConsumerGroupHandler
type AlertGroupHandler struct {
	store *Store
}

func (h *AlertGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *AlertGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *AlertGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Metrics
	tickCount := 0
	alertsTriggered := 0
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	msgChan := claim.Messages()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			// 1. Deserialize the tick
			var tick stock.StockTick
			if err := proto.Unmarshal(msg.Value, &tick); err != nil {
				slog.Error("Error unmarshaling message", "error", err)
				continue
			}
			tickCount++

			// 2. Check alerts for this symbol
			triggered, err := h.checkAlerts(tick.Symbol, tick.Price)
			if err != nil {
				slog.Error("Error checking alerts", "error", err)
				continue
			}
			alertsTriggered += triggered

			// Mark message
			session.MarkMessage(msg, "")

		case <-ticker.C:
			if tickCount > 0 || alertsTriggered > 0 {
				slog.Info("Metrics", "ticks_processed", tickCount, "alerts_triggered", alertsTriggered)
				tickCount = 0
				alertsTriggered = 0
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// checkAlerts compares the current price against all active alerts for a symbol.
func (h *AlertGroupHandler) checkAlerts(symbol string, price float64) (int, error) {
	alerts, err := h.store.GetActiveAlertsBySymbol(symbol)
	if err != nil {
		return 0, err
	}

	triggered := 0
	for _, alert := range alerts {
		if ShouldTriggerAlert(alert.Condition, alert.TargetPrice, price) {
			// Mark as triggered in database
			if err := h.store.MarkAlertTriggered(alert.ID); err != nil {
				slog.Error("Failed to mark alert as triggered", "alert_id", alert.ID, "error", err)
				continue
			}

			// Log the trigger
			slog.Info("🔔 ALERT TRIGGERED!",
				"user_id", alert.UserID,
				"symbol", symbol,
				"price", price,
				"condition", alert.Condition,
				"target_price", alert.TargetPrice)

			triggered++
		}
	}

	return triggered, nil
}

// ShouldTriggerAlert contains the pure logic for checking if an alert condition is met.
func ShouldTriggerAlert(condition string, targetPrice, currentPrice float64) bool {
	switch condition {
	case "ABOVE":
		return currentPrice >= targetPrice
	case "BELOW":
		return currentPrice <= targetPrice
	default:
		return false
	}
}
