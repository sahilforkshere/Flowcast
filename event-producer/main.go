package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

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

var writer *kafka.Writer

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	writer = &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        "pageview-events",
		Balancer:     &kafka.Hash{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
	}
	defer writer.Close()

	http.HandleFunc("/track", handleTrack)
	http.HandleFunc("/health", handleHealth)

	log.Println("Event Producer listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event PageviewEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if event.SessionID == "" || event.PageURL == "" {
		http.Error(w, "session_id and page_url are required", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(event)

	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.SessionID),
		Value: payload,
	})
	if err != nil {
		log.Printf("kafka write error: %v", err)
		http.Error(w, "failed to write event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
