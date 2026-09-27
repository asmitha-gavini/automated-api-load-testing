package database

import "time"

// TestRun represents a completed or ongoing load test execution record.
type TestRun struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	TargetURL          string     `json:"target_url"`
	Method             string     `json:"method"`
	VirtualUsers       int        `json:"virtual_users"`
	DurationSeconds    int        `json:"duration_seconds"`
	TotalRequests      int64      `json:"total_requests"`
	SuccessfulRequests int64      `json:"successful_requests"`
	FailedRequests     int64      `json:"failed_requests"`
	ErrorRate          float64    `json:"error_rate"`
	AvgLatencyMs       float64    `json:"avg_latency_ms"`
	MinLatencyMs       float64    `json:"min_latency_ms"`
	MaxLatencyMs       float64    `json:"max_latency_ms"`
	P50LatencyMs       float64    `json:"p50_latency_ms"`
	P90LatencyMs       float64    `json:"p90_latency_ms"`
	P95LatencyMs       float64    `json:"p95_latency_ms"`
	P99LatencyMs       float64    `json:"p99_latency_ms"`
	Status             string     `json:"status"` // "running", "completed", "aborted", "failed"
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

// HealthStatus represents the database and service health check response.
type HealthStatus struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
}

// TestRunFilter defines search, status filter, sorting, and pagination parameters.
type TestRunFilter struct {
	Search string `json:"search"`
	Status string `json:"status"` // "completed", "stopped", "failed", or "" (all)
	SortBy string `json:"sort_by"` // "started_at", "total_requests", "avg_latency_ms", "error_rate", "duration_seconds"
	Order  string `json:"order"`   // "ASC", "DESC"
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// PerformanceSummary encapsulates aggregated metrics across stored test runs.
type PerformanceSummary struct {
	TotalTests         int64   `json:"total_tests"`
	CompletedTests     int64   `json:"completed_tests"`
	StoppedTests       int64   `json:"stopped_tests"`
	FailedTests        int64   `json:"failed_tests"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	OverallErrorRate   float64 `json:"overall_error_rate"`
	AvgLatencyMs       float64 `json:"avg_latency_ms"`
	MinLatencyMs       float64 `json:"min_latency_ms"`
	MaxLatencyMs       float64 `json:"max_latency_ms"`
}

// TestHistoryResponse encapsulates paginated tests with overall summary metrics.
type TestHistoryResponse struct {
	Tests   []TestRun          `json:"tests"`
	Total   int64              `json:"total"`
	Summary PerformanceSummary `json:"summary"`
}
