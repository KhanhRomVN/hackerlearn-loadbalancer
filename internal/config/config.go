package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MainServiceAPIs     []string
	LoadbalancerAPI     string
	ServerPort          string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	MetricsInterval     time.Duration
	HealthCheckInterval time.Duration
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists, but don't return error if it doesn't
	godotenv.Load()

	// Get environment variables with defaults
	mainAPIs := os.Getenv("LIST_OF_MAIN_SERVICE_API")
	if mainAPIs == "" {
		return nil, errors.New("LIST_OF_MAIN_SERVICE_API environment variable is required")
	}

	loadbalancerAPI := os.Getenv("LOADBALANCER_API")
	if loadbalancerAPI == "" {
		return nil, errors.New("LOADBALANCER_API environment variable is required")
	}

	// Split the APIs and validate
	apiList := strings.Split(mainAPIs, ",")
	if len(apiList) == 0 {
		return nil, errors.New("at least one API must be provided in LIST_OF_MAIN_SERVICE_API")
	}

	return &Config{
		MainServiceAPIs:     apiList,
		LoadbalancerAPI:     loadbalancerAPI,
		ServerPort:          getEnvWithDefault("SERVER_PORT", ":9090"),
		ReadTimeout:         time.Duration(getEnvAsInt("READ_TIMEOUT_SECONDS", 15)) * time.Second,
		WriteTimeout:        time.Duration(getEnvAsInt("WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
		MetricsInterval:     time.Duration(getEnvAsInt("METRICS_INTERVAL_SECONDS", 60)) * time.Second,
		HealthCheckInterval: time.Duration(getEnvAsInt("HEALTH_CHECK_INTERVAL_SECONDS", 60)) * time.Second,
	}, nil
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
