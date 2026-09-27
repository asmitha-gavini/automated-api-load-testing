package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleHealth(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" || resp["service"] != "demo-mock-api" {
		t.Errorf("unexpected body: %v", resp)
	}
}

func TestHandleUsers(t *testing.T) {
	// Use small delay to keep unit tests fast
	req, _ := http.NewRequest(http.MethodGet, "/api/users?delay_ms=0", nil)
	w := httptest.NewRecorder()
	handleUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var users []User
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to decode users: %v", err)
	}

	if len(users) == 0 {
		t.Errorf("expected non-empty user list")
	}
}

func TestHandleProducts(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/products?delay_ms=0", nil)
	w := httptest.NewRecorder()
	handleProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var products []Product
	if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
		t.Fatalf("failed to decode products: %v", err)
	}

	if len(products) == 0 {
		t.Errorf("expected non-empty product list")
	}
}

func TestHandleOrders_Success(t *testing.T) {
	payload := OrderRequest{
		UserID:    1,
		ProductID: 101,
		Quantity:  2,
		Notes:     "Test order",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/orders?delay_ms=0", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleOrders(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp OrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode order response: %v", err)
	}

	if resp.OrderID == "" || resp.Status != "confirmed" || resp.Quantity != 2 {
		t.Errorf("unexpected order response data: %+v", resp)
	}
}

func TestHandleOrders_ValidationErrors(t *testing.T) {
	// Case 1: Bad quantity
	badPayload := OrderRequest{UserID: 1, ProductID: 101, Quantity: 0}
	body, _ := json.Marshal(badPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/orders?delay_ms=0", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for 0 quantity, got %d", w.Code)
	}

	// Case 2: Exceeding stock limit
	hugePayload := OrderRequest{UserID: 1, ProductID: 101, Quantity: 9999}
	body, _ = json.Marshal(hugePayload)
	req, _ = http.NewRequest(http.MethodPost, "/api/orders?delay_ms=0", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handleOrders(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422 for excessive quantity, got %d", w.Code)
	}
}

func TestDelaySimulation(t *testing.T) {
	start := time.Now()
	req, _ := http.NewRequest(http.MethodGet, "/api/users?delay_ms=30", nil)
	w := httptest.NewRecorder()
	handleUsers(w, req)

	elapsed := time.Since(start)
	if elapsed < 25*time.Millisecond {
		t.Errorf("expected request to take at least 25ms, took %v", elapsed)
	}
}
