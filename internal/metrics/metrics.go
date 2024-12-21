package metrics

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"hackerlearn-loadbalancer/internal/models"
)

func GetServerMetrics(api string, client *http.Client) (*models.ServerMetrics, error) {
	resp, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	metrics := &models.ServerMetrics{}
	lines := strings.Split(string(body), "\n")

	var reqDurationSum, reqCount float64
	for _, line := range lines {
		switch {
		case strings.Contains(line, "http_request_duration_seconds_sum"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				reqDurationSum, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "http_request_duration_seconds_count"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				reqCount, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "process_cpu_seconds_total"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.CPUUsage, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "process_resident_memory_bytes"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.MemoryUsage, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "http_active_connections"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.ActiveConns, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "http_requests_total"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.RequestsTotal, _ = strconv.ParseFloat(parts[1], 64)
			}
		}
	}

	if reqCount > 0 {
		metrics.ResponseTime = reqDurationSum / reqCount
	}

	return metrics, nil
}

func CalculateScore(metrics *models.ServerMetrics) float64 {
	const (
		responseTimeWeight = 0.35
		cpuUsageWeight     = 0.25
		memoryUsageWeight  = 0.20
		activeConnsWeight  = 0.10
		requestsWeight     = 0.10
	)

	normalizedResponseTime := metrics.ResponseTime / 1.0
	normalizedCPU := metrics.CPUUsage / 100.0
	normalizedMemory := metrics.MemoryUsage / 1000000000.0
	normalizedConns := metrics.ActiveConns / 1000.0
	normalizedRequests := 1.0 - (metrics.RequestsTotal / 10000.0)

	return (normalizedResponseTime * responseTimeWeight) +
		(normalizedCPU * cpuUsageWeight) +
		(normalizedMemory * memoryUsageWeight) +
		(normalizedConns * activeConnsWeight) +
		(normalizedRequests * requestsWeight)
}
