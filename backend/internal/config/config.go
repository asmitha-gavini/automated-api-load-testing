package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration parameters for the backend application.
type Config struct {
	Port           string
	Environment    string
	DBPath         string
	AllowedOrigins []string
}

// LoadConfig loads application configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENV", "development"),
		DBPath:         getEnv("DB_PATH", "loadtest.db"),
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
	}
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

// getEnvInt retrieves an integer environment variable or returns a fallback.
func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

// getEnvSlice retrieves a comma-separated environment variable as a slice of strings.
func getEnvSlice(key string, fallback []string) []string {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	items := strings.Split(val, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
