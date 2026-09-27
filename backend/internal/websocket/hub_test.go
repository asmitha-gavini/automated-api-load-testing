package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func setupTestWSServer(t *testing.T) (*Hub, *httptest.Server) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testID := r.URL.Query().Get("test_id")
		ServeWS(hub, w, r, testID)
	}))

	return hub, server
}

func TestWebSocket_ConnectionAndMetricStreaming(t *testing.T) {
	hub, server := setupTestWSServer(t)
	defer server.Close()
	defer hub.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?test_id=test-123"

	// Connect Client 1 (subscribed to test-123)
	conn1, resp1, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect client 1: %v", err)
	}
	defer conn1.Close()
	_ = resp1.Body.Close()

	// Connect Client 2 (subscribed to all / wildcard)
	wsAllURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn2, resp2, err := websocket.DefaultDialer.Dial(wsAllURL, nil)
	if err != nil {
		t.Fatalf("failed to connect client 2: %v", err)
	}
	defer conn2.Close()
	_ = resp2.Body.Close()

	// Wait briefly for registration to process in hub
	time.Sleep(50 * time.Millisecond)

	if clients := hub.ConnectedClients(); clients != 2 {
		t.Errorf("expected 2 connected clients, got %d", clients)
	}

	// 1. Broadcast live metric update
	metricPayload := WSMetricMessage{
		TestID:             "test-123",
		Status:             "running",
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		TotalRequests:      150,
		SuccessfulRequests: 145,
		FailedRequests:     5,
		ErrorRate:          3.33,
		RequestsPerSecond:  75.5,
		AverageLatencyMs:   18.2,
		P50LatencyMs:       15.0,
		P95LatencyMs:       28.5,
		P99LatencyMs:       45.0,
		ActiveUsers:        5,
	}

	if err := hub.BroadcastToTest("test-123", metricPayload); err != nil {
		t.Fatalf("BroadcastToTest failed: %v", err)
	}

	// Read and verify on Client 1
	_ = conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("client 1 failed to read message: %v", err)
	}

	var received1 WSMetricMessage
	if err := json.Unmarshal(msg1, &received1); err != nil {
		t.Fatalf("failed to unmarshal client 1 message: %v", err)
	}
	if received1.TestID != "test-123" || received1.Status != "running" || received1.TotalRequests != 150 {
		t.Errorf("client 1 received incorrect payload: %+v", received1)
	}

	// Read and verify on Client 2 (wildcard subscriber)
	_ = conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("client 2 failed to read message: %v", err)
	}

	var received2 WSMetricMessage
	if err := json.Unmarshal(msg2, &received2); err != nil {
		t.Fatalf("failed to unmarshal client 2 message: %v", err)
	}
	if received2.TestID != "test-123" {
		t.Errorf("client 2 received incorrect test ID: %s", received2.TestID)
	}

	// 2. Broadcast completion message
	completionPayload := WSMetricMessage{
		TestID:             "test-123",
		Status:             "completed",
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		TotalRequests:      500,
		SuccessfulRequests: 490,
		FailedRequests:     10,
		ActiveUsers:        0,
	}
	_ = hub.BroadcastToTest("test-123", completionPayload)

	_ = conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, compMsg, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("client 1 failed to read completion message: %v", err)
	}

	var finalMsg WSMetricMessage
	if err := json.Unmarshal(compMsg, &finalMsg); err != nil {
		t.Fatalf("failed to unmarshal final message: %v", err)
	}
	if finalMsg.Status != "completed" || finalMsg.TotalRequests != 500 {
		t.Errorf("expected completed status with 500 requests, got %+v", finalMsg)
	}

	// 3. Test cleanup on client disconnect
	_ = conn1.Close()
	_ = conn2.Close()
	time.Sleep(100 * time.Millisecond)

	if remaining := hub.ConnectedClients(); remaining != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", remaining)
	}
}

func TestWebSocket_SubscriptionFiltering(t *testing.T) {
	hub, server := setupTestWSServer(t)
	defer server.Close()
	defer hub.Close()

	// Client A subscribed to test-AAA
	urlA := "ws" + strings.TrimPrefix(server.URL, "http") + "?test_id=test-AAA"
	connA, _, err := websocket.DefaultDialer.Dial(urlA, nil)
	if err != nil {
		t.Fatalf("failed to connect client A: %v", err)
	}
	defer connA.Close()

	// Client B subscribed to test-BBB
	urlB := "ws" + strings.TrimPrefix(server.URL, "http") + "?test_id=test-BBB"
	connB, _, err := websocket.DefaultDialer.Dial(urlB, nil)
	if err != nil {
		t.Fatalf("failed to connect client B: %v", err)
	}
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	// Broadcast exclusively to test-AAA
	msgA := WSMetricMessage{TestID: "test-AAA", Status: "running", TotalRequests: 10}
	_ = hub.BroadcastToTest("test-AAA", msgA)

	// Client A should receive
	_ = connA.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rawA, err := connA.ReadMessage()
	if err != nil {
		t.Fatalf("client A failed to read message: %v", err)
	}
	var resA WSMetricMessage
	_ = json.Unmarshal(rawA, &resA)
	if resA.TestID != "test-AAA" {
		t.Errorf("expected test-AAA, got %s", resA.TestID)
	}

	// Client B should NOT receive anything (timeout expected)
	_ = connB.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = connB.ReadMessage()
	if err == nil {
		t.Errorf("expected timeout for client B, but received a message")
	}
}

func TestWebSocket_ConcurrentBroadcast(t *testing.T) {
	hub, server := setupTestWSServer(t)
	defer server.Close()
	defer hub.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	// Concurrently broadcast 20 messages from different goroutines
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = hub.BroadcastToTest("test-conc", WSMetricMessage{
				TestID:        "test-conc",
				TotalRequests: int64(id),
			})
		}(i)
	}
	wg.Wait()

	// Drain messages
	for i := 0; i < 20; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read message %d during concurrent broadcast: %v", i, err)
		}
	}
}
