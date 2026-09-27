package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics encapsulates all Prometheus collectors and registry for load testing telemetry.
type PrometheusMetrics struct {
	registry           *prometheus.Registry
	requestsTotal      *prometheus.CounterVec
	requestsSuccessful prometheus.Counter
	requestsFailed     prometheus.Counter
	requestDuration    prometheus.Histogram
	requestsPerSecond  prometheus.Gauge
	activeVirtualUsers prometheus.Gauge
	runningTests       prometheus.Gauge
}

// NewPrometheusMetrics initializes custom Prometheus metrics using an isolated registry.
func NewPrometheusMetrics() *PrometheusMetrics {
	reg := prometheus.NewRegistry()

	// 1. Total Requests Counter (labeled by outcome status)
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "loadtest_requests_total",
			Help: "Total number of HTTP requests dispatched by the load testing engine.",
		},
		[]string{"status"},
	)

	// 2. Successful Requests Counter
	requestsSuccessful := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "loadtest_requests_successful_total",
			Help: "Total number of successful HTTP requests (HTTP 2xx-3xx).",
		},
	)

	// 3. Failed Requests Counter
	requestsFailed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "loadtest_requests_failed_total",
			Help: "Total number of failed HTTP requests (HTTP 4xx-5xx or network errors).",
		},
	)

	// 4. Request Latency Histogram
	requestDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "loadtest_request_duration_seconds",
			Help:    "Latency distribution of HTTP requests dispatched by the engine in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
	)

	// 5. Requests Per Second Gauge
	requestsPerSecond := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "loadtest_requests_per_second",
			Help: "Current calculated throughput of the running load test.",
		},
	)

	// 6. Active Virtual Users Gauge
	activeVirtualUsers := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "loadtest_active_virtual_users",
			Help: "Current count of active virtual user goroutines generating load.",
		},
	)

	// 7. Running Tests Gauge
	runningTests := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "loadtest_running_tests",
			Help: "Number of load tests currently running (0 or 1).",
		},
	)

	// Register all collectors
	reg.MustRegister(
		requestsTotal,
		requestsSuccessful,
		requestsFailed,
		requestDuration,
		requestsPerSecond,
		activeVirtualUsers,
		runningTests,
	)

	// Pre-initialize label values so metrics are visible immediately upon scrape
	requestsTotal.WithLabelValues("success")
	requestsTotal.WithLabelValues("failed")

	return &PrometheusMetrics{
		registry:           reg,
		requestsTotal:      requestsTotal,
		requestsSuccessful: requestsSuccessful,
		requestsFailed:     requestsFailed,
		requestDuration:    requestDuration,
		requestsPerSecond:  requestsPerSecond,
		activeVirtualUsers: activeVirtualUsers,
		runningTests:       runningTests,
	}
}

// RecordRequest records the outcome and duration of a completed HTTP request.
func (m *PrometheusMetrics) RecordRequest(statusCode int, duration time.Duration, isSuccess bool) {
	if m == nil {
		return
	}

	durationSec := duration.Seconds()
	m.requestDuration.Observe(durationSec)

	if isSuccess {
		m.requestsTotal.WithLabelValues("success").Inc()
		m.requestsSuccessful.Inc()
	} else {
		m.requestsTotal.WithLabelValues("failed").Inc()
		m.requestsFailed.Inc()
	}
}

// UpdateLiveGauges sets instantaneous metrics for throughput and active workers.
func (m *PrometheusMetrics) UpdateLiveGauges(rps float64, activeVUs int) {
	if m == nil {
		return
	}
	m.requestsPerSecond.Set(rps)
	m.activeVirtualUsers.Set(float64(activeVUs))
}

// SetRunningTests updates the running test gauge indicator.
func (m *PrometheusMetrics) SetRunningTests(count float64) {
	if m == nil {
		return
	}
	m.runningTests.Set(count)
}

// Handler returns an HTTP handler to serve the /metrics endpoint.
func (m *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
