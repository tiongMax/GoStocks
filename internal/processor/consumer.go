package processor

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	stock "github.com/tiongMax/gostocks/proto"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	brokers []string
	topic   string
}

func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
	}
}

func (c *Consumer) Start() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(c.brokers, config)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing consumer: %v", err)
		}
	}()

	// Consume partition 0
	partitionConsumer, err := consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Error closing partition consumer: %v", err)
		}
	}()

	log.Printf("Connected to Kafka, consuming topic: %s", c.topic)

	// Metrics loop
	msgCount := 0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var tick stock.StockTick
			if err := proto.Unmarshal(msg.Value, &tick); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}
			msgCount++

		case <-ticker.C:
			if msgCount > 0 {
				log.Printf("Throughput: %d messages/sec", msgCount)
				msgCount = 0
			}

		case err := <-partitionConsumer.Errors():
			log.Printf("Kafka error: %v", err)
		}
	}
}


