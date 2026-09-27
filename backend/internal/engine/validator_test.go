package engine

import (
	"testing"
)

func TestValidateConfig_Valid(t *testing.T) {
	cfg := &TestConfig{
		TargetURL:       "http://localhost:8081/api/users",
		Method:          "GET",
		VirtualUsers:    10,
		DurationSeconds: 30,
		RampUpSeconds:   5,
		TimeoutMs:       3000,
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}

	if cfg.Method != "GET" {
		t.Errorf("expected normalized method 'GET', got '%s'", cfg.Method)
	}
}

func TestValidateConfig_InvalidURLs(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"not-a-valid-url",
		"ftp://example.com/test",
		"file:///etc/passwd",
	}

	for _, u := range cases {
		cfg := &TestConfig{
			TargetURL:       u,
			Method:          "GET",
			VirtualUsers:    5,
			DurationSeconds: 10,
		}
		if err := ValidateConfig(cfg); err == nil {
			t.Errorf("expected error for invalid URL '%s', got nil", u)
		}
	}
}

func TestValidateConfig_InvalidMethod(t *testing.T) {
	cfg := &TestConfig{
		TargetURL:       "http://localhost:8080/test",
		Method:          "INVALID_METHOD",
		VirtualUsers:    5,
		DurationSeconds: 10,
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Errorf("expected error for unsupported HTTP method, got nil")
	}
}

func TestValidateConfig_VirtualUsersBounds(t *testing.T) {
	badVUs := []int{0, -5, 501, 1000}
	for _, vu := range badVUs {
		cfg := &TestConfig{
			TargetURL:       "http://localhost:8080/test",
			Method:          "GET",
			VirtualUsers:    vu,
			DurationSeconds: 10,
		}
		if err := ValidateConfig(cfg); err == nil {
			t.Errorf("expected error for virtual users %d, got nil", vu)
		}
	}
}

func TestValidateConfig_MissingDurationAndRequests(t *testing.T) {
	cfg := &TestConfig{
		TargetURL:    "http://localhost:8080/test",
		Method:       "GET",
		VirtualUsers: 5,
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Errorf("expected error when neither duration nor total requests provided, got nil")
	}
}

func TestValidateConfig_RampUpExceedsDuration(t *testing.T) {
	cfg := &TestConfig{
		TargetURL:       "http://localhost:8080/test",
		Method:          "GET",
		VirtualUsers:    5,
		DurationSeconds: 10,
		RampUpSeconds:   15,
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Errorf("expected error when ramp up exceeds duration, got nil")
	}
}

func TestValidateConfig_DefaultTimeout(t *testing.T) {
	cfg := &TestConfig{
		TargetURL:       "http://localhost:8080/test",
		Method:          "GET",
		VirtualUsers:    5,
		DurationSeconds: 10,
		TimeoutMs:       0,
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
	if cfg.TimeoutMs != DefaultTimeoutMs {
		t.Errorf("expected default timeout %d, got %d", DefaultTimeoutMs, cfg.TimeoutMs)
	}
}
