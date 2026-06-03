package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	kafka "github.com/segmentio/kafka-go"
)

type PageviewEvent struct {
	SessionID  string `json:"session_id"`
	UserID     string `json:"user_id"`
	PageURL    string `json:"page_url"`
	Referrer   string `json:"referrer"`
	Country    string `json:"country"`
	DeviceType string `json:"device_type"`
	DurationMS int    `json:"duration_ms"`
}

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	dsn := os.Getenv("CLICKHOUSE_DSN")
	if dsn == "" {
		dsn = "clickhouse://default:@localhost:9000/analytics"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{Database: "analytics"},
	})
	if err != nil {
		log.Fatalf("clickhouse connect: %v", err)
	}
	defer conn.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "pageview-events",
		GroupID: "clickhouse-consumer",
	})
	defer reader.Close()

	log.Println("ClickHouse Consumer started")

	batch := make([]PageviewEvent, 0, 5000)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	msgCh := make(chan PageviewEvent, 10000)

	// reader goroutine
	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("kafka read error: %v", err)
				continue
			}
			var event PageviewEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}
			msgCh <- event
		}
	}()

	// micro-batch flush loop
	for {
		select {
		case event := <-msgCh:
			batch = append(batch, event)
			if len(batch) >= 5000 {
				flush(conn, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flush(conn, batch)
				batch = batch[:0]
			}
		}
	}
}

func flush(conn clickhouse.Conn, events []PageviewEvent) {
	ctx := context.Background()
	b, err := conn.PrepareBatch(ctx, "INSERT INTO analytics.pageviews")
	if err != nil {
		log.Printf("prepare batch error: %v", err)
		return
	}

	now := time.Now()
	for _, e := range events {
		if err := b.Append(
			now,
			e.SessionID,
			e.UserID,
			e.PageURL,
			e.Referrer,
			e.Country,
			e.DeviceType,
			uint32(e.DurationMS),
		); err != nil {
			log.Printf("append error: %v", err)
			return
		}
	}

	if err := b.Send(); err != nil {
		log.Printf("batch send error: %v", err)
		return
	}

	log.Printf("inserted %d events", len(events))
}
