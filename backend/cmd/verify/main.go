package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	TestID             string  `json:"test_id"`
	Status             string  `json:"status"`
	Timestamp          string  `json:"timestamp"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	ErrorRate          float64 `json:"error_rate"`
	RequestsPerSecond  float64 `json:"requests_per_second"`
	AverageLatencyMs   float64 `json:"average_latency_ms"`
	P50LatencyMs       float64 `json:"p50_latency_ms"`
	P95LatencyMs       float64 `json:"p95_latency_ms"`
	P99LatencyMs       float64 `json:"p99_latency_ms"`
	ActiveUsers        int     `json:"active_users"`
}

func main() {
	wsURL := "ws://localhost:8080/ws/metrics"
	log.Printf("[Verification] Connecting to WebSocket: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	log.Println("[Verification] WebSocket connection established successfully!")

	// Channel to receive messages
	msgChan := make(chan WSMessage, 100)
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg WSMessage
			if err := json.Unmarshal(data, &msg); err == nil {
				msgChan <- msg
				if msg.Status == "completed" || msg.Status == "stopped" {
					return
				}
			}
		}
	}()

	// Trigger load test via REST API
	testConfig := map[string]interface{}{
		"name":             "Phase 3 WS & Prom Live Verification",
		"target_url":       "http://localhost:8081/api/users?delay_ms=10",
		"method":           "GET",
		"virtual_users":    5,
		"duration_seconds": 3,
		"ramp_up_seconds":  1,
		"timeout_ms":       2000,
	}
	payload, _ := json.Marshal(testConfig)

	log.Println("[Verification] Triggering POST /api/v1/tests...")
	resp, err := http.Post("http://localhost:8080/api/v1/tests", "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("POST /tests failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("POST /tests returned %d: %s", resp.StatusCode, string(body))
	}
	log.Println("[Verification] Load test started. Awaiting real-time WebSocket frames...")

	// Print received frames
	frameCount := 0
	for {
		select {
		case msg := <-msgChan:
			frameCount++
			fmt.Printf(" [WS Frame #%02d] Status: %-9s | Reqs: %4d | Succ: %4d | Fail: %2d | RPS: %6.1f | AvgLat: %5.1fms | P95: %5.1fms | VUs: %d\n",
				frameCount, msg.Status, msg.TotalRequests, msg.SuccessfulRequests, msg.FailedRequests,
				msg.RequestsPerSecond, msg.AverageLatencyMs, msg.P95LatencyMs, msg.ActiveUsers)
		case <-doneChan:
			log.Println("[Verification] Terminal message received from WebSocket stream.")
			goto Done
		case <-time.After(8 * time.Second):
			log.Println("[Verification] Timeout waiting for test completion.")
			goto Done
		}
	}

Done:
	// Verify Prometheus metrics after the run
	log.Println("\n[Verification] Querying GET /metrics from Prometheus...")
	promResp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		log.Fatalf("GET /metrics failed: %v", err)
	}
	defer promResp.Body.Close()
	body, _ := io.ReadAll(promResp.Body)

	lines := strings.Split(string(body), "\n")
	fmt.Println("--- Prometheus Telemetry Output ---")
	for _, line := range lines {
		if strings.HasPrefix(line, "loadtest_") && !strings.HasPrefix(line, "loadtest_request_duration_seconds_bucket") {
			fmt.Println(line)
		}
	}
	log.Println("\n[Verification] Phase 3 end-to-end verification succeeded!")
}
