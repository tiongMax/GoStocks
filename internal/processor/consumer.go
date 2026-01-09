package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	stock "github.com/tiongMax/gostocks/proto/stock"
	"google.golang.org/protobuf/proto"
)

// Consumer manages the connection to Kafka and processing logic.
type Consumer struct {
	brokers []string
	topic   string
	groupID string
}

// NewConsumer creates a Consumer instance.
func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
		groupID: "processor-group",
	}
}

// Start initializes the Kafka consumer group and starts processing.
func (c *Consumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	// Start consuming from the oldest offset if no offset is committed
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(c.brokers, c.groupID, config)
	if err != nil {
		return err
	}
	defer group.Close()

	slog.Info("Connected to Kafka Consumer Group", "group", c.groupID, "topic", c.topic)

	// Handler for consumer group
	handler := &GroupHandler{}

	for {
		// Consume claims runs the handler for the claimed partitions
		if err := group.Consume(ctx, []string{c.topic}, handler); err != nil {
			slog.Error("Error from consumer group", "error", err)
			return err
		}
		// Check if context was cancelled
		if ctx.Err() != nil {
			return nil
		}
	}
}

// GroupHandler implements sarama.ConsumerGroupHandler
type GroupHandler struct {
	msgCount int
}

func (h *GroupHandler) Setup(sarama.ConsumerGroupSession) error {
	slog.Info("Consumer group session setup")
	return nil
}

func (h *GroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	slog.Info("Consumer group session cleanup")
	return nil
}

func (h *GroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Report throughput
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	msgChan := claim.Messages()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			// Process message
			var tick stock.StockTick
			if err := proto.Unmarshal(msg.Value, &tick); err != nil {
				slog.Error("Error unmarshaling message", "error", err)
				continue
			}

			h.msgCount++

			// Mark message as marked
			session.MarkMessage(msg, "")

		case <-ticker.C:
			if h.msgCount > 0 {
				slog.Info("Throughput", "msgs_per_sec", h.msgCount)
				h.msgCount = 0
			}

		case <-session.Context().Done():
			return nil
		}
	}
}
