package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrometheusMetrics_RegistrationAndHandler(t *testing.T) {
	prom := NewPrometheusMetrics()
	if prom == nil {
		t.Fatal("expected non-nil PrometheusMetrics")
	}

	// Verify handler responds to /metrics
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	prom.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /metrics handler, got %d", w.Code)
	}

	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "loadtest_requests_total") {
		t.Errorf("expected body to contain 'loadtest_requests_total'")
	}
	if !strings.Contains(bodyStr, "loadtest_running_tests") {
		t.Errorf("expected body to contain 'loadtest_running_tests'")
	}
}

func TestPrometheusMetrics_RecordAndLiveGauges(t *testing.T) {
	prom := NewPrometheusMetrics()

	// 1. Record requests
	prom.RecordRequest(200, 15*time.Millisecond, true)
	prom.RecordRequest(200, 25*time.Millisecond, true)
	prom.RecordRequest(500, 50*time.Millisecond, false)

	// 2. Update gauges
	prom.UpdateLiveGauges(150.5, 10)
	prom.SetRunningTests(1)

	// 3. Inspect /metrics output
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	prom.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	output := w.Body.String()

	// Check successful requests count
	if !strings.Contains(output, `loadtest_requests_successful_total 2`) {
		t.Errorf("expected successful requests counter to be 2, output: %s", output)
	}

	// Check failed requests count
	if !strings.Contains(output, `loadtest_requests_failed_total 1`) {
		t.Errorf("expected failed requests counter to be 1, output: %s", output)
	}

	// Check RPS gauge
	if !strings.Contains(output, `loadtest_requests_per_second 150.5`) {
		t.Errorf("expected RPS gauge to be 150.5, output: %s", output)
	}

	// Check active users gauge
	if !strings.Contains(output, `loadtest_active_virtual_users 10`) {
		t.Errorf("expected active virtual users gauge to be 10, output: %s", output)
	}

	// Check running tests gauge
	if !strings.Contains(output, `loadtest_running_tests 1`) {
		t.Errorf("expected running tests gauge to be 1, output: %s", output)
	}
}
