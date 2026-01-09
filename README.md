# 📈 GoStocks: Real-Time Event-Driven Market Monitor

GoStocks is a high-performance, distributed microservices backend designed to handle real-time financial data streams. It demonstrates the use of event-streaming, low-latency caching, and inter-service communication using industry-standard tools.

## 🏗 Architecture Overview

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Finnhub   │──────▶│  Ingestor   │──────▶│    Kafka    │
│  WebSocket  │       │  (Producer) │       │ market_ticks│
└─────────────┘       └─────────────┘       └──────┬──────┘
                                                   │
                            ┌──────────────────────┴──────────────────────┐
                            │                                             │
                   ┌────────▼────────┐                           ┌────────▼────────┐
                   │    Processor    │                           │  Alert Service  │
                   │ (Consumer Group)│                           │ (Consumer Group)│
                   └────────┬────────┘                           └────────┬────────┘
                            │                                             │
                   ┌────────▼────────┐                           ┌────────▼────────┐
                   │     Redis       │                           │   PostgreSQL    │
                   │  (Speed Layer)  │                           │    (Alerts)     │
                   └────────┬────────┘                           └────────┬────────┘
                            │                                             │
                            └──────────────────────┬──────────────────────┘
                                                   │
                                          ┌────────▼────────┐
                                          │   API Gateway   │
                                          │   (REST/Gin)    │
                                          └────────┬────────┘
                                                   │
                                          ┌────────▼────────┐
                                          │     Clients     │
                                          └─────────────────┘
```

The system is split into specialized services:

* **Ingestor Service:** Connects to Finnhub WebSocket and pushes raw market ticks into Kafka.
* **Processor Service:** Consumes from Kafka and updates Redis for instant price lookups.
* **Alert Service:** Consumes from Kafka, checks price conditions, and exposes gRPC API for alert management.
* **API Gateway:** REST API for clients to query prices and manage alerts.

## 🛠 Tech Stack

| Layer | Technology |
| --- | --- |
| **Language** | Go (Golang) |
| **Communication** | gRPC / Protocol Buffers (Internal) & REST / Gin (Public) |
| **Messaging** | Apache Kafka with Consumer Groups |
| **Speed Layer** | Redis |
| **Persistence** | PostgreSQL with GORM |
| **Logging** | Structured JSON logging (slog) |
| **Infrastructure** | Docker & Docker Compose |

## 🚦 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Finnhub API Key (free at [finnhub.io](https://finnhub.io))

### 1. Start Infrastructure

```bash
docker-compose up -d
```

### 2. Set Environment Variables

Create a `.env` file in the project root:

```env
FINNHUB_API_KEY=your_api_key_here
KAFKA_BROKERS=localhost:9092
REDIS_ADDR=localhost:6379
DATABASE_URL=user=user password=password dbname=gostocks sslmode=disable host=127.0.0.1 port=5433
```

### 3. Run Services (in separate terminals)

```bash
# Terminal 1: Alert Service (gRPC + Kafka Consumer)
go run cmd/alert/main.go

# Terminal 2: API Gateway
go run cmd/gateway/main.go

# Terminal 3: Processor (Kafka → Redis)
go run cmd/processor/main.go

# Terminal 4: Ingestor (Finnhub → Kafka)
go run cmd/ingestor/main.go
```

## 📡 API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check |
| `GET` | `/price/:symbol` | Get latest price from Redis |
| `POST` | `/alerts` | Create a new price alert |
| `GET` | `/alerts?user_id=1&active_only=true` | List alerts |

### Examples

```bash
# Get price
curl http://localhost:8080/price/AAPL

# Create alert
curl -X POST http://localhost:8080/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "symbol": "AAPL", "target_price": 150.00, "condition": "ABOVE"}'

# List active alerts for user
curl "http://localhost:8080/alerts?user_id=1&active_only=true"
```

## 🧪 Running Tests

```bash
go test ./...
```

## 📁 Project Structure

```
.
├── cmd/
│   ├── alert/          # Alert Service entry point
│   ├── gateway/        # API Gateway entry point
│   ├── ingestor/       # Ingestor Service entry point
│   └── processor/      # Processor Service entry point
├── internal/
│   ├── alert/          # Alert business logic, gRPC server, Kafka consumer
│   ├── gateway/        # HTTP handlers, Redis & gRPC clients
│   ├── ingestor/       # WebSocket client, Kafka producer
│   └── processor/      # Kafka consumer, Redis updater
├── proto/
│   ├── alert/          # Generated gRPC code for alerts
│   ├── stock/          # Generated Protobuf code for stock ticks
│   ├── alert.proto     # Alert service definition
│   └── stock.proto     # Stock tick message definition
├── docker-compose.yml  # Infrastructure (Kafka, Redis, Postgres)
├── go.mod
└── README.md
```

## 🎯 Engineering Highlights

* **Event-Driven Architecture:** Kafka decouples services for independent scaling and fault tolerance.
* **Consumer Groups:** Multiple instances can share the workload; if one crashes, others take over.
* **Graceful Shutdown:** All services handle SIGINT/SIGTERM for clean resource cleanup.
* **Structured Logging:** JSON logs (slog) for easy parsing by monitoring tools.
* **Protocol Buffers:** Efficient serialization with forward/backward compatibility.
* **Speed Layer Pattern:** Redis for sub-millisecond reads, Postgres for durability.
