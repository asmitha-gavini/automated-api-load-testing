package engine

import (
	"math"
	"testing"
)

func TestCalculateLatencyStats_Empty(t *testing.T) {
	stats := CalculateLatencyStats([]float64{})
	if stats.MinMs != 0 || stats.MaxMs != 0 || stats.AvgMs != 0 || stats.P50Ms != 0 {
		t.Errorf("expected all zeros for empty slice, got %+v", stats)
	}
}

func TestCalculateLatencyStats_Single(t *testing.T) {
	stats := CalculateLatencyStats([]float64{42.5})
	if stats.MinMs != 42.5 || stats.MaxMs != 42.5 || stats.AvgMs != 42.5 || stats.P50Ms != 42.5 {
		t.Errorf("expected 42.5 for single element, got %+v", stats)
	}
}

func TestCalculateLatencyStats_KnownValues(t *testing.T) {
	// Sample data from 10ms to 100ms in steps of 10ms
	samples := []float64{100, 10, 50, 20, 90, 30, 80, 40, 70, 60}
	stats := CalculateLatencyStats(samples)

	if stats.MinMs != 10.0 {
		t.Errorf("expected MinMs 10.0, got %f", stats.MinMs)
	}
	if stats.MaxMs != 100.0 {
		t.Errorf("expected MaxMs 100.0, got %f", stats.MaxMs)
	}
	if stats.AvgMs != 55.0 {
		t.Errorf("expected AvgMs 55.0, got %f", stats.AvgMs)
	}
	if math.Abs(stats.P50Ms-55.0) > 0.01 {
		t.Errorf("expected P50Ms ~55.0, got %f", stats.P50Ms)
	}
	if stats.P90Ms < 90.0 {
		t.Errorf("expected P90Ms >= 90.0, got %f", stats.P90Ms)
	}
	if stats.P99Ms < 95.0 {
		t.Errorf("expected P99Ms >= 95.0, got %f", stats.P99Ms)
	}
}
