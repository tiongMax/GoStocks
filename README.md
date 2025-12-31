# 📈 GoStocks: Real-Time Event-Driven Market Monitor

GoStocks is a high-performance, distributed microservices backend designed to handle real-time financial data streams. It demonstrates the use of event-streaming, low-latency caching, and inter-service communication using industry-standard tools.

## 🏗 Architecture Overview

The system is split into specialized services to ensure scalability and fault tolerance:

* **Ingestor Service:** Connects to external market WebSockets (Finnhub/Alpaca). It acts as a producer, pushing raw market "ticks" into a Kafka topic.
* **Processor Service:** A Kafka consumer that performs real-time calculations (e.g., Price Volatility, Moving Averages) and updates a **Redis Speed Layer** for instant UI lookups.
* **Alert Service:** Manages user-defined price thresholds stored in **PostgreSQL**. It monitors the live stream and triggers notifications when price conditions are met.
* **API Gateway:** The entry point for clients. It provides REST endpoints for price queries (via Redis) and alert management (via gRPC to the Alert Service).

## 🛠 Tech Stack

| Layer | Technology |
| --- | --- |
| **Language** | **Go (Golang)** - Chosen for concurrency and performance |
| **Communication** | **gRPC / Protocol Buffers** (Internal) & **REST / Gin** (Public) |
| **Messaging** | **Apache Kafka** - Distributed message broker for event streaming |
| **Speed Layer** | **Redis** - For sub-millisecond price lookups |
| **Persistence** | **PostgreSQL** - Relational storage for user alerts and profiles |
| **Infrastructure** | **Docker & Docker Compose** - For containerized orchestration |

---

## 🗺 14-Day Development Roadmap

This roadmap is designed to produce a "Working App" at the end of every day.

### **Phase 1: Streaming Infrastructure (Week 1)**

* **Day 1: Scaffolding** – Init Go modules, set up `docker-compose.yml` with Kafka, Redis, and Postgres.
* **Day 2: The Data Contract** – Define `stock.proto` and generate Go code for Protobuf structs.
* **Day 3: Live Ingestion** – Build the Ingestor Service to connect to a WebSocket (Finnhub) and log prices.
* **Day 4: Kafka Integration** – Ingestor Service now publishes JSON/Protobuf messages to the Kafka topic.
* **Day 5: Processor Service** – Build a consumer that reads from Kafka and logs throughput metrics.
* **Day 6: The Speed Layer** – Processor service updates Redis keys (`price:AAPL`) for every incoming tick.
* **Day 7: E2E Verification** – A small CLI tool that queries Redis to verify real-time flow from Web → Kafka → Redis.

### **Phase 2: Logic, Persistence & APIs (Week 2)**

* **Day 8: Database Design** – Create Postgres schema for `users` and `alerts` (symbol, target_price, condition).
* **Day 9: Alert Service (gRPC)** – Implement the gRPC server to handle `CreateAlert` and `GetAlerts` calls.
* **Day 10: Trigger Logic** – Alert Service consumes Kafka; checks every tick against the Postgres database.
* **Day 11: The Gateway** – Build a Gin-based REST API to expose `GET /price/:symbol` (hitting Redis).
* **Day 12: Command Flow** – Connect Gateway to Alert Service via gRPC for `POST /alerts`.
* **Day 13: Error Handling** – Implement Graceful Shutdown and Kafka "Consumer Group" logic for fault tolerance.
* **Day 14: Final Polish** – Add structured logging (Zap/Slog), write unit tests for the Processor logic, and record a demo.

---

## 🚦 Getting Started

1. **Clone the repo:** `git clone https://github.com/youruser/gostocks`
2. **Launch Infra:** `docker-compose up -d`
3. **Run Ingestor:** `go run cmd/ingestor/main.go`
4. **Run Processor:** `go run cmd/processor/main.go`

---

### **Engineering Challenges Solved (Resume Talking Points)**

* **Handling Backpressure:** Using Kafka as a buffer to ensure the Processor doesn't crash during market volatility.
* **State Management:** Balancing ephemeral data (Redis) with persistent data (Postgres).
* **Schema Evolution:** Using Protobuf to ensure forward/backward compatibility between microservices.

