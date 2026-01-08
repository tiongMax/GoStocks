package ingestor

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	stock "github.com/tiongMax/gostocks/proto/stock"
	"google.golang.org/protobuf/proto"
)

// Client represents the ingestor client that connects to the Finnhub WebSocket API
// and produces market data messages to a Kafka topic.
type Client struct {
	apiKey   string
	conn     *websocket.Conn
	symbols  []string
	producer sarama.SyncProducer
}

// NewClient creates a new instance of the ingestor Client.
// It initializes the Kafka producer with the provided broker addresses.
func NewClient(apiKey string, symbols []string, kafkaBrokers []string) (*Client, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(kafkaBrokers, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:   apiKey,
		symbols:  symbols,
		producer: producer,
	}, nil
}

// Start connects to the Finnhub WebSocket API, subscribes to the configured symbols,
// and starts the background read loop to process incoming trade messages.
func (c *Client) Start() error {
	url := "wss://ws.finnhub.io?token=" + c.apiKey
	log.Printf("Connecting to %s", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	// Subscribe to symbols
	for _, s := range c.symbols {
		msg := map[string]interface{}{
			"type":   "subscribe",
			"symbol": s,
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			return err
		}
		log.Printf("Subscribed to %s", s)
	}

	// Start reading loop
	go c.readLoop()
	return nil
}

// readLoop continuously reads messages from the WebSocket connection,
// processes trade data, and sends it to the Kafka topic.
// It handles JSON unmarshalling, Protobuf serialization, and Kafka production.
func (c *Client) readLoop() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var response FinnhubResponse
		if err := json.Unmarshal(message, &response); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			continue
		}

		if response.Type == "trade" {
			for _, trade := range response.Data {
				log.Printf("Tick: %s | Price: %.2f | Time: %d", trade.Symbol, trade.Price, trade.Timestamp)

				// Create Protobuf message
				tick := &stock.StockTick{
					Symbol:    trade.Symbol,
					Price:     trade.Price,
					Timestamp: trade.Timestamp,
				}

				// Serialize to bytes
				bytes, err := proto.Marshal(tick)
				if err != nil {
					log.Printf("Failed to marshal protobuf: %v", err)
					continue
				}

				// Send to Kafka
				msg := &sarama.ProducerMessage{
					Topic: "market_ticks",
					Key:   sarama.StringEncoder(trade.Symbol), // Use Symbol as Key for ordering
					Value: sarama.ByteEncoder(bytes),
				}

				partition, offset, err := c.producer.SendMessage(msg)
				if err != nil {
					log.Printf("Failed to send message to Kafka: %v", err)
				} else {
					log.Printf("Message sent to partition %d at offset %d", partition, offset)
				}
			}
		}
	}
}

// Close gracefully shuts down the client by closing the Kafka producer
// and sending a close message to the WebSocket connection.
func (c *Client) Close() error {
	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			log.Println("Error closing Kafka producer:", err)
		}
	}
	if c.conn != nil {
		return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
	return nil
}
