package ingestor

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	stock "github.com/tiongMax/gostocks/proto/stock"
	"google.golang.org/protobuf/proto"
)

// Client represents the ingestor client.
type Client struct {
	apiKey   string
	conn     *websocket.Conn
	symbols  []string
	producer sarama.SyncProducer
	done     chan struct{}
}

// NewClient creates a new ingestor client.
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
		done:     make(chan struct{}),
	}, nil
}

// Start connects to WebSocket and starts the read loop.
func (c *Client) Start() error {
	url := "wss://ws.finnhub.io?token=" + c.apiKey
	slog.Info("Connecting to Finnhub", "url", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	for _, s := range c.symbols {
		msg := map[string]interface{}{
			"type":   "subscribe",
			"symbol": s,
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			return err
		}
		slog.Info("Subscribed to symbol", "symbol", s)
	}

	go c.readLoop()
	return nil
}

func (c *Client) readLoop() {
	defer c.conn.Close()

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				// Only log error if not closing
				select {
				case <-c.done:
					return
				default:
					slog.Error("WebSocket read error", "error", err)
					return
				}
			}

			var response FinnhubResponse
			if err := json.Unmarshal(message, &response); err != nil {
				slog.Error("JSON parse error", "error", err)
				continue
			}

			if response.Type == "trade" {
				c.processTrades(response.Data)
			}
		}
	}
}

func (c *Client) processTrades(trades []TradeData) {
	for _, trade := range trades {
		slog.Debug("Processing tick", "symbol", trade.Symbol, "price", trade.Price)

		tick := &stock.StockTick{
			Symbol:    trade.Symbol,
			Price:     trade.Price,
			Timestamp: trade.Timestamp,
		}

		bytes, err := proto.Marshal(tick)
		if err != nil {
			slog.Error("Protobuf marshal error", "error", err)
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: "market_ticks",
			Key:   sarama.StringEncoder(trade.Symbol),
			Value: sarama.ByteEncoder(bytes),
		}

		partition, offset, err := c.producer.SendMessage(msg)
		if err != nil {
			slog.Error("Kafka send error", "error", err)
		} else {
			slog.Debug("Message sent", "partition", partition, "offset", offset)
		}
	}
}

// Close shuts down the client.
func (c *Client) Close() error {
	close(c.done)

	if c.conn != nil {
		// Send close message
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}

	if c.producer != nil {
		return c.producer.Close()
	}
	return nil
}
