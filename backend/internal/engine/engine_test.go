package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"backend/internal/database"
)

func TestEngine_Lifecycle_TotalRequests(t *testing.T) {
	// Setup target mock server
	var hits int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Setup temporary SQLite DB
	tempDir, err := os.MkdirTemp("", "engine_test_db_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := database.InitDB(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	eng := NewEngine(db, nil, nil)

	cfg := TestConfig{
		Name:          "Test-Lifecycle",
		TargetURL:     server.URL,
		Method:        "GET",
		VirtualUsers:  4,
		TotalRequests: 40,
		TimeoutMs:     2000,
	}

	startedCfg, err := eng.StartTest(cfg)
	if err != nil {
		t.Fatalf("StartTest failed: %v", err)
	}

	// Verify status is running or rapidly progressing
	status, _, _ := eng.GetStatus()
	if status != StatusRunning && status != StatusCompleted {
		t.Errorf("expected StatusRunning or StatusCompleted, got %s", status)
	}

	// Wait for test to conclude
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s, _, _ := eng.GetStatus()
		if s == StatusCompleted {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	finalStatus, _, _ := eng.GetStatus()
	if finalStatus != StatusCompleted {
		t.Fatalf("expected test to complete, got status: %s", finalStatus)
	}

	// Check collected metrics
	metrics := eng.GetMetrics()
	if metrics.TotalRequests < 40 {
		t.Errorf("expected at least 40 requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests < 40 {
		t.Errorf("expected 40 successful requests, got %d", metrics.SuccessfulRequests)
	}
	if metrics.Latency.AvgMs <= 0 {
		t.Errorf("expected AvgMs > 0, got %f", metrics.Latency.AvgMs)
	}

	// Verify persistence in SQLite
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	savedRun, err := db.GetTestRun(ctx, startedCfg.ID)
	if err != nil {
		t.Fatalf("failed to get saved test run: %v", err)
	}
	if savedRun == nil {
		t.Fatalf("expected test run %s to be saved in database", startedCfg.ID)
	}
	if savedRun.Status != string(StatusCompleted) {
		t.Errorf("expected saved status 'completed', got '%s'", savedRun.Status)
	}
}

func TestEngine_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eng := NewEngine(nil, nil, nil)

	cfg := TestConfig{
		TargetURL:       server.URL,
		Method:          "GET",
		VirtualUsers:    5,
		DurationSeconds: 15, // Long duration
		TimeoutMs:       2000,
	}

	_, err := eng.StartTest(cfg)
	if err != nil {
		t.Fatalf("StartTest failed: %v", err)
	}

	// Wait a moment for workers to start running
	time.Sleep(50 * time.Millisecond)

	status, _, _ := eng.GetStatus()
	if status != StatusRunning {
		t.Fatalf("expected status running before cancellation, got %s", status)
	}

	// Stop the test
	if err := eng.StopTest(); err != nil {
		t.Fatalf("StopTest failed: %v", err)
	}

	finalStatus, _, _ := eng.GetStatus()
	if finalStatus != StatusStopped {
		t.Errorf("expected status 'stopped', got '%s'", finalStatus)
	}
}

func TestEngine_PreventDuplicateStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	eng := NewEngine(nil, nil, nil)

	cfg := TestConfig{
		TargetURL:       server.URL,
		Method:          "GET",
		VirtualUsers:    2,
		DurationSeconds: 10,
		TimeoutMs:       2000,
	}

	_, err := eng.StartTest(cfg)
	if err != nil {
		t.Fatalf("first StartTest failed: %v", err)
	}
	defer eng.StopTest()

	// Attempt second start while first is running
	_, err = eng.StartTest(cfg)
	if err == nil {
		t.Errorf("expected duplicate StartTest to return error, got nil")
	}
}
