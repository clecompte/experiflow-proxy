package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the proxy configuration
type Config struct {
	// Proxy settings
	Port         string
	OriginURL    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// ExperiFlow API settings
	APIBaseURL string
	EdgeToken  string
	Timeout    time.Duration

	// Feature flags
	FailOpen      bool
	EnableLogging bool
	EnableMetrics bool
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	return &Config{
		Port:          getEnv("PORT", "8090"),
		OriginURL:     getEnv("ORIGIN_URL", "http://localhost:8080"),
		ReadTimeout:   getDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:  getDuration("WRITE_TIMEOUT", 10*time.Second),
		APIBaseURL:    getEnv("EXPERIFLOW_API_URL", "http://localhost:8000"),
		EdgeToken:     getEnv("EXPERIFLOW_EDGE_TOKEN", ""),
		Timeout:       getDuration("TRANSFORM_TIMEOUT", 50*time.Millisecond),
		FailOpen:      getBool("FAIL_OPEN", true),
		EnableLogging: getBool("ENABLE_LOGGING", true),
		EnableMetrics: getBool("ENABLE_METRICS", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
