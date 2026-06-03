# Real-Time Analytics Pipeline

A high-throughput analytics pipeline that ingests page-view events, stores them in a columnar database, and streams live metrics to a dashboard over WebSocket. Built as a data engineering project to demonstrate skills beyond standard CRUD applications.

## What it does

Every time a user visits a page, an event is fired. This system captures those events at high speed, stores them in ClickHouse, and shows live charts on a dashboard that update every two seconds — active users, top pages, and views per minute.

## Architecture

```
Browser / k6
     |
     | POST /track
     v
Event Producer (Go, :8080)
     |
     | kafka-go, BatchSize 100
     v
Kafka (pageview-events topic, 3 partitions)
     |
     | consumer group: clickhouse-consumer
     v
ClickHouse Consumer (Go)
     |
     | batch insert, 5000 events or 2s
     v
ClickHouse — analytics.pageviews (MergeTree)
     |
     | query every 2s
     v
WebSocket Server (Go, :8081)
     |
     | JSON push
     v
Next.js Dashboard (Recharts)
```

## Tech stack

| Technology | Role |
|---|---|
| Go | Event Producer, ClickHouse Consumer, WebSocket Server |
| Apache Kafka | Durable message queue decoupling ingestion from storage |
| ClickHouse | Columnar database for sub-100ms analytical queries |
| Next.js + Recharts | Live dashboard |
| Docker Compose | Local infrastructure (Kafka, Zookeeper, ClickHouse) |
| k6 | Load testing — 500 virtual users |

## Project structure

```
/event-producer        Go HTTP server — POST /track writes events to Kafka
/clickhouse-consumer   Go binary — reads Kafka, batch-inserts into ClickHouse
/ws-server             Go WebSocket server — queries ClickHouse, pushes stats
/dashboard             Next.js app — live charts via WebSocket
/k6                    load_test.js — 500 VU load test script
schema.sql             ClickHouse CREATE TABLE statement
docker-compose.yml     Spins up Kafka, Zookeeper, ClickHouse
```

## Getting started

### Prerequisites

- Go 1.22+
- Docker Desktop
- Node.js 20+
- k6

### 1. Start infrastructure

```bash
docker compose up -d
```

### 2. Apply database schema

```bash
docker exec clickhouse clickhouse-client --query "CREATE DATABASE IF NOT EXISTS analytics"

docker exec clickhouse clickhouse-client --query "
CREATE TABLE IF NOT EXISTS analytics.pageviews (
    event_time   DateTime,
    session_id   String,
    user_id      String,
    page_url     String,
    referrer     String,
    country      String,
    device_type  String,
    duration_ms  UInt32
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (event_time, session_id)"
```

### 3. Configure environment

Create a `.env` file (already in `.gitignore`):

```
KAFKA_BROKER=localhost:9092
CLICKHOUSE_DSN=clickhouse://default:@localhost:9000/analytics
```

### 4. Run the services

Open four terminal tabs and run one service per tab:

```bash
# Tab 1 — Event Producer
cd event-producer && go run .

# Tab 2 — ClickHouse Consumer
cd clickhouse-consumer && go run .

# Tab 3 — WebSocket Server
cd ws-server && go run .

# Tab 4 — Dashboard
cd dashboard && npm install && npm run dev
```

### 5. Send a test event

```bash
curl -X POST http://localhost:8080/track \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "sess_001",
    "user_id": "user_1",
    "page_url": "/home",
    "referrer": "google.com",
    "country": "IN",
    "device_type": "mobile",
    "duration_ms": 3200
  }'
```

Expected response: `HTTP 202 Accepted`

### 6. Verify data in ClickHouse

```bash
docker exec clickhouse clickhouse-client \
  --query "SELECT count(*) FROM analytics.pageviews"
```

## Load testing

```bash
k6 run k6/load_test.js
```

The test ramps to 500 virtual users over 30 seconds, holds for 2 minutes, then ramps down. Thresholds: p95 latency under 100ms, error rate under 0.1%.

## Performance results

| Metric | Result |
|---|---|
| Throughput | TBD — fill after load test |
| p95 latency | TBD — fill after load test |
| ClickHouse query latency | TBD — fill after load test |
| Rows at test end | TBD — fill after load test |

## Key design decisions

**Why ClickHouse instead of Postgres?**
Postgres stores data row by row. A query like `SELECT page_url, count(*) FROM pageviews` must read every column of every row even though only one column is needed. ClickHouse stores each column separately on disk, so the same query reads only the `page_url` column — skipping 90% of the data. The same query that takes 4-8 seconds in Postgres runs in under 50ms in ClickHouse at 10M+ rows.

**Why Kafka between the producer and the database?**
The Event Producer responds with 202 immediately after writing to Kafka — it never waits for the database. If the consumer crashes or the database is slow, events are safely queued in Kafka and processed when the consumer recovers. The producer's latency is never affected by database performance.

**Why micro-batching in the consumer?**
Every insert into ClickHouse creates a new "part" on disk. Inserting 5,000 events one by one creates 5,000 parts, which triggers excessive background merging and degrades performance. Collecting 5,000 events and inserting them in a single batch creates one part instead of 5,000.

**Why WebSocket instead of polling?**
With 100 users polling every 2 seconds, that is 50 HTTP requests per second just for dashboard refreshes. WebSocket is a persistent connection — the server pushes data when it is ready. 100 users cost 100 persistent connections and zero polling requests.

## API reference

### POST /track

Ingest a page-view event.

**Request body:**

```json
{
  "session_id": "string (required)",
  "user_id": "string",
  "page_url": "string (required)",
  "referrer": "string",
  "country": "string",
  "device_type": "string",
  "duration_ms": 0
}
```

**Responses:**
- `202 Accepted` — event written to Kafka
- `400 Bad Request` — missing `session_id` or `page_url`

### GET /health

Returns `200 OK`. Used for load balancer health checks.

## Author

Sahil Pal — IIITM Gwalior
