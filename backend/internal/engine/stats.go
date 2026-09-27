package engine

import (
	"math"
	"sort"
)

// CalculateLatencyStats calculates statistical summaries and percentiles from duration samples (in ms).
func CalculateLatencyStats(durations []float64) LatencyStats {
	n := len(durations)
	if n == 0 {
		return LatencyStats{}
	}

	// Create a sorted copy to avoid mutating the source
	sorted := make([]float64, n)
	copy(sorted, durations)
	sort.Float64s(sorted)

	var sum float64
	for _, d := range sorted {
		sum += d
	}

	return LatencyStats{
		MinMs: roundToTwoDecimals(sorted[0]),
		MaxMs: roundToTwoDecimals(sorted[n-1]),
		AvgMs: roundToTwoDecimals(sum / float64(n)),
		P50Ms: roundToTwoDecimals(percentile(sorted, 50.0)),
		P90Ms: roundToTwoDecimals(percentile(sorted, 90.0)),
		P95Ms: roundToTwoDecimals(percentile(sorted, 95.0)),
		P99Ms: roundToTwoDecimals(percentile(sorted, 99.0)),
	}
}

// percentile computes the p-th percentile from an already sorted slice using linear interpolation.
func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0.0
	}
	if n == 1 || p <= 0.0 {
		return sorted[0]
	}
	if p >= 100.0 {
		return sorted[n-1]
	}

	index := (p / 100.0) * float64(n-1)
	lower := int(math.Floor(index))
	upper := lower + 1

	if upper >= n {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1.0-weight) + sorted[upper]*weight
}

func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100.0) / 100.0
}
