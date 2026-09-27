package engine

import (
	"sync"
	"testing"
	"time"
)

func TestMetricsCollector_SingleThread(t *testing.T) {
	c := NewMetricsCollector()

	c.Record(RequestResult{
		DurationMs: 15.0,
		StatusCode: 200,
		Success:    true,
	})
	c.Record(RequestResult{
		DurationMs: 25.0,
		StatusCode: 200,
		Success:    true,
	})
	c.Record(RequestResult{
		DurationMs: 50.0,
		StatusCode: 500,
		Success:    false,
		Error:      "Internal Server Error",
	})

	snap := c.Snapshot("test-1", StatusRunning, 2)
	if snap.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", snap.TotalRequests)
	}
	if snap.SuccessfulRequests != 2 {
		t.Errorf("expected 2 successful requests, got %d", snap.SuccessfulRequests)
	}
	if snap.FailedRequests != 1 {
		t.Errorf("expected 1 failed request, got %d", snap.FailedRequests)
	}
	if snap.StatusCodes[200] != 2 {
		t.Errorf("expected 2 status 200s, got %d", snap.StatusCodes[200])
	}
	if snap.StatusCodes[500] != 1 {
		t.Errorf("expected 1 status 500, got %d", snap.StatusCodes[500])
	}
	if snap.ErrorRate != 33.33 {
		t.Errorf("expected error rate ~33.33, got %f", snap.ErrorRate)
	}
}

func TestMetricsCollector_ConcurrentSafety(t *testing.T) {
	c := NewMetricsCollector()
	var wg sync.WaitGroup

	numGoroutines := 20
	requestsPerGoroutine := 250
	expectedTotal := int64(numGoroutines * requestsPerGoroutine)

	// Launch concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				isSuccess := (j % 5) != 0 // 20% failure
				code := 200
				if !isSuccess {
					code = 500
				}
				c.Record(RequestResult{
					StartTime:  time.Now(),
					EndTime:    time.Now(),
					DurationMs: float64(10 + (j % 50)),
					StatusCode: code,
					Success:    isSuccess,
					Error:      "",
				})
			}
		}(i)
	}

	// Concurrent reader running concurrently with writers
	stopReader := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopReader:
				return
			default:
				_ = c.Snapshot("test-race", StatusRunning, 5)
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	close(stopReader)

	finalSnap := c.Snapshot("test-race", StatusCompleted, 0)
	if finalSnap.TotalRequests != expectedTotal {
		t.Fatalf("expected total %d, got %d", expectedTotal, finalSnap.TotalRequests)
	}

	expectedFailures := int64(numGoroutines * (requestsPerGoroutine / 5))
	if finalSnap.FailedRequests != expectedFailures {
		t.Errorf("expected failures %d, got %d", expectedFailures, finalSnap.FailedRequests)
	}
	if finalSnap.StatusCodes[200] != expectedTotal-expectedFailures {
		t.Errorf("expected status 200 count %d, got %d", expectedTotal-expectedFailures, finalSnap.StatusCodes[200])
	}
}
