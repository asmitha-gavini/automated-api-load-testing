package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"backend/internal/database"
	"backend/internal/metrics"
	"backend/internal/websocket"

	"github.com/google/uuid"
)

// Engine orchestrates concurrent load generation, lifecycle management, and metric collection.
type Engine struct {
	mu        sync.RWMutex
	status    TestStatus
	config    *TestConfig
	cancel    context.CancelFunc
	collector *MetricsCollector
	activeVUs int32
	db        *database.DB
	hub       *websocket.Hub
	prom      *metrics.PrometheusMetrics
	doneChan  chan struct{}
}

// NewEngine creates a new load test engine instance.
func NewEngine(db *database.DB, hub *websocket.Hub, prom *metrics.PrometheusMetrics) *Engine {
	return &Engine{
		status:    StatusIdle,
		collector: NewMetricsCollector(),
		db:        db,
		hub:       hub,
		prom:      prom,
	}
}

// StartTest validates and initiates a new concurrent load test.
func (e *Engine) StartTest(cfg TestConfig) (*TestConfig, error) {
	e.mu.Lock()
	if e.status == StatusRunning {
		activeID := ""
		if e.config != nil {
			activeID = e.config.ID
		}
		e.mu.Unlock()
		return nil, fmt.Errorf("a test is already running (test ID: %s)", activeID)
	}

	// 1. Validate configuration
	if err := ValidateConfig(&cfg); err != nil {
		e.mu.Unlock()
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 2. Assign unique ID and timestamps
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.Name == "" {
		cfg.Name = fmt.Sprintf("LoadTest-%s", cfg.ID[:8])
	}
	cfg.CreatedAt = time.Now().UTC()

	// 3. Reset collector & initialize execution context
	e.collector.Reset()
	e.status = StatusRunning
	e.config = &cfg
	atomic.StoreInt32(&e.activeVUs, 0)

	baseCtx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	doneChan := make(chan struct{})
	e.doneChan = doneChan
	e.mu.Unlock()

	// 4. Launch asynchronous execution orchestrator
	go e.runOrchestrator(baseCtx, cfg, doneChan)

	return &cfg, nil
}

// StopTest gracefully terminates an in-flight load test.
func (e *Engine) StopTest() error {
	e.mu.Lock()
	if e.status != StatusRunning || e.cancel == nil {
		e.mu.Unlock()
		return errors.New("no active test is currently running")
	}

	log.Printf("[Engine] Cancellation requested for test %s", e.config.ID)
	e.cancel()
	e.status = StatusStopped
	doneChan := e.doneChan
	e.mu.Unlock()

	// Wait briefly for workers to finish draining
	if doneChan != nil {
		select {
		case <-doneChan:
		case <-time.After(3 * time.Second):
		}
	}

	return nil
}

// GetStatus returns the current lifecycle status, active configuration, and count of active VUs.
func (e *Engine) GetStatus() (TestStatus, *TestConfig, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var cfgCopy *TestConfig
	if e.config != nil {
		cpy := *e.config
		cfgCopy = &cpy
	}

	return e.status, cfgCopy, int(atomic.LoadInt32(&e.activeVUs))
}

// GetMetrics returns a fresh point-in-time calculation of performance metrics.
func (e *Engine) GetMetrics() MetricsSnapshot {
	e.mu.RLock()
	testID := ""
	if e.config != nil {
		testID = e.config.ID
	}
	currentStatus := e.status
	e.mu.RUnlock()

	active := int(atomic.LoadInt32(&e.activeVUs))
	return e.collector.Snapshot(testID, currentStatus, active)
}

// runOrchestrator manages the worker pool lifecycle, ramp-up schedule, and persistence.
func (e *Engine) runOrchestrator(parentCtx context.Context, cfg TestConfig, doneChan chan struct{}) {
	defer close(doneChan)

	var runCtx context.Context
	var runCancel context.CancelFunc

	if cfg.DurationSeconds > 0 {
		runCtx, runCancel = context.WithTimeout(parentCtx, time.Duration(cfg.DurationSeconds)*time.Second)
	} else {
		runCtx, runCancel = context.WithCancel(parentCtx)
	}
	defer runCancel()

	startedAt := time.Now().UTC()
	client := createHTTPClient(cfg.TimeoutMs)
	targetURL := buildTargetURL(cfg.TargetURL, cfg.QueryParams)

	var reqCounter int64
	var wg sync.WaitGroup

	if e.prom != nil {
		e.prom.SetRunningTests(1)
	}

	// Periodic ticker to recalculate instant throughput (RPS) and stream to WebSocket & Prometheus
	ticker := time.NewTicker(500 * time.Millisecond)
	tickerStop := make(chan struct{})
	go func() {
		for {
			select {
			case <-tickerStop:
				ticker.Stop()
				return
			case <-ticker.C:
				e.collector.Tick()
				currentVUs := int(atomic.LoadInt32(&e.activeVUs))
				snapshot := e.collector.Snapshot(cfg.ID, StatusRunning, currentVUs)

				// Update Prometheus live gauges
				if e.prom != nil {
					e.prom.UpdateLiveGauges(snapshot.CurrentRPS, snapshot.ActiveVUs)
				}

				// Broadcast real-time snapshot over WebSocket
				if e.hub != nil {
					msg := websocket.WSMetricMessage{
						TestID:             cfg.ID,
						Status:             string(snapshot.Status),
						Timestamp:          time.Now().UTC().Format(time.RFC3339),
						TotalRequests:      snapshot.TotalRequests,
						SuccessfulRequests: snapshot.SuccessfulRequests,
						FailedRequests:     snapshot.FailedRequests,
						ErrorRate:          snapshot.ErrorRate,
						RequestsPerSecond:  snapshot.CurrentRPS,
						AverageLatencyMs:   snapshot.Latency.AvgMs,
						P50LatencyMs:       snapshot.Latency.P50Ms,
						P95LatencyMs:       snapshot.Latency.P95Ms,
						P99LatencyMs:       snapshot.Latency.P99Ms,
						ActiveUsers:        snapshot.ActiveVUs,
						StatusCodes:        snapshot.StatusCodes,
					}
					_ = e.hub.BroadcastToTest(cfg.ID, msg)
				}
			}
		}
	}()

	log.Printf("[Engine] Starting test '%s' (ID: %s) against %s [VUs: %d, Duration: %ds, RampUp: %ds]",
		cfg.Name, cfg.ID, targetURL, cfg.VirtualUsers, cfg.DurationSeconds, cfg.RampUpSeconds)

	// Ramp-up delay calculation
	var delayBetweenWorkers time.Duration
	if cfg.RampUpSeconds > 0 && cfg.VirtualUsers > 1 {
		delayBetweenWorkers = time.Duration(cfg.RampUpSeconds) * time.Second / time.Duration(cfg.VirtualUsers)
	}

	// Spawn Virtual Users according to ramp-up schedule
	for i := 0; i < cfg.VirtualUsers; i++ {
		select {
		case <-runCtx.Done():
			break
		default:
		}

		if delayBetweenWorkers > 0 && i > 0 {
			select {
			case <-runCtx.Done():
				break
			case <-time.After(delayBetweenWorkers):
			}
		}

		wg.Add(1)
		atomic.AddInt32(&e.activeVUs, 1)

		go func(workerID int) {
			defer func() {
				atomic.AddInt32(&e.activeVUs, -1)
				wg.Done()
			}()
			runWorker(runCtx, workerID, &cfg, client, targetURL, e.collector, &reqCounter, e.prom)
		}(i + 1)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(tickerStop)
	completedAt := time.Now().UTC()

	// Determine terminal state
	e.mu.Lock()
	finalStatus := StatusCompleted
	if parentCtx.Err() != nil {
		finalStatus = StatusStopped
	}
	e.status = finalStatus
	e.mu.Unlock()

	log.Printf("[Engine] Test '%s' (ID: %s) concluded with status: %s", cfg.Name, cfg.ID, finalStatus)

	finalSnapshot := e.collector.Snapshot(cfg.ID, finalStatus, 0)

	// 1. Update Prometheus final state
	if e.prom != nil {
		e.prom.UpdateLiveGauges(finalSnapshot.CurrentRPS, 0)
		e.prom.SetRunningTests(0)
	}

	// 2. Broadcast terminal test message over WebSocket
	if e.hub != nil {
		terminalMsg := websocket.WSMetricMessage{
			TestID:             cfg.ID,
			Status:             string(finalStatus),
			Timestamp:          completedAt.Format(time.RFC3339),
			TotalRequests:      finalSnapshot.TotalRequests,
			SuccessfulRequests: finalSnapshot.SuccessfulRequests,
			FailedRequests:     finalSnapshot.FailedRequests,
			ErrorRate:          finalSnapshot.ErrorRate,
			RequestsPerSecond:  finalSnapshot.CurrentRPS,
			AverageLatencyMs:   finalSnapshot.Latency.AvgMs,
			P50LatencyMs:       finalSnapshot.Latency.P50Ms,
			P95LatencyMs:       finalSnapshot.Latency.P95Ms,
			P99LatencyMs:       finalSnapshot.Latency.P99Ms,
			ActiveUsers:        0,
			StatusCodes:        finalSnapshot.StatusCodes,
		}
		_ = e.hub.BroadcastToTest(cfg.ID, terminalMsg)
	}

	// 3. Persist final metrics to SQLite if DB is connected
	if e.db != nil {
		testRun := database.TestRun{
			ID:                 cfg.ID,
			Name:               cfg.Name,
			TargetURL:          cfg.TargetURL,
			Method:             cfg.Method,
			VirtualUsers:       cfg.VirtualUsers,
			DurationSeconds:    cfg.DurationSeconds,
			TotalRequests:      finalSnapshot.TotalRequests,
			SuccessfulRequests: finalSnapshot.SuccessfulRequests,
			FailedRequests:     finalSnapshot.FailedRequests,
			ErrorRate:          finalSnapshot.ErrorRate,
			AvgLatencyMs:       finalSnapshot.Latency.AvgMs,
			MinLatencyMs:       finalSnapshot.Latency.MinMs,
			MaxLatencyMs:       finalSnapshot.Latency.MaxMs,
			P50LatencyMs:       finalSnapshot.Latency.P50Ms,
			P90LatencyMs:       finalSnapshot.Latency.P90Ms,
			P95LatencyMs:       finalSnapshot.Latency.P95Ms,
			P99LatencyMs:       finalSnapshot.Latency.P99Ms,
			Status:             string(finalStatus),
			StartedAt:          startedAt,
			CompletedAt:        &completedAt,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := e.db.SaveTestRun(ctx, &testRun); err != nil {
			log.Printf("[Engine] Warning: failed to persist completed test run to database: %v", err)
		} else {
			log.Printf("[Engine] Test run summary successfully stored in database")
		}
	}
}
