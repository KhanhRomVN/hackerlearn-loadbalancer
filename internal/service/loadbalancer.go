package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"hackerlearn-loadbalancer/internal/config"
	"hackerlearn-loadbalancer/internal/logger"
	"hackerlearn-loadbalancer/internal/metrics"
	"hackerlearn-loadbalancer/internal/models"
)

type LoadBalancer struct {
	config     *config.Config
	logger     *logger.Logger
	bestAPI    string
	httpClient *http.Client
	mu         sync.RWMutex
}

func NewLoadBalancer(cfg *config.Config, log *logger.Logger) *LoadBalancer {
	return &LoadBalancer{
		config:     cfg,
		logger:     log,
		httpClient: &http.Client{},
	}
}

func (lb *LoadBalancer) FindBestAPI() {
	var bestScore float64 = -1
	var bestAPI string

	results := make(chan models.MetricsResponse, len(lb.config.MainServiceAPIs))
	var wg sync.WaitGroup

	for _, api := range lb.config.MainServiceAPIs {
		wg.Add(1)
		go func(api string) {
			defer wg.Done()
			serverMetrics, err := metrics.GetServerMetrics(api, lb.httpClient)
			results <- models.MetricsResponse{
				API:     api,
				Metrics: serverMetrics,
				Error:   err,
			}
		}(api)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Error != nil {
			lb.logger.Error(result.Error.Error(), "Failed to get metrics from "+result.API)
			continue
		}

		score := metrics.CalculateScore(result.Metrics)
		if bestScore == -1 || score < bestScore {
			bestScore = score
			bestAPI = result.API
		}
	}

	if bestAPI != "" {
		if lastIndex := strings.LastIndex(bestAPI, "/metrics"); lastIndex != -1 {
			bestAPI = bestAPI[:lastIndex]
		}
		lb.mu.Lock()
		lb.bestAPI = bestAPI
		lb.mu.Unlock()
	}
}

func (lb *LoadBalancer) GetBestAPI() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.bestAPI
}

func (lb *LoadBalancer) HandleRequest(reqData *models.RequestData) (*models.APIResponse, error) {
	bestAPI := lb.GetBestAPI()
	if bestAPI == "" {
		return nil, errors.New("no available API")
	}

	fullAPI := bestAPI + reqData.Endpoint
	var req *http.Request
	var err error

	if reqData.ReqBody != nil {
		reqBody, err := json.Marshal(reqData.ReqBody)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(strings.TrimSuffix(reqData.Method, "PUBLIC"), fullAPI, bytes.NewReader(reqBody))
	} else {
		req, err = http.NewRequest(strings.TrimSuffix(reqData.Method, "PUBLIC"), fullAPI, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if !strings.HasSuffix(reqData.Method, "PUBLIC") && reqData.AccessToken != "" {
		req.Header.Set("accessToken", reqData.AccessToken)
	}

	resp, err := lb.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &models.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}
