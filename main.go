package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type RequestData struct {
	Endpoint    string      `json:"endpoint"`
	ReqBody     interface{} `json:"reqBody,omitempty"`
	AccessToken string      `json:"accessToken,omitempty"`
	Method      string      `json:"method"`
}

type ServerMetrics struct {
	responseTime  float64
	cpuUsage      float64
	memoryUsage   float64
	activeConns   float64
	requestsTotal float64
	score         float64
}

type PingResponse struct {
	Status    string `json:"status"`
	BestAPI   string `json:"bestAPI"`
	Timestamp string `json:"timestamp"`
}

var bestAPI string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func FindBestAPI() string {
	apis := strings.Split(os.Getenv("LIST_OF_MAIN_SERVICE_API"), ",")
	if len(apis) == 0 {
		log.Println("Error: LIST_OF_MAIN_SERVICE_API is empty")
		return ""
	}

	client := &http.Client{}
	var bestAPI string
	var bestScore float64 = -1

	for _, api := range apis {
		metrics := getServerMetrics(api, client)
		if metrics == nil {
			continue
		}

		score := calculateScore(*metrics)

		if bestScore == -1 || score < bestScore {
			bestScore = score
			bestAPI = api
		}
	}

	if bestAPI == "" {
		log.Println("No available APIs found")
		return ""
	}

	if lastIndex := strings.LastIndex(bestAPI, "/metrics"); lastIndex != -1 {
		bestAPI = bestAPI[:lastIndex]
	} else {
		log.Println("Error: '/metrics' not found in bestAPI")
		return ""
	}

	log.Printf("Best API found: %s (score: %.2f)", bestAPI, bestScore)
	return bestAPI
}

func getServerMetrics(api string, client *http.Client) *ServerMetrics {
	resp, err := client.Get(api)
	if err != nil {
		log.Println("Error calling API:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return nil
	}

	lines := strings.Split(string(body), "\n")
	metrics := &ServerMetrics{}

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
				metrics.cpuUsage, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "process_resident_memory_bytes"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.memoryUsage, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "http_active_connections"):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.activeConns, _ = strconv.ParseFloat(parts[1], 64)
			}
		case strings.Contains(line, "http_requests_total") && strings.Contains(line, "status=\"200\""):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				metrics.requestsTotal, _ = strconv.ParseFloat(parts[1], 64)
			}
		}
	}

	if reqCount > 0 {
		metrics.responseTime = reqDurationSum / reqCount
	}

	return metrics
}

func calculateScore(metrics ServerMetrics) float64 {
	const (
		responseTimeWeight = 0.35
		cpuUsageWeight     = 0.25
		memoryUsageWeight  = 0.20
		activeConnsWeight  = 0.10
		requestsWeight     = 0.10
	)

	normalizedResponseTime := metrics.responseTime / 1.0
	normalizedCPU := metrics.cpuUsage / 100.0
	normalizedMemory := metrics.memoryUsage / 1000000000.0
	normalizedConns := metrics.activeConns / 1000.0
	normalizedRequests := 1.0 - (metrics.requestsTotal / 10000.0)

	score := (normalizedResponseTime * responseTimeWeight) +
		(normalizedCPU * cpuUsageWeight) +
		(normalizedMemory * memoryUsageWeight) +
		(normalizedConns * activeConnsWeight) +
		(normalizedRequests * requestsWeight)

	return score
}

func handleAPICheck(w http.ResponseWriter, r *http.Request) {
	apiURL := r.URL.Query().Get("api")
	if apiURL == "" {
		http.Error(w, "Missing api parameter", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	resp, err := client.Get(apiURL)
	if err != nil {
		http.Error(w, "Error calling API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	response := PingResponse{
		Status:    "ok",
		BestAPI:   bestAPI,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	fmt.Println(response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/loadbalancer", handleLoadBalancer).Methods("POST", "OPTIONS")
	r.HandleFunc("/api-check", handleAPICheck).Methods("GET")
	r.HandleFunc("/ping", handlePing).Methods("GET")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Authorization", "accessToken"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Handler:      corsHandler.Handler(r),
		Addr:         ":9090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start the FindBestAPI goroutine
	go func() {
		for {
			bestAPI = FindBestAPI()
			time.Sleep(1 * time.Minute)
		}
	}()

	// Add new goroutine for pinging loadbalancer
	go func() {
		loadbalancerAPI := os.Getenv("LOADBALANCER_API")
		if loadbalancerAPI == "" {
			log.Println("Warning: LOADBALANCER_API not set in environment")
			return
		}

		pingURL := loadbalancerAPI + "/ping"
		client := &http.Client{}

		for {
			resp, err := client.Get(pingURL)
			if err != nil {
				log.Printf("Error pinging loadbalancer: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("Successfully pinged loadbalancer at %s", pingURL)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	log.Println("Starting load balancer on :9090")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var requestData RequestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	isPublicMethod := strings.HasSuffix(requestData.Method, "PUBLIC")
	method := strings.TrimSuffix(requestData.Method, "PUBLIC")

	if bestAPI == "" {
		http.Error(w, "No available API", http.StatusServiceUnavailable)
		return
	}

	fullAPI := bestAPI + requestData.Endpoint
	log.Printf("Forwarding request to: %s", fullAPI)

	client := &http.Client{}
	var req *http.Request
	var err error

	if requestData.ReqBody != nil {
		reqBody, err := json.Marshal(requestData.ReqBody)
		if err != nil {
			http.Error(w, "Error marshalling request body", http.StatusInternalServerError)
			return
		}
		req, err = http.NewRequest(method, fullAPI, bytes.NewReader(reqBody))
	} else {
		req, err = http.NewRequest(method, fullAPI, nil)
	}

	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if !isPublicMethod && requestData.AccessToken != "" {
		req.Header.Set("accessToken", requestData.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	log.Printf("Response Headers: %+v", resp.Header)
	log.Printf("Response Body Length: %d", len(responseBody))

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))

	w.WriteHeader(resp.StatusCode)

	_, err = w.Write(responseBody)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
