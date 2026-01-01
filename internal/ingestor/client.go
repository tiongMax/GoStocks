package ingestor

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	apiKey  string
	conn    *websocket.Conn
	symbols []string
}

func NewClient(apiKey string, symbols []string) *Client {
	return &Client{
		apiKey:  apiKey,
		symbols: symbols,
	}
}

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
			}
		}
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
	return nil
}
