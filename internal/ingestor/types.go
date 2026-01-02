package ingestor

// FinnhubResponse matches the JSON structure from Finnhub's WebSocket
type FinnhubResponse struct {
	Type string      `json:"type"`
	Data []TradeData `json:"data"`
}

type TradeData struct {
	Symbol    string  `json:"s"` // Symbol
	Price     float64 `json:"p"` // Last Price
	Timestamp int64   `json:"t"` // UNIX milliseconds timestamp
	Volume    float64 `json:"v"` // Volume
}

