package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"hackerlearn-loadbalancer/internal/logger"
	"hackerlearn-loadbalancer/internal/models"
	"hackerlearn-loadbalancer/internal/service"
)

type Handlers struct {
	lb     *service.LoadBalancer
	logger *logger.Logger
}

func NewHandlers(lb *service.LoadBalancer, logger *logger.Logger) *Handlers {
	return &Handlers{
		lb:     lb,
		logger: logger,
	}
}

func (h *Handlers) HandleLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var reqData models.RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		h.logger.Error(err.Error(), "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.lb.HandleRequest(&reqData)
	if err != nil {
		h.logger.Error(err.Error(), "Failed to handle request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(resp.Body)))
	w.WriteHeader(resp.StatusCode)
	w.Write(resp.Body)
}

func (h *Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
	response := models.PingResponse{
		Status:    "ok",
		BestAPI:   h.lb.GetBestAPI(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) HandleAPICheck(w http.ResponseWriter, r *http.Request) {
	apiURL := r.URL.Query().Get("api")
	if apiURL == "" {
		h.logger.Error("Missing api parameter", "API URL not provided")
		http.Error(w, "Missing api parameter", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	resp, err := client.Get(apiURL)
	if err != nil {
		h.logger.Error(err.Error(), "Error calling API")
		http.Error(w, "Error calling API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error(err.Error(), "Error reading response")
		http.Error(w, "Error reading response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
