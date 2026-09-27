package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitDB_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_loadtest_db_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Errorf("expected db.Ping to succeed, got: %v", err)
	}

	// Verify test_runs table was created
	var tableName string
	err = db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='test_runs';").Scan(&tableName)
	if err != nil {
		t.Fatalf("expected test_runs table to exist: %v", err)
	}

	if tableName != "test_runs" {
		t.Errorf("expected table name 'test_runs', got '%s'", tableName)
	}
}

func TestDB_CRUD_And_Filtering(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_db_crud_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Insert 3 test runs
	run1 := &TestRun{
		ID:                 "run-1",
		Name:               "Users Benchmark",
		TargetURL:          "http://localhost:8081/api/users",
		Method:             "GET",
		VirtualUsers:       5,
		DurationSeconds:    10,
		TotalRequests:      1000,
		SuccessfulRequests: 950,
		FailedRequests:     50,
		ErrorRate:          5.0,
		AvgLatencyMs:       20.5,
		MinLatencyMs:       10.0,
		MaxLatencyMs:       150.0,
		P50LatencyMs:       18.0,
		P90LatencyMs:       25.0,
		P95LatencyMs:       35.0,
		P99LatencyMs:       70.0,
		Status:             "completed",
		StartedAt:          now.Add(-20 * time.Minute),
		CompletedAt:        &now,
	}

	run2 := &TestRun{
		ID:                 "run-2",
		Name:               "Orders Heavy Write",
		TargetURL:          "http://localhost:8081/api/orders",
		Method:             "POST",
		VirtualUsers:       10,
		DurationSeconds:    15,
		TotalRequests:      500,
		SuccessfulRequests: 500,
		FailedRequests:     0,
		ErrorRate:          0.0,
		AvgLatencyMs:       55.2,
		MinLatencyMs:       25.0,
		MaxLatencyMs:       210.0,
		P50LatencyMs:       50.0,
		P90LatencyMs:       80.0,
		P95LatencyMs:       110.0,
		P99LatencyMs:       180.0,
		Status:             "completed",
		StartedAt:          now.Add(-10 * time.Minute),
		CompletedAt:        &now,
	}

	run3 := &TestRun{
		ID:                 "run-3",
		Name:               "Products Stress Test",
		TargetURL:          "http://localhost:8081/api/products",
		Method:             "GET",
		VirtualUsers:       20,
		DurationSeconds:    30,
		TotalRequests:      200,
		SuccessfulRequests: 180,
		FailedRequests:     20,
		ErrorRate:          10.0,
		AvgLatencyMs:       85.0,
		MinLatencyMs:       30.0,
		MaxLatencyMs:       300.0,
		P50LatencyMs:       80.0,
		P90LatencyMs:       120.0,
		P95LatencyMs:       160.0,
		P99LatencyMs:       250.0,
		Status:             "stopped",
		StartedAt:          now.Add(-5 * time.Minute),
		CompletedAt:        &now,
	}

	if err := db.SaveTestRun(ctx, run1); err != nil {
		t.Fatalf("failed to save run1: %v", err)
	}
	if err := db.SaveTestRun(ctx, run2); err != nil {
		t.Fatalf("failed to save run2: %v", err)
	}
	if err := db.SaveTestRun(ctx, run3); err != nil {
		t.Fatalf("failed to save run3: %v", err)
	}

	// 1. Get by ID
	fetched, err := db.GetTestRun(ctx, "run-1")
	if err != nil || fetched == nil {
		t.Fatalf("expected to get run-1, got err: %v", err)
	}
	if fetched.Name != "Users Benchmark" || fetched.TotalRequests != 1000 {
		t.Errorf("unexpected run-1 data: %+v", fetched)
	}

	// 2. Filter by search query
	runs, total, err := db.ListTestRunsFiltered(ctx, TestRunFilter{
		Search: "Orders",
	})
	if err != nil {
		t.Fatalf("ListTestRunsFiltered error: %v", err)
	}
	if total != 1 || len(runs) != 1 || runs[0].ID != "run-2" {
		t.Errorf("expected 1 result for search 'Orders', got total=%d, len=%d", total, len(runs))
	}

	// 3. Filter by status
	runs, total, err = db.ListTestRunsFiltered(ctx, TestRunFilter{
		Status: "stopped",
	})
	if err != nil {
		t.Fatalf("status filter error: %v", err)
	}
	if total != 1 || runs[0].ID != "run-3" {
		t.Errorf("expected 1 stopped run, got total=%d", total)
	}

	// 4. Sort by avg_latency_ms ASC
	runs, total, err = db.ListTestRunsFiltered(ctx, TestRunFilter{
		SortBy: "avg_latency_ms",
		Order:  "ASC",
	})
	if err != nil {
		t.Fatalf("sort error: %v", err)
	}
	if total != 3 || runs[0].ID != "run-1" || runs[2].ID != "run-3" {
		t.Errorf("expected sort ASC by latency: run-1, run-2, run-3, got %s, %s, %s",
			runs[0].ID, runs[1].ID, runs[2].ID)
	}

	// 5. Test Performance Summary
	summary, err := db.GetPerformanceSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceSummary error: %v", err)
	}
	if summary.TotalTests != 3 {
		t.Errorf("expected 3 total tests, got %d", summary.TotalTests)
	}
	if summary.CompletedTests != 2 || summary.StoppedTests != 1 {
		t.Errorf("unexpected test statuses count: %+v", summary)
	}
	if summary.TotalRequests != 1700 {
		t.Errorf("expected 1700 total requests, got %d", summary.TotalRequests)
	}
	if summary.SuccessfulRequests != 1630 || summary.FailedRequests != 70 {
		t.Errorf("unexpected success/fail counts: %+v", summary)
	}
	if summary.MinLatencyMs != 10.0 || summary.MaxLatencyMs != 300.0 {
		t.Errorf("unexpected min/max latency: min=%f, max=%f", summary.MinLatencyMs, summary.MaxLatencyMs)
	}

	// 6. Delete single run
	deleted, err := db.DeleteTestRun(ctx, "run-2")
	if err != nil || !deleted {
		t.Fatalf("expected successful delete of run-2, got deleted=%v, err=%v", deleted, err)
	}
	fetchedDeleted, err := db.GetTestRun(ctx, "run-2")
	if err != nil || fetchedDeleted != nil {
		t.Errorf("expected run-2 to be nil after deletion, got: %+v", fetchedDeleted)
	}

	// 7. Clear all runs
	if err := db.ClearTestRuns(ctx); err != nil {
		t.Fatalf("ClearTestRuns error: %v", err)
	}
	summaryAfterClear, err := db.GetPerformanceSummary(ctx)
	if err != nil {
		t.Fatalf("GetPerformanceSummary after clear error: %v", err)
	}
	if summaryAfterClear.TotalTests != 0 {
		t.Errorf("expected 0 total tests after clear, got %d", summaryAfterClear.TotalTests)
	}
}
