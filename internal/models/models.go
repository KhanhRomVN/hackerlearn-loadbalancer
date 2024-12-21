package models

type RequestData struct {
	Endpoint    string      `json:"endpoint"`
	ReqBody     interface{} `json:"reqBody,omitempty"`
	AccessToken string      `json:"accessToken,omitempty"`
	Method      string      `json:"method"`
}

type ServerMetrics struct {
	ResponseTime  float64
	CPUUsage      float64
	MemoryUsage   float64
	ActiveConns   float64
	RequestsTotal float64
	Score         float64
}

type PingResponse struct {
	Status    string `json:"status"`
	BestAPI   string `json:"bestAPI"`
	Timestamp string `json:"timestamp"`
}

type APIResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type MetricsResponse struct {
	API     string
	Metrics *ServerMetrics
	Error   error
}
