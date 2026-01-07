package processor

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	stock "github.com/tiongMax/gostocks/proto"
	"google.golang.org/protobuf/proto"
)

// Consumer manages the connection to Kafka and processing logic.
type Consumer struct {
	brokers []string
	topic   string
}

// NewConsumer creates a Consumer instance.
// brokers: List of Kafka bootstrap servers (e.g., ["localhost:9092"])
// topic: The Kafka topic name to consume from (e.g., "market_ticks")
func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
	}
}

// Start initializes the Kafka connection and enters a blocking loop.
// It consumes messages from Partition 0 and logs throughput metrics.
func (c *Consumer) Start() error {
	// 1. Configure Sarama
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// 2. Connect to Broker
	consumer, err := sarama.NewConsumer(c.brokers, config)
	if err != nil {
		return err
	}
	// ... defer close ...

	// 3. Start Consuming Partition 0
	// OffsetNewest means we only see messages sent after we connect.
	partitionConsumer, err := consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	// ... defer close ...

	log.Printf("Connected to Kafka, consuming topic: %s", c.topic)

	// 4. Main Event Loop
	msgCount := 0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			// A. Deserialize Protobuf
			var tick stock.StockTick
			if err := proto.Unmarshal(msg.Value, &tick); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}
			msgCount++ // B. Increment metric

		case <-ticker.C:
			// C. Report Throughput
			if msgCount > 0 {
				log.Printf("Throughput: %d messages/sec", msgCount)
				msgCount = 0
			}

		case err := <-partitionConsumer.Errors():
			log.Printf("Kafka error: %v", err)
		}
	}
}
