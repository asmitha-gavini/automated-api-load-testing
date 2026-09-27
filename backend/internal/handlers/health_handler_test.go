package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"backend/internal/database"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup temporary database
	tempDir, err := os.MkdirTemp("", "test_health_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "health_test.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	handler := NewHealthHandler(db)

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", resp["status"])
	}
	if resp["database"] != "connected" {
		t.Errorf("expected database 'connected', got '%v'", resp["database"])
	}
	if resp["service"] != "automated-api-load-testing-backend" {
		t.Errorf("expected correct service name, got '%v'", resp["service"])
	}
}

func TestHealthCheck_Degraded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Handler with nil DB
	handler := NewHealthHandler(nil)

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["status"] != "degraded" {
		t.Errorf("expected status 'degraded', got '%v'", resp["status"])
	}
}
