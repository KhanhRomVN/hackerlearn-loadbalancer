package config

import (
	"errors"
	"os"
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
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	mainAPIs := strings.Split(os.Getenv("LIST_OF_MAIN_SERVICE_API"), ",")
	if len(mainAPIs) == 0 {
		return nil, errors.New("LIST_OF_MAIN_SERVICE_API is required")
	}

	loadbalancerAPI := os.Getenv("LOADBALANCER_API")
	if loadbalancerAPI == "" {
		return nil, errors.New("LOADBALANCER_API is required")
	}

	return &Config{
		MainServiceAPIs:     mainAPIs,
		LoadbalancerAPI:     loadbalancerAPI,
		ServerPort:          ":9090",
		ReadTimeout:         15 * time.Second,
		WriteTimeout:        15 * time.Second,
		MetricsInterval:     time.Minute,
		HealthCheckInterval: time.Minute,
	}, nil
}
