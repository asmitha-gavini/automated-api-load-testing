package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the standard sql.DB pool with helper methods.
type DB struct {
	*sql.DB
}

// InitDB initializes a SQLite database connection with WAL mode and runs initial schema migrations.
func InitDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// For SQLite, 1 open writer prevents Windows file lock collisions.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Set pragmas for performance and concurrency resilience
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to apply pragma '%s': %w", pragma, err)
		}
	}

	wrapped := &DB{DB: db}
	if err := wrapped.runMigrations(); err != nil {
		wrapped.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return wrapped, nil
}

// runMigrations applies table definitions and indexes if they do not exist.
func (db *DB) runMigrations() error {
	schema := `
	CREATE TABLE IF NOT EXISTS test_runs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		target_url TEXT NOT NULL,
		method TEXT NOT NULL,
		virtual_users INTEGER NOT NULL,
		duration_seconds INTEGER NOT NULL,
		total_requests INTEGER NOT NULL DEFAULT 0,
		successful_requests INTEGER NOT NULL DEFAULT 0,
		failed_requests INTEGER NOT NULL DEFAULT 0,
		error_rate REAL NOT NULL DEFAULT 0.0,
		avg_latency_ms REAL NOT NULL DEFAULT 0.0,
		min_latency_ms REAL NOT NULL DEFAULT 0.0,
		max_latency_ms REAL NOT NULL DEFAULT 0.0,
		p50_latency_ms REAL NOT NULL DEFAULT 0.0,
		p90_latency_ms REAL NOT NULL DEFAULT 0.0,
		p95_latency_ms REAL NOT NULL DEFAULT 0.0,
		p99_latency_ms REAL NOT NULL DEFAULT 0.0,
		status TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_test_runs_started_at ON test_runs(started_at DESC);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error executing schema migration: %w", err)
	}

	return nil
}

// SaveTestRun inserts or updates a test run summary.
func (db *DB) SaveTestRun(ctx context.Context, run *TestRun) error {
	query := `
	INSERT INTO test_runs (
		id, name, target_url, method, virtual_users, duration_seconds,
		total_requests, successful_requests, failed_requests, error_rate,
		avg_latency_ms, min_latency_ms, max_latency_ms, p50_latency_ms,
		p90_latency_ms, p95_latency_ms, p99_latency_ms, status,
		started_at, completed_at
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?
	)
	ON CONFLICT(id) DO UPDATE SET
		total_requests = excluded.total_requests,
		successful_requests = excluded.successful_requests,
		failed_requests = excluded.failed_requests,
		error_rate = excluded.error_rate,
		avg_latency_ms = excluded.avg_latency_ms,
		min_latency_ms = excluded.min_latency_ms,
		max_latency_ms = excluded.max_latency_ms,
		p50_latency_ms = excluded.p50_latency_ms,
		p90_latency_ms = excluded.p90_latency_ms,
		p95_latency_ms = excluded.p95_latency_ms,
		p99_latency_ms = excluded.p99_latency_ms,
		status = excluded.status,
		completed_at = excluded.completed_at;
	`

	_, err := db.ExecContext(ctx, query,
		run.ID, run.Name, run.TargetURL, run.Method, run.VirtualUsers, run.DurationSeconds,
		run.TotalRequests, run.SuccessfulRequests, run.FailedRequests, run.ErrorRate,
		run.AvgLatencyMs, run.MinLatencyMs, run.MaxLatencyMs, run.P50LatencyMs,
		run.P90LatencyMs, run.P95LatencyMs, run.P99LatencyMs, run.Status,
		run.StartedAt, run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save test run: %w", err)
	}
	return nil
}

// GetTestRun retrieves a specific test run by ID.
func (db *DB) GetTestRun(ctx context.Context, id string) (*TestRun, error) {
	query := `
	SELECT id, name, target_url, method, virtual_users, duration_seconds,
	       total_requests, successful_requests, failed_requests, error_rate,
	       avg_latency_ms, min_latency_ms, max_latency_ms, p50_latency_ms,
	       p90_latency_ms, p95_latency_ms, p99_latency_ms, status,
	       started_at, completed_at
	FROM test_runs WHERE id = ?;
	`

	var run TestRun
	err := db.QueryRowContext(ctx, query, id).Scan(
		&run.ID, &run.Name, &run.TargetURL, &run.Method, &run.VirtualUsers, &run.DurationSeconds,
		&run.TotalRequests, &run.SuccessfulRequests, &run.FailedRequests, &run.ErrorRate,
		&run.AvgLatencyMs, &run.MinLatencyMs, &run.MaxLatencyMs, &run.P50LatencyMs,
		&run.P90LatencyMs, &run.P95LatencyMs, &run.P99LatencyMs, &run.Status,
		&run.StartedAt, &run.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get test run: %w", err)
	}
	return &run, nil
}

// ListTestRuns retrieves the most recent test runs.
func (db *DB) ListTestRuns(ctx context.Context, limit int) ([]TestRun, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, name, target_url, method, virtual_users, duration_seconds,
	       total_requests, successful_requests, failed_requests, error_rate,
	       avg_latency_ms, min_latency_ms, max_latency_ms, p50_latency_ms,
	       p90_latency_ms, p95_latency_ms, p99_latency_ms, status,
	       started_at, completed_at
	FROM test_runs ORDER BY started_at DESC LIMIT ?;
	`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list test runs: %w", err)
	}
	defer rows.Close()

	runs := make([]TestRun, 0)
	for rows.Next() {
		var run TestRun
		if err := rows.Scan(
			&run.ID, &run.Name, &run.TargetURL, &run.Method, &run.VirtualUsers, &run.DurationSeconds,
			&run.TotalRequests, &run.SuccessfulRequests, &run.FailedRequests, &run.ErrorRate,
			&run.AvgLatencyMs, &run.MinLatencyMs, &run.MaxLatencyMs, &run.P50LatencyMs,
			&run.P90LatencyMs, &run.P95LatencyMs, &run.P99LatencyMs, &run.Status,
			&run.StartedAt, &run.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan test run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// Ping checks if the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// ListTestRunsFiltered retrieves test runs matching search criteria, status filter, and sorted with pagination.
// Also returns total count of matching records.
func (db *DB) ListTestRunsFiltered(ctx context.Context, filter TestRunFilter) ([]TestRun, int64, error) {
	whereClauses := make([]string, 0)
	args := make([]interface{}, 0)

	if strings.TrimSpace(filter.Search) != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		whereClauses = append(whereClauses, "(name LIKE ? OR target_url LIKE ?)")
		args = append(args, s, s)
	}

	if strings.TrimSpace(filter.Status) != "" && strings.ToLower(filter.Status) != "all" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(filter.Status)))
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Get total count
	countQuery := "SELECT COUNT(*) FROM test_runs" + whereSQL
	var total int64
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count test runs: %w", err)
	}

	// 2. Validate sort column against whitelist
	allowedSorts := map[string]string{
		"started_at":       "started_at",
		"duration_seconds": "duration_seconds",
		"total_requests":   "total_requests",
		"avg_latency_ms":   "avg_latency_ms",
		"error_rate":       "error_rate",
		"virtual_users":    "virtual_users",
	}
	sortCol := "started_at"
	if col, ok := allowedSorts[strings.ToLower(filter.SortBy)]; ok {
		sortCol = col
	}

	orderDir := "DESC"
	if strings.ToUpper(filter.Order) == "ASC" {
		orderDir = "ASC"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
	SELECT id, name, target_url, method, virtual_users, duration_seconds,
	       total_requests, successful_requests, failed_requests, error_rate,
	       avg_latency_ms, min_latency_ms, max_latency_ms, p50_latency_ms,
	       p90_latency_ms, p95_latency_ms, p99_latency_ms, status,
	       started_at, completed_at
	FROM test_runs%s
	ORDER BY %s %s
	LIMIT ? OFFSET ?;
	`, whereSQL, sortCol, orderDir)

	queryArgs := append(args, limit, offset)
	rows, err := db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list filtered test runs: %w", err)
	}
	defer rows.Close()

	runs := make([]TestRun, 0)
	for rows.Next() {
		var run TestRun
		if err := rows.Scan(
			&run.ID, &run.Name, &run.TargetURL, &run.Method, &run.VirtualUsers, &run.DurationSeconds,
			&run.TotalRequests, &run.SuccessfulRequests, &run.FailedRequests, &run.ErrorRate,
			&run.AvgLatencyMs, &run.MinLatencyMs, &run.MaxLatencyMs, &run.P50LatencyMs,
			&run.P90LatencyMs, &run.P95LatencyMs, &run.P99LatencyMs, &run.Status,
			&run.StartedAt, &run.CompletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan test run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, total, rows.Err()
}

// DeleteTestRun deletes a single test run by its ID. Returns true if a record was deleted.
func (db *DB) DeleteTestRun(ctx context.Context, id string) (bool, error) {
	res, err := db.ExecContext(ctx, "DELETE FROM test_runs WHERE id = ?;", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete test run: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// ClearTestRuns removes all test records from the database.
func (db *DB) ClearTestRuns(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "DELETE FROM test_runs;")
	if err != nil {
		return fmt.Errorf("failed to clear test runs: %w", err)
	}
	return nil
}

// GetPerformanceSummary aggregates performance statistics across all test runs.
func (db *DB) GetPerformanceSummary(ctx context.Context) (PerformanceSummary, error) {
	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'stopped' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(total_requests), 0),
		COALESCE(SUM(successful_requests), 0),
		COALESCE(SUM(failed_requests), 0),
		COALESCE(AVG(avg_latency_ms), 0.0),
		COALESCE(MIN(CASE WHEN min_latency_ms > 0 THEN min_latency_ms ELSE NULL END), 0.0),
		COALESCE(MAX(max_latency_ms), 0.0)
	FROM test_runs;
	`

	var s PerformanceSummary
	err := db.QueryRowContext(ctx, query).Scan(
		&s.TotalTests,
		&s.CompletedTests,
		&s.StoppedTests,
		&s.FailedTests,
		&s.TotalRequests,
		&s.SuccessfulRequests,
		&s.FailedRequests,
		&s.AvgLatencyMs,
		&s.MinLatencyMs,
		&s.MaxLatencyMs,
	)
	if err != nil {
		return s, fmt.Errorf("failed to calculate performance summary: %w", err)
	}

	if s.TotalRequests > 0 {
		s.OverallErrorRate = float64(s.FailedRequests) / float64(s.TotalRequests) * 100.0
	}

	// Round latency numbers to 2 decimal places
	s.AvgLatencyMs = float64(int64(s.AvgLatencyMs*100+0.5)) / 100.0
	s.MinLatencyMs = float64(int64(s.MinLatencyMs*100+0.5)) / 100.0
	s.MaxLatencyMs = float64(int64(s.MaxLatencyMs*100+0.5)) / 100.0
	s.OverallErrorRate = float64(int64(s.OverallErrorRate*100+0.5)) / 100.0

	return s, nil
}

