package engine

import (
	"sync"
	"time"
)

// MetricsCollector accumulates request outcomes concurrently and generates metric snapshots.
type MetricsCollector struct {
	mu                 sync.RWMutex
	startTime          time.Time
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	statusCodes        map[int]int64
	errorCounts        map[string]int64
	latencies          []float64

	// Sliding window for instant throughput (RPS)
	lastTickTime  time.Time
	lastTickTotal int64
	currentRPS    float64
}

// NewMetricsCollector initializes a thread-safe metrics collector.
func NewMetricsCollector() *MetricsCollector {
	now := time.Now()
	return &MetricsCollector{
		startTime:    now,
		lastTickTime: now,
		statusCodes:  make(map[int]int64),
		errorCounts:  make(map[string]int64),
		latencies:    make([]float64, 0, 1024),
	}
}

// Reset clears all recorded metrics and re-anchors the start timestamp.
func (c *MetricsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.startTime = now
	c.lastTickTime = now
	c.totalRequests = 0
	c.successfulRequests = 0
	c.failedRequests = 0
	c.currentRPS = 0.0
	c.lastTickTotal = 0
	c.statusCodes = make(map[int]int64)
	c.errorCounts = make(map[string]int64)
	c.latencies = make([]float64, 0, 1024)
}

// Record thread-safely logs the result of an executed HTTP request.
func (c *MetricsCollector) Record(res RequestResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests++
	if res.Success {
		c.successfulRequests++
	} else {
		c.failedRequests++
	}

	if res.StatusCode > 0 {
		c.statusCodes[res.StatusCode]++
	}

	if res.Error != "" {
		c.errorCounts[res.Error]++
	}

	if res.DurationMs >= 0 {
		c.latencies = append(c.latencies, res.DurationMs)
	}
}

// Tick updates the instantaneous requests-per-second rate based on requests completed since the last tick.
func (c *MetricsCollector) Tick() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastTickTime).Seconds()
	if elapsed >= 0.5 {
		deltaReqs := c.totalRequests - c.lastTickTotal
		c.currentRPS = roundToTwoDecimals(float64(deltaReqs) / elapsed)
		c.lastTickTime = now
		c.lastTickTotal = c.totalRequests
	}
}

// Snapshot extracts a point-in-time calculation of all performance metrics.
func (c *MetricsCollector) Snapshot(testID string, status TestStatus, activeVUs int) MetricsSnapshot {
	c.mu.RLock()
	now := time.Now()
	elapsed := now.Sub(c.startTime).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}

	total := c.totalRequests
	success := c.successfulRequests
	failed := c.failedRequests
	rps := c.currentRPS

	// If currentRPS is 0 (e.g. at the very start or end), calculate average RPS
	if rps == 0 && total > 0 {
		rps = roundToTwoDecimals(float64(total) / elapsed)
	}

	// Copy status codes map
	statusCodesCopy := make(map[int]int64, len(c.statusCodes))
	for k, v := range c.statusCodes {
		statusCodesCopy[k] = v
	}

	// Copy error counts map
	errorCountsCopy := make(map[string]int64, len(c.errorCounts))
	for k, v := range c.errorCounts {
		errorCountsCopy[k] = v
	}

	// Copy latencies slice for safe out-of-lock percentile computation
	latenciesCopy := make([]float64, len(c.latencies))
	copy(latenciesCopy, c.latencies)
	c.mu.RUnlock()

	// Calculate error rate
	var errorRate float64
	if total > 0 {
		errorRate = roundToTwoDecimals((float64(failed) / float64(total)) * 100.0)
	}

	latencyStats := CalculateLatencyStats(latenciesCopy)

	return MetricsSnapshot{
		TestID:             testID,
		Status:             status,
		ElapsedSeconds:     roundToTwoDecimals(elapsed),
		ActiveVUs:          activeVUs,
		TotalRequests:      total,
		SuccessfulRequests: success,
		FailedRequests:     failed,
		ErrorRate:          errorRate,
		CurrentRPS:         rps,
		Latency:            latencyStats,
		StatusCodes:        statusCodesCopy,
		Errors:             errorCountsCopy,
	}
}
