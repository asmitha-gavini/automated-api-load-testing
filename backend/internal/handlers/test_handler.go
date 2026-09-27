package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/database"
	"backend/internal/engine"

	"github.com/gin-gonic/gin"
)

// TestHandler handles REST endpoints for load test execution and metrics inspection.
type TestHandler struct {
	engine *engine.Engine
	db     *database.DB
}

// NewTestHandler creates a new TestHandler instance.
func NewTestHandler(eng *engine.Engine, db *database.DB) *TestHandler {
	return &TestHandler{
		engine: eng,
		db:     db,
	}
}

// StartTest handles POST /api/v1/tests to validate and begin a load test.
func (h *TestHandler) StartTest(c *gin.Context) {
	var cfg engine.TestConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	startedConfig, err := h.engine.StartTest(cfg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Load test initiated successfully",
		"test":    startedConfig,
	})
}

// GetStatus handles GET /api/v1/tests/status to retrieve lifecycle state.
func (h *TestHandler) GetStatus(c *gin.Context) {
	status, cfg, activeVUs := h.engine.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"active_vus": activeVUs,
		"config":     cfg,
	})
}

// GetMetrics handles GET /api/v1/tests/metrics to retrieve real-time performance statistics.
func (h *TestHandler) GetMetrics(c *gin.Context) {
	metrics := h.engine.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// StopTest handles POST /api/v1/tests/stop to cancel the active load test.
func (h *TestHandler) StopTest(c *gin.Context) {
	if err := h.engine.StopTest(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Load test stopped successfully",
	})
}

// ListTests handles GET /api/v1/tests to list historical test executions with search, filtering, and sorting.
func (h *TestHandler) ListTests(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	search := c.DefaultQuery("q", c.Query("search"))
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "started_at")
	order := c.DefaultQuery("order", "DESC")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := database.TestRunFilter{
		Search: search,
		Status: status,
		SortBy: sortBy,
		Order:  order,
		Limit:  limit,
		Offset: offset,
	}

	tests, total, err := h.db.ListTestRunsFiltered(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query test history", "details": err.Error()})
		return
	}

	summary, err := h.db.GetPerformanceSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate performance summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, database.TestHistoryResponse{
		Tests:   tests,
		Total:   total,
		Summary: summary,
	})
}

// GetPerformanceSummary handles GET /api/v1/tests/summary to retrieve aggregated benchmark stats.
func (h *TestHandler) GetPerformanceSummary(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	summary, err := h.db.GetPerformanceSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTestByID handles GET /api/v1/tests/:id to query historical test results from the database.
func (h *TestHandler) GetTestByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test ID is required"})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	run, err := h.db.GetTestRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test run not found"})
		return
	}

	c.JSON(http.StatusOK, run)
}

// DeleteTest handles DELETE /api/v1/tests/:id to remove a specific test run record.
func (h *TestHandler) DeleteTest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test ID is required"})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	deleted, err := h.db.DeleteTestRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete test run", "details": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test run deleted successfully",
		"id":      id,
	})
}

// ClearAllTests handles DELETE /api/v1/tests to clear the entire test history.
func (h *TestHandler) ClearAllTests(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	if err := h.db.ClearTestRuns(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear test runs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All test history cleared successfully",
	})
}

// ExportTestReport handles GET /api/v1/tests/:id/export to download test reports in json, csv, or markdown.
func (h *TestHandler) ExportTestReport(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test ID is required"})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database is not connected"})
		return
	}

	run, err := h.db.GetTestRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test run not found"})
		return
	}

	format := strings.ToLower(c.DefaultQuery("format", "json"))

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"report-%s.csv\"", run.ID))
		csvData := fmt.Sprintf(
			"Metric,Value\n"+
				"Test ID,%s\n"+
				"Test Name,%s\n"+
				"Target URL,%s\n"+
				"HTTP Method,%s\n"+
				"Virtual Users,%d\n"+
				"Duration (s),%d\n"+
				"Status,%s\n"+
				"Total Requests,%d\n"+
				"Successful Requests,%d\n"+
				"Failed Requests,%d\n"+
				"Error Rate (%%),%.2f\n"+
				"Average Latency (ms),%.2f\n"+
				"Min Latency (ms),%.2f\n"+
				"Max Latency (ms),%.2f\n"+
				"P50 Latency (ms),%.2f\n"+
				"P90 Latency (ms),%.2f\n"+
				"P95 Latency (ms),%.2f\n"+
				"P99 Latency (ms),%.2f\n"+
				"Started At,%s\n",
			run.ID, run.Name, run.TargetURL, run.Method, run.VirtualUsers, run.DurationSeconds,
			run.Status, run.TotalRequests, run.SuccessfulRequests, run.FailedRequests, run.ErrorRate,
			run.AvgLatencyMs, run.MinLatencyMs, run.MaxLatencyMs, run.P50LatencyMs, run.P90LatencyMs,
			run.P95LatencyMs, run.P99LatencyMs, run.StartedAt.Format(time.RFC3339),
		)
		c.String(http.StatusOK, csvData)

	case "markdown", "md":
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"report-%s.md\"", run.ID))

		successPct := 100.0
		rps := 0.0
		if run.TotalRequests > 0 {
			successPct = (float64(run.SuccessfulRequests) / float64(run.TotalRequests)) * 100.0
		}
		if run.DurationSeconds > 0 {
			rps = float64(run.TotalRequests) / float64(run.DurationSeconds)
		}

		md := fmt.Sprintf(
			"# API Load Test Report: %s\n\n"+
				"- **Test ID**: `%s`\n"+
				"- **Target Endpoint**: `%s %s`\n"+
				"- **Execution Time**: %s\n"+
				"- **Concurrency**: %d Virtual Users\n"+
				"- **Duration**: %d seconds\n"+
				"- **Result Status**: **%s**\n\n"+
				"## Executive Performance Summary\n\n"+
				"| Metric | Result |\n"+
				"| :--- | :--- |\n"+
				"| **Total Dispatched Requests** | %d |\n"+
				"| **Successful Requests (2xx/3xx)** | %d (%.1f%%) |\n"+
				"| **Failed Requests** | %d |\n"+
				"| **Error Rate** | %.2f%% |\n"+
				"| **Throughput (RPS)** | %.1f req/s |\n"+
				"| **Average Round-Trip Latency** | %.2f ms |\n"+
				"| **Min Latency** | %.2f ms |\n"+
				"| **Max Latency** | %.2f ms |\n\n"+
				"## Latency Percentiles (SLA Breakdown)\n\n"+
				"| Percentile | Response Time |\n"+
				"| :--- | :--- |\n"+
				"| **P50 (Median)** | %.2f ms |\n"+
				"| **P90** | %.2f ms |\n"+
				"| **P95** | %.2f ms |\n"+
				"| **P99 (Tail Outlier)** | %.2f ms |\n",
			run.Name, run.ID, run.Method, run.TargetURL, run.StartedAt.Format("2006-01-02 15:04:05 UTC"),
			run.VirtualUsers, run.DurationSeconds, strings.ToUpper(run.Status),
			run.TotalRequests, run.SuccessfulRequests, successPct,
			run.FailedRequests, run.ErrorRate, rps,
			run.AvgLatencyMs, run.MinLatencyMs, run.MaxLatencyMs,
			run.P50LatencyMs, run.P90LatencyMs, run.P95LatencyMs, run.P99LatencyMs,
		)
		c.String(http.StatusOK, md)

	default: // JSON
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"report-%s.json\"", run.ID))
		c.JSON(http.StatusOK, run)
	}
}
