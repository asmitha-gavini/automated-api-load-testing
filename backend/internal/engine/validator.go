package engine

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	MinVirtualUsers  = 1
	MaxVirtualUsers  = 500
	MinDurationSec   = 1
	MaxDurationSec   = 600 // 10 minutes
	MinTotalRequests = 1
	MaxTotalRequests = 1000000
	MaxRampUpSec     = 60
	MinTimeoutMs     = 50
	MaxTimeoutMs     = 30000 // 30 seconds
	DefaultTimeoutMs = 5000
)

var allowedMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

// ValidateConfig verifies that the user test configuration conforms to safety guidelines and limits.
func ValidateConfig(cfg *TestConfig) error {
	if cfg == nil {
		return errors.New("test configuration is nil")
	}

	// 1. Target URL validation
	trimmedURL := strings.TrimSpace(cfg.TargetURL)
	if trimmedURL == "" {
		return errors.New("target_url is required")
	}
	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil || parsedURL.Host == "" {
		return fmt.Errorf("invalid target_url '%s': must be a valid absolute URL", cfg.TargetURL)
	}
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid scheme '%s': only http and https are allowed for safe testing", scheme)
	}

	// 2. HTTP Method validation
	cfg.Method = strings.ToUpper(strings.TrimSpace(cfg.Method))
	if !allowedMethods[cfg.Method] {
		return fmt.Errorf("invalid HTTP method '%s': supported methods are GET, POST, PUT, PATCH, DELETE", cfg.Method)
	}

	// 3. Virtual Users validation
	if cfg.VirtualUsers < MinVirtualUsers || cfg.VirtualUsers > MaxVirtualUsers {
		return fmt.Errorf("virtual_users must be between %d and %d (received %d)", MinVirtualUsers, MaxVirtualUsers, cfg.VirtualUsers)
	}

	// 4. Test Duration or Total Requests validation
	if cfg.DurationSeconds <= 0 && cfg.TotalRequests <= 0 {
		return errors.New("either duration_seconds (1-600) or total_requests (1-1000000) must be specified")
	}

	if cfg.DurationSeconds > 0 {
		if cfg.DurationSeconds < MinDurationSec || cfg.DurationSeconds > MaxDurationSec {
			return fmt.Errorf("duration_seconds must be between %d and %d seconds", MinDurationSec, MaxDurationSec)
		}
	}

	if cfg.TotalRequests > 0 {
		if cfg.TotalRequests < MinTotalRequests || cfg.TotalRequests > MaxTotalRequests {
			return fmt.Errorf("total_requests must be between %d and %d", MinTotalRequests, MaxTotalRequests)
		}
	}

	// 5. Ramp-Up validation
	if cfg.RampUpSeconds < 0 {
		return errors.New("ramp_up_seconds cannot be negative")
	}
	if cfg.RampUpSeconds > MaxRampUpSec {
		return fmt.Errorf("ramp_up_seconds cannot exceed %d seconds", MaxRampUpSec)
	}
	if cfg.DurationSeconds > 0 && cfg.RampUpSeconds >= cfg.DurationSeconds {
		return fmt.Errorf("ramp_up_seconds (%d) must be less than duration_seconds (%d)", cfg.RampUpSeconds, cfg.DurationSeconds)
	}

	// 6. Timeout validation
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = DefaultTimeoutMs
	} else if cfg.TimeoutMs < MinTimeoutMs || cfg.TimeoutMs > MaxTimeoutMs {
		return fmt.Errorf("timeout_ms must be between %d and %d milliseconds", MinTimeoutMs, MaxTimeoutMs)
	}

	return nil
}
