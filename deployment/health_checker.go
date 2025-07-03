package deployment

import (
	"net/http"
	"sync"
	"time"
)

// HealthChecker validates system health before and after deployment
type HealthChecker struct {
	mu              sync.RWMutex
	httpClient      *http.Client
	healthChecks    map[string]*HealthCheck
	healthHistory   []HealthCheckResult
	alertThresholds map[string]float64
	retryConfig     *RetryConfig
	healthStatus    HealthStatus
	lastHealthCheck time.Time
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Name            string            `json:"name"`
	Type            HealthCheckType   `json:"type"`
	Endpoint        string            `json:"endpoint"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	ExpectedStatus  int               `json:"expected_status"`
	ExpectedContent string            `json:"expected_content"`
	Timeout         time.Duration     `json:"timeout"`
	Interval        time.Duration     `json:"interval"`
	Retries         int               `json:"retries"`
	Enabled         bool              `json:"enabled"`
	Critical        bool              `json:"critical"`
	Tags            []string          `json:"tags"`
}

// HealthCheckType represents the type of health check
type HealthCheckType string

const (
	HealthCheckTypeHTTP       HealthCheckType = "http"
	HealthCheckTypeDatabase   HealthCheckType = "database"
	HealthCheckTypeAIProvider HealthCheckType = "ai_provider"
	HealthCheckTypeN8N        HealthCheckType = "n8n"
	HealthCheckTypeRedis      HealthCheckType = "redis"
	HealthCheckTypeFileSystem HealthCheckType = "filesystem"
	HealthCheckTypeCustom     HealthCheckType = "custom"
)

// HealthStatus represents overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// RetryConfig represents retry configuration for health checks
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		healthChecks:    make(map[string]*HealthCheck),
		healthHistory:   make([]HealthCheckResult, 0),
		alertThresholds: make(map[string]float64),
		retryConfig: &RetryConfig{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		},
		healthStatus: HealthStatusUnknown,
	}
}

// CheckHealth performs health checks on the provided endpoints
func (hc *HealthChecker) CheckHealth(endpoints []string) []HealthCheckResult {
	results := make([]HealthCheckResult, 0, len(endpoints))

	for _, endpoint := range endpoints {
		result := hc.checkEndpoint(endpoint)
		results = append(results, result)
	}

	hc.updateHealthHistory(results)
	return results
}

// checkEndpoint performs a health check on a single endpoint
func (hc *HealthChecker) checkEndpoint(endpoint string) HealthCheckResult {
	start := time.Now()

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return HealthCheckResult{
			Endpoint:     endpoint,
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Timestamp:    time.Now(),
			Error:        err.Error(),
		}
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return HealthCheckResult{
			Endpoint:     endpoint,
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Timestamp:    time.Now(),
			Error:        err.Error(),
		}
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "unhealthy"
	}

	return HealthCheckResult{
		Endpoint:     endpoint,
		Status:       status,
		ResponseTime: time.Since(start),
		Timestamp:    time.Now(),
		StatusCode:   resp.StatusCode,
	}
}

// updateHealthHistory updates the health check history
func (hc *HealthChecker) updateHealthHistory(results []HealthCheckResult) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.healthHistory = append(hc.healthHistory, results...)

	// Keep only last 1000 results
	if len(hc.healthHistory) > 1000 {
		hc.healthHistory = hc.healthHistory[len(hc.healthHistory)-1000:]
	}

	hc.lastHealthCheck = time.Now()
}
