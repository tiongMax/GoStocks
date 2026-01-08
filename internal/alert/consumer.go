package alert

import (
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	stock "github.com/tiongMax/gostocks/proto/stock"
	"google.golang.org/protobuf/proto"
)

// Consumer listens to Kafka and triggers alerts based on price conditions.
type Consumer struct {
	brokers  []string
	topic    string
	store    *Store
	done     chan struct{}
	wg       sync.WaitGroup
	consumer sarama.Consumer
}

// NewConsumer creates a new Kafka consumer for the Alert Service.
func NewConsumer(brokers []string, topic string, store *Store) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
		store:   store,
		done:    make(chan struct{}),
	}
}

// Start begins consuming messages from Kafka and checking alerts.
func (c *Consumer) Start() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(c.brokers, config)
	if err != nil {
		return err
	}
	c.consumer = consumer

	partitionConsumer, err := consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		consumer.Close()
		return err
	}

	log.Printf("[Alert Consumer] Connected to Kafka, consuming topic: %s", c.topic)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer partitionConsumer.Close()

		// Metrics
		tickCount := 0
		alertsTriggered := 0
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.done:
				log.Println("[Alert Consumer] Shutting down...")
				return

			case msg := <-partitionConsumer.Messages():
				// 1. Deserialize the tick
				var tick stock.StockTick
				if err := proto.Unmarshal(msg.Value, &tick); err != nil {
					log.Printf("[Alert Consumer] Error unmarshaling: %v", err)
					continue
				}
				tickCount++

				// 2. Check alerts for this symbol
				triggered, err := c.checkAlerts(tick.Symbol, tick.Price)
				if err != nil {
					log.Printf("[Alert Consumer] Error checking alerts: %v", err)
					continue
				}
				alertsTriggered += triggered

			case err := <-partitionConsumer.Errors():
				log.Printf("[Alert Consumer] Kafka error: %v", err)

			case <-ticker.C:
				if tickCount > 0 || alertsTriggered > 0 {
					log.Printf("[Alert Consumer] Processed %d ticks, triggered %d alerts in last 10s",
						tickCount, alertsTriggered)
					tickCount = 0
					alertsTriggered = 0
				}
			}
		}
	}()

	return nil
}

// checkAlerts compares the current price against all active alerts for a symbol.
// Returns the number of alerts triggered.
func (c *Consumer) checkAlerts(symbol string, price float64) (int, error) {
	alerts, err := c.store.GetActiveAlertsBySymbol(symbol)
	if err != nil {
		return 0, err
	}

	triggered := 0
	for _, alert := range alerts {
		shouldTrigger := false

		switch alert.Condition {
		case "ABOVE":
			shouldTrigger = price >= alert.TargetPrice
		case "BELOW":
			shouldTrigger = price <= alert.TargetPrice
		}

		if shouldTrigger {
			// Mark as triggered in database
			if err := c.store.MarkAlertTriggered(alert.ID); err != nil {
				log.Printf("[Alert Consumer] Failed to mark alert %d as triggered: %v", alert.ID, err)
				continue
			}

			// Log the trigger (in production, this would send a notification)
			log.Printf("🔔 ALERT TRIGGERED! User %d: %s is now $%.2f (%s $%.2f)",
				alert.UserID, symbol, price, alert.Condition, alert.TargetPrice)

			triggered++
		}
	}

	return triggered, nil
}

// Stop gracefully shuts down the consumer.
func (c *Consumer) Stop() {
	close(c.done)
	c.wg.Wait()
	if c.consumer != nil {
		c.consumer.Close()
	}
	log.Println("[Alert Consumer] Stopped.")
}

