package main

import (
	"log"
	"net/http"
	"time"

	"hackerlearn-loadbalancer/internal/api"
	"hackerlearn-loadbalancer/internal/config"
	"hackerlearn-loadbalancer/internal/logger"
	"hackerlearn-loadbalancer/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize logger
	logger := logger.New(true)

	// Initialize load balancer
	lb := service.NewLoadBalancer(cfg, logger)

	// Initialize and start server
	server := api.NewServer(cfg, lb, logger)

	// Start background tasks
	go func() {
		for {
			lb.FindBestAPI()
			time.Sleep(cfg.MetricsInterval)
		}
	}()

	// Start health check
	go func() {
		client := &http.Client{}
		pingURL := cfg.LoadbalancerAPI + "/ping"

		for {
			if _, err := client.Get(pingURL); err != nil {
				logger.Error(err.Error(), "Health check failed")
			}
			time.Sleep(cfg.HealthCheckInterval)
		}
	}()

	// Start server
	logger.Info("Starting server on "+cfg.ServerPort, "Server initialization")
	if err := server.Start(); err != nil {
		logger.Error(err.Error(), "Server failed to start")
		log.Fatal(err)
	}
}
