package engine

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"backend/internal/metrics"
)

// createHTTPClient builds an optimized HTTP client configured for high concurrency and connection reuse.
func createHTTPClient(timeoutMs int) *http.Client {
	timeout := time.Duration(timeoutMs) * time.Millisecond

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// buildTargetURL constructs the final URL including query parameters.
func buildTargetURL(rawURL string, queryParams map[string]string) string {
	if demoHost := os.Getenv("DEMO_API_HOST"); demoHost != "" {
		if strings.Contains(rawURL, "localhost:8081") {
			rawURL = strings.Replace(rawURL, "localhost:8081", demoHost, 1)
		} else if strings.Contains(rawURL, "127.0.0.1:8081") {
			rawURL = strings.Replace(rawURL, "127.0.0.1:8081", demoHost, 1)
		}
	}

	if len(queryParams) == 0 {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := parsed.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

// runWorker executes requests repeatedly in a dedicated goroutine until the context terminates or max total requests is reached.
func runWorker(
	ctx context.Context,
	workerID int,
	cfg *TestConfig,
	client *http.Client,
	targetURL string,
	collector *MetricsCollector,
	reqCounter *int64,
	prom *metrics.PrometheusMetrics,
) {
	for {
		// 1. Check context termination
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 2. If TotalRequests limit is configured, ensure we do not exceed it
		if cfg.TotalRequests > 0 && reqCounter != nil {
			cur := atomic.AddInt64(reqCounter, 1)
			if cur > cfg.TotalRequests {
				return
			}
		}

		// 3. Prepare request payload
		var bodyReader io.Reader
		if cfg.Body != "" {
			bodyReader = bytes.NewReader([]byte(cfg.Body))
		}

		req, err := http.NewRequestWithContext(ctx, cfg.Method, targetURL, bodyReader)
		if err != nil {
			collector.Record(RequestResult{
				StartTime:  time.Now(),
				EndTime:    time.Now(),
				DurationMs: 0.0,
				StatusCode: 0,
				Success:    false,
				Error:      err.Error(),
			})
			if prom != nil {
				prom.RecordRequest(0, 0, false)
			}
			continue
		}

		// Set headers
		for k, v := range cfg.Headers {
			req.Header.Set(k, v)
		}
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "Automated-Load-Tester/1.0")
		}

		// 4. Execute request and measure latency
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)
		durationMs := float64(duration.Microseconds()) / 1000.0
		end := time.Now()

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			collector.Record(RequestResult{
				StartTime:  start,
				EndTime:    end,
				DurationMs: durationMs,
				StatusCode: 0,
				Success:    false,
				Error:      simplifyError(err),
			})
			if prom != nil {
				prom.RecordRequest(0, duration, false)
			}
			// Small backoff on network/connection errors to prevent CPU spinning
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
			continue
		}

		// 5. Read and drain response body so TCP connection can be returned to pool
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400
		var errStr string
		if !isSuccess {
			errStr = http.StatusText(resp.StatusCode)
		}

		collector.Record(RequestResult{
			StartTime:  start,
			EndTime:    end,
			DurationMs: durationMs,
			StatusCode: resp.StatusCode,
			Success:    isSuccess,
			Error:      errStr,
		})

		if prom != nil {
			prom.RecordRequest(resp.StatusCode, duration, isSuccess)
		}

		// Minimum pacing floor to prevent CPU spin loops when responses are sub-millisecond
		if duration < 2*time.Millisecond {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2*time.Millisecond - duration):
			}
		}
	}
}

func simplifyError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if len(s) > 80 {
		return s[:80]
	}
	return s
}
