package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/database"
	"backend/internal/engine"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *engine.Engine, *httptest.Server) {
	gin.SetMode(gin.TestMode)

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	eng := engine.NewEngine(nil, nil, nil)
	handler := NewTestHandler(eng, nil)

	r := gin.New()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/tests", handler.StartTest)
		v1.GET("/tests/status", handler.GetStatus)
		v1.GET("/tests/metrics", handler.GetMetrics)
		v1.POST("/tests/stop", handler.StopTest)
	}

	return r, eng, targetServer
}

func TestTestHandler_StartTest_ValidAndInvalid(t *testing.T) {
	router, eng, targetServer := setupTestRouter()
	defer targetServer.Close()
	defer eng.StopTest()

	// 1. Invalid payload: bad URL
	badPayload := map[string]interface{}{
		"target_url":     "invalid-url",
		"method":         "GET",
		"virtual_users":  5,
		"total_requests": 10,
	}
	body, _ := json.Marshal(badPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for bad URL, got %d", w.Code)
	}

	// 2. Valid payload
	goodPayload := map[string]interface{}{
		"target_url":       targetServer.URL,
		"method":           "GET",
		"virtual_users":    2,
		"duration_seconds": 5,
		"timeout_ms":       1000,
	}
	body, _ = json.Marshal(goodPayload)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/tests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for valid config, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTestHandler_GetStatusAndMetrics(t *testing.T) {
	router, _, targetServer := setupTestRouter()
	defer targetServer.Close()

	// Query status
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tests/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for status, got %d", w.Code)
	}

	// Query metrics
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for metrics, got %d", w.Code)
	}
}

func TestTestHandler_StopTest(t *testing.T) {
	router, _, targetServer := setupTestRouter()
	defer targetServer.Close()

	// 1. Stop when nothing running
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tests/stop", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when stopping with no active test, got %d", w.Code)
	}

	// 2. Start a test
	startPayload := map[string]interface{}{
		"target_url":       targetServer.URL,
		"method":           "GET",
		"virtual_users":    2,
		"duration_seconds": 10,
	}
	body, _ := json.Marshal(startPayload)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/tests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to start test: %d", w.Code)
	}

	// 3. Stop running test
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/tests/stop", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on stopping active test, got %d", w.Code)
	}
}

func TestTestHandler_Phase5_HistoryAndReports(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup SQLite DB
	tempDir, err := os.MkdirTemp("", "test_handler_phase5_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	eng := engine.NewEngine(db, nil, nil)
	handler := NewTestHandler(eng, db)

	r := gin.New()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/tests", handler.ListTests)
		v1.DELETE("/tests", handler.ClearAllTests)
		v1.GET("/tests/summary", handler.GetPerformanceSummary)
		v1.GET("/tests/:id", handler.GetTestByID)
		v1.DELETE("/tests/:id", handler.DeleteTest)
		v1.GET("/tests/:id/export", handler.ExportTestReport)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	// Seed test runs
	run1 := &database.TestRun{
		ID:                 "hist-1",
		Name:               "Baseline Load Test",
		TargetURL:          "http://localhost:8081/api/users",
		Method:             "GET",
		VirtualUsers:       2,
		DurationSeconds:    5,
		TotalRequests:      300,
		SuccessfulRequests: 300,
		FailedRequests:     0,
		ErrorRate:          0.0,
		AvgLatencyMs:       16.5,
		MinLatencyMs:       15.0,
		MaxLatencyMs:       45.0,
		P50LatencyMs:       16.2,
		P90LatencyMs:       17.1,
		P95LatencyMs:       18.0,
		P99LatencyMs:       25.0,
		Status:             "completed",
		StartedAt:          now.Add(-10 * time.Minute),
		CompletedAt:        &now,
	}

	run2 := &database.TestRun{
		ID:                 "hist-2",
		Name:               "Failure Injection Test",
		TargetURL:          "http://localhost:8081/api/users?error_rate=0.5",
		Method:             "GET",
		VirtualUsers:       4,
		DurationSeconds:    5,
		TotalRequests:      200,
		SuccessfulRequests: 100,
		FailedRequests:     100,
		ErrorRate:          50.0,
		AvgLatencyMs:       22.0,
		MinLatencyMs:       15.0,
		MaxLatencyMs:       80.0,
		P50LatencyMs:       20.0,
		P90LatencyMs:       30.0,
		P95LatencyMs:       40.0,
		P99LatencyMs:       60.0,
		Status:             "completed",
		StartedAt:          now.Add(-5 * time.Minute),
		CompletedAt:        &now,
	}

	_ = db.SaveTestRun(ctx, run1)
	_ = db.SaveTestRun(ctx, run2)

	// 1. GET /api/v1/tests (List)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tests", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for ListTests, got %d", w.Code)
	}

	var histResp database.TestHistoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &histResp); err != nil {
		t.Fatalf("failed to decode history response: %v", err)
	}
	if histResp.Total != 2 || len(histResp.Tests) != 2 {
		t.Errorf("expected 2 tests in history, got total=%d, len=%d", histResp.Total, len(histResp.Tests))
	}
	if histResp.Summary.TotalTests != 2 || histResp.Summary.TotalRequests != 500 {
		t.Errorf("unexpected summary in history response: %+v", histResp.Summary)
	}

	// 2. GET /api/v1/tests/summary
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/summary", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GetPerformanceSummary, got %d", w.Code)
	}

	// 3. GET /api/v1/tests/:id
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/hist-1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GetTestByID, got %d", w.Code)
	}

	// 4. GET /api/v1/tests/:id/export?format=json
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/hist-1/export?format=json", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for JSON export, got %d", w.Code)
	}

	// 5. GET /api/v1/tests/:id/export?format=csv
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/hist-1/export?format=csv", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Total Requests,300") {
		t.Fatalf("expected 200 with valid CSV content, got code %d: %s", w.Code, w.Body.String())
	}

	// 6. GET /api/v1/tests/:id/export?format=md
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tests/hist-1/export?format=md", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "API Load Test Report") {
		t.Fatalf("expected 200 with Markdown content, got code %d: %s", w.Code, w.Body.String())
	}

	// 7. DELETE /api/v1/tests/:id
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/tests/hist-2", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for DeleteTest, got %d", w.Code)
	}

	// 8. DELETE /api/v1/tests (Clear all)
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/tests", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for ClearAllTests, got %d", w.Code)
	}
}

