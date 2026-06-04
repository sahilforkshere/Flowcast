package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gorilla/websocket"
)

type TopPage struct {
	PageURL string `json:"page_url"`
	Views   uint64 `json:"views"`
}

type ViewsPerMinute struct {
	Minute string `json:"minute"`
	Views  uint64 `json:"views"`
}

type DashboardStats struct {
	ActiveUsers    uint64           `json:"active_users"`
	TopPages       []TopPage        `json:"top_pages"`
	ViewsPerMinute []ViewsPerMinute `json:"views_per_minute"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

var conn clickhouse.Conn

func main() {
	var err error
	conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{Database: "analytics"},
	})
	if err != nil {
		log.Fatalf("clickhouse connect: %v", err)
	}
	defer conn.Close()

	go broadcastLoop()

	http.HandleFunc("/ws", handleWS)

	port := os.Getenv("WS_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("WebSocket server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()
	log.Printf("client connected, total: %d", len(clients))

	// block until client disconnects
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}

	clientsMu.Lock()
	delete(clients, c)
	clientsMu.Unlock()
	c.Close()
	log.Printf("client disconnected, total: %d", len(clients))
}

func broadcastLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		stats, err := queryStats()
		if err != nil {
			log.Printf("query error: %v", err)
			continue
		}
		payload, _ := json.Marshal(stats)

		clientsMu.Lock()
		for c := range clients {
			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				c.Close()
				delete(clients, c)
			}
		}
		clientsMu.Unlock()
	}
}

func queryStats() (*DashboardStats, error) {
	ctx := context.Background()
	stats := &DashboardStats{}

	// active users in last 5 minutes
	row := conn.QueryRow(ctx,
		`SELECT uniqExact(session_id) FROM analytics.pageviews
		 WHERE event_time >= now() - INTERVAL 5 MINUTE`)
	if err := row.Scan(&stats.ActiveUsers); err != nil {
		return nil, err
	}

	// top 5 pages in last 1 hour
	rows, err := conn.Query(ctx,
		`SELECT page_url, count() AS views FROM analytics.pageviews
		 WHERE event_time >= now() - INTERVAL 1 HOUR
		 GROUP BY page_url ORDER BY views DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p TopPage
		if err := rows.Scan(&p.PageURL, &p.Views); err != nil {
			return nil, err
		}
		stats.TopPages = append(stats.TopPages, p)
	}

	// views per minute for last 30 minutes
	rows2, err := conn.Query(ctx,
		`SELECT toStartOfMinute(event_time) AS minute, count() AS views
		 FROM analytics.pageviews
		 WHERE event_time >= now() - INTERVAL 30 MINUTE
		 GROUP BY minute ORDER BY minute ASC`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var v ViewsPerMinute
		var t time.Time
		if err := rows2.Scan(&t, &v.Views); err != nil {
			return nil, err
		}
		v.Minute = t.Format("15:04")
		stats.ViewsPerMinute = append(stats.ViewsPerMinute, v)
	}

	return stats, nil
}
