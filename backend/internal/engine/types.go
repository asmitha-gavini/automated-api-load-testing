package engine

import "time"

// TestStatus represents the current lifecycle state of a load test.
type TestStatus string

const (
	StatusIdle      TestStatus = "idle"
	StatusPending   TestStatus = "pending"
	StatusRunning   TestStatus = "running"
	StatusCompleted TestStatus = "completed"
	StatusStopped   TestStatus = "stopped"
	StatusFailed    TestStatus = "failed"
)

// TestConfig contains all user-configurable parameters for a load test run.
type TestConfig struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	TargetURL       string            `json:"target_url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers,omitempty"`
	QueryParams     map[string]string `json:"query_params,omitempty"`
	Body            string            `json:"body,omitempty"`
	VirtualUsers    int               `json:"virtual_users"`
	DurationSeconds int               `json:"duration_seconds,omitempty"`
	TotalRequests   int64             `json:"total_requests,omitempty"`
	RampUpSeconds   int               `json:"ramp_up_seconds,omitempty"`
	TimeoutMs       int               `json:"timeout_ms"`
	CreatedAt       time.Time         `json:"created_at"`
}

// RequestResult captures the execution metrics for a single HTTP request.
type RequestResult struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	DurationMs float64   `json:"duration_ms"`
	StatusCode int       `json:"status_code"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// LatencyStats contains calculated statistical latency percentiles and boundaries in milliseconds.
type LatencyStats struct {
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	AvgMs float64 `json:"avg_ms"`
	P50Ms float64 `json:"p50_ms"`
	P90Ms float64 `json:"p90_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

// MetricsSnapshot provides a point-in-time calculation of overall load test metrics.
type MetricsSnapshot struct {
	TestID             string           `json:"test_id"`
	Status             TestStatus       `json:"status"`
	ElapsedSeconds     float64          `json:"elapsed_seconds"`
	ActiveVUs          int              `json:"active_vus"`
	TotalRequests      int64            `json:"total_requests"`
	SuccessfulRequests int64            `json:"successful_requests"`
	FailedRequests     int64            `json:"failed_requests"`
	ErrorRate          float64          `json:"error_rate"`
	CurrentRPS         float64          `json:"current_rps"`
	Latency            LatencyStats     `json:"latency"`
	StatusCodes        map[int]int64    `json:"status_codes"`
	Errors             map[string]int64 `json:"errors,omitempty"`
}
