package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// User represents a mock user profile.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Product represents a mock store catalog item.
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	InStock  bool    `json:"in_stock"`
}

// OrderRequest is the JSON payload submitted to create an order.
type OrderRequest struct {
	UserID    int    `json:"user_id"`
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Notes     string `json:"notes,omitempty"`
}

// OrderResponse is the response returned upon order creation.
type OrderResponse struct {
	OrderID     string    `json:"order_id"`
	UserID      int       `json:"user_id"`
	ProductID   int       `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Global mock datasets
var mockUsers = []User{
	{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Role: "engineer", CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Role: "product_manager", CreatedAt: time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)},
	{ID: 3, Name: "Charlie Davis", Email: "charlie@example.com", Role: "designer", CreatedAt: time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)},
	{ID: 4, Name: "Diana Prince", Email: "diana@example.com", Role: "architect", CreatedAt: time.Date(2025, 4, 18, 0, 0, 0, 0, time.UTC)},
	{ID: 5, Name: "Evan Wright", Email: "evan@example.com", Role: "devops", CreatedAt: time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC)},
}

var mockProducts = []Product{
	{ID: 101, Name: "Mechanical Keyboard", Category: "Electronics", Price: 129.99, Stock: 45, InStock: true},
	{ID: 102, Name: "Ergonomic Mouse", Category: "Electronics", Price: 79.50, Stock: 80, InStock: true},
	{ID: 103, Name: "4K USB-C Monitor", Category: "Displays", Price: 499.00, Stock: 15, InStock: true},
	{ID: 104, Name: "Noise-Cancelling Headphones", Category: "Audio", Price: 249.99, Stock: 30, InStock: true},
	{ID: 105, Name: "Aluminum Laptop Stand", Category: "Accessories", Price: 39.95, Stock: 110, InStock: true},
}

func main() {
	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/products", handleProducts)
	mux.HandleFunc("/api/orders", handleOrders)

	handler := loggingAndCorsMiddleware(mux)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Demo Mock API listening on http://localhost%s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  - GET/POST/PUT/DELETE /api/users     (fast baseline, ~15ms)")
	log.Printf("  - GET/POST/PUT/DELETE /api/products  (moderate with jitter, ~50ms)")
	log.Printf("  - GET/POST            /api/orders    (transaction simulation, ~70ms)")
	log.Printf("Optional query params:")
	log.Printf("  - ?delay_ms=<int>       (custom latency in milliseconds)")
	log.Printf("  - ?error_rate=<0.0-1.0> (probability of simulating HTTP 500)")

	if err := http.ListenAndServe(addr, handler); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Demo API failed: %v", err)
	}
}

func loggingAndCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applyArtificialDelay(r *http.Request, defaultDelay time.Duration, jitterMax time.Duration) time.Duration {
	delay := defaultDelay

	q := r.URL.Query()
	if val := q.Get("delay_ms"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			delay = time.Duration(parsed) * time.Millisecond
		}
	} else if val := q.Get("delay"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil && parsed >= 0 {
			delay = parsed
		}
	} else if jitterMax > 0 {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(jitterMax)))
		if err == nil {
			delay += time.Duration(nBig.Int64())
		}
	}

	if delay > 0 {
		time.Sleep(delay)
	}
	return delay
}

func checkSimulatedError(r *http.Request, defaultRate float64) bool {
	rate := defaultRate
	if qRate := r.URL.Query().Get("error_rate"); qRate != "" {
		if parsed, err := strconv.ParseFloat(qRate, 64); err == nil && parsed >= 0.0 && parsed <= 1.0 {
			rate = parsed
		}
	}

	if rate <= 0.0 {
		return false
	}
	if rate >= 1.0 {
		return true
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err == nil {
		threshold := int64(rate * 1000)
		return nBig.Int64() < threshold
	}
	return false
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "demo-mock-api",
		"version": "1.0.0",
	})
}

// handleUsers returns or updates users with baseline delay. Supports GET, POST, PUT, DELETE.
func handleUsers(w http.ResponseWriter, r *http.Request) {
	actualDelay := applyArtificialDelay(r, 15*time.Millisecond, 10*time.Millisecond)

	if checkSimulatedError(r, 0.0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Simulated upstream database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Simulated-Delay-Ms", fmt.Sprintf("%d", actualDelay.Milliseconds()))

	switch r.Method {
	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         6,
			"name":       "New User",
			"status":     "created",
			"created_at": time.Now().UTC(),
		})
	case http.MethodPut, http.MethodPatch:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         1,
			"status":     "updated",
			"updated_at": time.Now().UTC(),
		})
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted": true,
			"status":  "deleted",
		})
	default: // GET
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockUsers)
	}
}

// handleProducts returns or modifies product catalog items. Supports GET, POST, PUT, DELETE.
func handleProducts(w http.ResponseWriter, r *http.Request) {
	actualDelay := applyArtificialDelay(r, 40*time.Millisecond, 25*time.Millisecond)

	if checkSimulatedError(r, 0.0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Simulated inventory service failure"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Simulated-Delay-Ms", fmt.Sprintf("%d", actualDelay.Milliseconds()))

	switch r.Method {
	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       106,
			"name":     "New Product",
			"price":    99.99,
			"in_stock": true,
		})
	case http.MethodPut, http.MethodPatch:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     101,
			"status": "updated",
		})
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted": true,
		})
	default: // GET
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockProducts)
	}
}

// handleOrders processes order placement or retrieval requests.
func handleOrders(w http.ResponseWriter, r *http.Request) {
	actualDelay := applyArtificialDelay(r, 60*time.Millisecond, 20*time.Millisecond)

	if checkSimulatedError(r, 0.0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Simulated payment gateway timeout"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Simulated-Delay-Ms", fmt.Sprintf("%d", actualDelay.Milliseconds()))

	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"order_id": "ORD-INIT01", "total": 129.99, "status": "confirmed"},
			{"order_id": "ORD-INIT02", "total": 499.00, "status": "confirmed"},
		})
		return
	}

	// For POST: Decode optional body or fallback to default order
	var req OrderRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Quantity <= 0 && (req.UserID > 0 || req.ProductID > 0) {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Quantity must be greater than zero"})
				return
			}
		}
	}

	userID := req.UserID
	if userID <= 0 {
		userID = 1
	}
	productID := req.ProductID
	if productID <= 0 {
		productID = 101
	}
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	if qty > 500 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Requested quantity exceeds available stock limit of 500"})
		return
	}

	var unitPrice float64 = 50.0
	for _, p := range mockProducts {
		if p.ID == productID {
			unitPrice = p.Price
			break
		}
	}

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	orderID := fmt.Sprintf("ORD-%s", strings.ToUpper(hex.EncodeToString(randBytes)))

	resp := OrderResponse{
		OrderID:     orderID,
		UserID:      userID,
		ProductID:   productID,
		Quantity:    qty,
		TotalAmount: float64(qty) * unitPrice,
		Status:      "confirmed",
		CreatedAt:   time.Now().UTC(),
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
