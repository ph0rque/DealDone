package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// N8nIntegrationService manages interactions with n8n workflows
type N8nIntegrationService struct {
	config         *N8nConfig
	client         *http.Client
	jobTracker     *JobTracker
	webhookService *WebhookService
	requestQueue   chan *N8nWorkflowRequest
	activeRequests map[string]*N8nWorkflowRequest
	mu             sync.RWMutex
	isRunning      bool
	stopChan       chan struct{}
}

// N8nConfig holds configuration for n8n integration
type N8nConfig struct {
	BaseURL               string            `json:"base_url"`
	APIKey                string            `json:"api_key"`
	WorkflowEndpoints     map[string]string `json:"workflow_endpoints"`
	DefaultTimeout        time.Duration     `json:"default_timeout"`
	MaxRetries            int               `json:"max_retries"`
	RetryDelay            time.Duration     `json:"retry_delay"`
	MaxConcurrentJobs     int               `json:"max_concurrent_jobs"`
	EnableBatchProcessing bool              `json:"enable_batch_processing"`
	BatchSize             int               `json:"batch_size"`
	BatchTimeout          time.Duration     `json:"batch_timeout"`
	HealthCheckInterval   time.Duration     `json:"health_check_interval"`
	LogRequests           bool              `json:"log_requests"`
}

// N8nWorkflowRequest represents a request to execute an n8n workflow
type N8nWorkflowRequest struct {
	ID             string                  `json:"id"`
	JobID          string                  `json:"jobId"`
	WorkflowName   string                  `json:"workflowName"`
	Payload        *DocumentWebhookPayload `json:"payload"`
	Priority       int                     `json:"priority"` // 1=high, 2=normal, 3=low
	CreatedAt      int64                   `json:"createdAt"`
	StartedAt      int64                   `json:"startedAt,omitempty"`
	CompletedAt    int64                   `json:"completedAt,omitempty"`
	Status         string                  `json:"status"`
	RetryCount     int                     `json:"retryCount"`
	MaxRetries     int                     `json:"maxRetries"`
	LastError      string                  `json:"lastError,omitempty"`
	N8nExecutionID string                  `json:"n8nExecutionId,omitempty"`
	ResponseData   map[string]interface{}  `json:"responseData,omitempty"`
	Metadata       map[string]interface{}  `json:"metadata,omitempty"`
}

// N8nWorkflowResponse represents the response from n8n workflow execution
type N8nWorkflowResponse struct {
	ExecutionID  string                 `json:"executionId"`
	Status       string                 `json:"status"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Error        string                 `json:"error,omitempty"`
	StartTime    int64                  `json:"startTime"`
	EndTime      int64                  `json:"endTime,omitempty"`
	Duration     int64                  `json:"duration,omitempty"`
	WorkflowName string                 `json:"workflowName"`
	TriggerData  map[string]interface{} `json:"triggerData,omitempty"`
}

// N8nBatchRequest represents a batch of workflow requests
type N8nBatchRequest struct {
	BatchID     string                 `json:"batchId"`
	Requests    []*N8nWorkflowRequest  `json:"requests"`
	CreatedAt   int64                  `json:"createdAt"`
	ProcessedAt int64                  `json:"processedAt,omitempty"`
	CompletedAt int64                  `json:"completedAt,omitempty"`
	Status      string                 `json:"status"`
	Results     []*N8nWorkflowResponse `json:"results,omitempty"`
}

// NewN8nIntegrationService creates a new n8n integration service
func NewN8nIntegrationService(config *N8nConfig, jobTracker *JobTracker, webhookService *WebhookService) (*N8nIntegrationService, error) {
	if config == nil {
		return nil, fmt.Errorf("n8n config cannot be nil")
	}

	// Set defaults
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}
	if config.MaxConcurrentJobs == 0 {
		config.MaxConcurrentJobs = 5
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 5 * time.Minute
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 1 * time.Minute
	}

	// Initialize workflow endpoints if not provided
	if config.WorkflowEndpoints == nil {
		config.WorkflowEndpoints = map[string]string{
			"document-analysis": "/webhook/enhanced-analyze-all-v3",
			"error-handling":    "/webhook/dealdone-error-handling",
			"user-corrections":  "/webhook/dealdone-user-corrections",
			"cleanup":           "/webhook/dealdone-cleanup",
		}
	}

	client := &http.Client{
		Timeout: config.DefaultTimeout,
	}

	service := &N8nIntegrationService{
		config:         config,
		client:         client,
		jobTracker:     jobTracker,
		webhookService: webhookService,
		requestQueue:   make(chan *N8nWorkflowRequest, 1000), // Buffered queue
		activeRequests: make(map[string]*N8nWorkflowRequest),
		stopChan:       make(chan struct{}),
		isRunning:      false,
	}

	return service, nil
}

// Start starts the n8n integration service
func (n8n *N8nIntegrationService) Start() error {
	n8n.mu.Lock()
	defer n8n.mu.Unlock()

	if n8n.isRunning {
		return fmt.Errorf("service already running")
	}

	n8n.isRunning = true

	// Start worker goroutines for processing requests
	for i := 0; i < n8n.config.MaxConcurrentJobs; i++ {
		go n8n.requestWorker(i)
	}

	// Start health check goroutine
	go n8n.healthChecker()

	log.Printf("N8n integration service started with %d workers", n8n.config.MaxConcurrentJobs)
	return nil
}

// Stop stops the n8n integration service
func (n8n *N8nIntegrationService) Stop() error {
	n8n.mu.Lock()
	defer n8n.mu.Unlock()

	if !n8n.isRunning {
		return fmt.Errorf("service not running")
	}

	n8n.isRunning = false
	close(n8n.stopChan)
	close(n8n.requestQueue)

	log.Println("N8n integration service stopped")
	return nil
}

// SendDocumentAnalysisRequest sends a document analysis request to n8n workflow
func (n8n *N8nIntegrationService) SendDocumentAnalysisRequest(ctx context.Context, payload *DocumentWebhookPayload) (*N8nWorkflowRequest, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload cannot be nil")
	}

	// Create workflow request
	request := &N8nWorkflowRequest{
		ID:           fmt.Sprintf("req_%d_%s", time.Now().UnixMilli(), payload.JobID),
		JobID:        payload.JobID,
		WorkflowName: "document-analysis",
		Payload:      payload,
		Priority:     n8n.determinePriority(payload),
		CreatedAt:    time.Now().UnixMilli(),
		Status:       "queued",
		MaxRetries:   n8n.config.MaxRetries,
		Metadata:     make(map[string]interface{}),
	}

	// Add request metadata
	request.Metadata["dealName"] = payload.DealName
	request.Metadata["triggerType"] = string(payload.TriggerType)
	request.Metadata["documentCount"] = len(payload.FilePaths)

	// Update job tracker if available
	if n8n.jobTracker != nil {
		n8n.jobTracker.UpdateJob(payload.JobID, map[string]interface{}{
			"status":      string(JobStatusQueued),
			"currentStep": "Queued for n8n processing",
			"metadata": map[string]interface{}{
				"n8n_request_id": request.ID,
				"workflow_name":  request.WorkflowName,
			},
		})
	}

	// Queue the request
	select {
	case n8n.requestQueue <- request:
		if n8n.config.LogRequests {
			log.Printf("Queued n8n request %s for job %s", request.ID, request.JobID)
		}
		return request, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request queuing canceled: %w", ctx.Err())
	default:
		return nil, fmt.Errorf("request queue full")
	}
}

// GetWorkflowStatus gets the status of a workflow execution
func (n8n *N8nIntegrationService) GetWorkflowStatus(ctx context.Context, executionID string) (*N8nWorkflowResponse, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID cannot be empty")
	}

	// Build status URL
	statusURL := fmt.Sprintf("%s/api/v1/executions/%s", n8n.config.BaseURL, executionID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	// Add authentication
	n8n.addAuthHeaders(req)

	// Send request
	resp, err := n8n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var workflowResp N8nWorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&workflowResp); err != nil {
		return nil, fmt.Errorf("failed to decode workflow status: %w", err)
	}

	return &workflowResp, nil
}

// CancelWorkflow cancels a running workflow execution
func (n8n *N8nIntegrationService) CancelWorkflow(ctx context.Context, executionID string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID cannot be empty")
	}

	// Build cancel URL
	cancelURL := fmt.Sprintf("%s/api/v1/executions/%s/stop", n8n.config.BaseURL, executionID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", cancelURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}

	// Add authentication
	n8n.addAuthHeaders(req)

	// Send request
	resp, err := n8n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to cancel workflow: status %d", resp.StatusCode)
	}

	return nil
}

// GetActiveRequests returns currently active workflow requests
func (n8n *N8nIntegrationService) GetActiveRequests() []*N8nWorkflowRequest {
	n8n.mu.RLock()
	defer n8n.mu.RUnlock()

	requests := make([]*N8nWorkflowRequest, 0, len(n8n.activeRequests))
	for _, req := range n8n.activeRequests {
		requests = append(requests, req)
	}

	return requests
}

// GetRequestStatus gets the status of a specific request
func (n8n *N8nIntegrationService) GetRequestStatus(requestID string) (*N8nWorkflowRequest, error) {
	n8n.mu.RLock()
	defer n8n.mu.RUnlock()

	if req, exists := n8n.activeRequests[requestID]; exists {
		return req, nil
	}

	return nil, fmt.Errorf("request not found: %s", requestID)
}

// Internal methods

// requestWorker processes workflow requests from the queue
func (n8n *N8nIntegrationService) requestWorker(workerID int) {
	log.Printf("N8n worker %d started", workerID)

	for {
		select {
		case request, ok := <-n8n.requestQueue:
			if !ok {
				log.Printf("N8n worker %d stopped (queue closed)", workerID)
				return
			}

			n8n.processWorkflowRequest(request, workerID)

		case <-n8n.stopChan:
			log.Printf("N8n worker %d stopped (shutdown signal)", workerID)
			return
		}
	}
}

// processWorkflowRequest processes a single workflow request
func (n8n *N8nIntegrationService) processWorkflowRequest(request *N8nWorkflowRequest, workerID int) {
	ctx, cancel := context.WithTimeout(context.Background(), n8n.config.DefaultTimeout)
	defer cancel()

	// Add to active requests
	n8n.mu.Lock()
	n8n.activeRequests[request.ID] = request
	request.Status = "processing"
	request.StartedAt = time.Now().UnixMilli()
	n8n.mu.Unlock()

	// Update job tracker
	if n8n.jobTracker != nil {
		n8n.jobTracker.UpdateJob(request.JobID, map[string]interface{}{
			"status":      string(JobStatusProcessing),
			"currentStep": "Executing n8n workflow",
		})
	}

	// Execute workflow with retries
	var response *N8nWorkflowResponse
	var err error

	for attempt := 0; attempt <= request.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(n8n.config.RetryDelay * time.Duration(attempt))
			log.Printf("Retrying n8n request %s (attempt %d/%d)", request.ID, attempt, request.MaxRetries)
		}

		response, err = n8n.executeWorkflow(ctx, request)
		if err == nil {
			break
		}

		request.RetryCount = attempt
		request.LastError = err.Error()

		if n8n.config.LogRequests {
			log.Printf("N8n request %s failed (attempt %d): %v", request.ID, attempt+1, err)
		}
	}

	// Update request status
	n8n.mu.Lock()
	request.CompletedAt = time.Now().UnixMilli()

	if err != nil {
		request.Status = "failed"
		// Update job tracker with failure
		if n8n.jobTracker != nil {
			n8n.jobTracker.FailJob(request.JobID, fmt.Sprintf("N8n workflow failed: %v", err))
		}
	} else {
		request.Status = "completed"
		request.ResponseData = response.Data
		request.N8nExecutionID = response.ExecutionID
		// Job tracker will be updated by webhook result handler
	}

	// Remove from active requests
	delete(n8n.activeRequests, request.ID)
	n8n.mu.Unlock()

	if n8n.config.LogRequests {
		log.Printf("N8n worker %d completed request %s (status: %s)", workerID, request.ID, request.Status)
	}
}

// executeWorkflow executes a workflow in n8n
func (n8n *N8nIntegrationService) executeWorkflow(ctx context.Context, request *N8nWorkflowRequest) (*N8nWorkflowResponse, error) {
	// Get workflow endpoint
	endpoint, exists := n8n.config.WorkflowEndpoints[request.WorkflowName]
	if !exists {
		return nil, fmt.Errorf("unknown workflow: %s", request.WorkflowName)
	}

	// Build URL (handle trailing/leading slashes)
	baseURL := strings.TrimSuffix(n8n.config.BaseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	workflowURL := fmt.Sprintf("%s/%s", baseURL, endpoint)

	// Convert payload to JSON
	jsonData, err := json.Marshal(request.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", workflowURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	n8n.addAuthHeaders(req)

	// Send request
	resp, err := n8n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("workflow execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var workflowResp N8nWorkflowResponse
	if err := json.Unmarshal(body, &workflowResp); err != nil {
		// If parsing fails, create a basic response
		workflowResp = N8nWorkflowResponse{
			Status:       "executed",
			WorkflowName: request.WorkflowName,
			StartTime:    time.Now().UnixMilli(),
			Data:         map[string]interface{}{"rawResponse": string(body)},
		}
	}

	return &workflowResp, nil
}

// addAuthHeaders adds authentication headers to the request
func (n8n *N8nIntegrationService) addAuthHeaders(req *http.Request) {
	if n8n.config.APIKey != "" {
		req.Header.Set("X-N8N-API-KEY", n8n.config.APIKey)
	}
}

// determinePriority determines the priority of a request based on payload
func (n8n *N8nIntegrationService) determinePriority(payload *DocumentWebhookPayload) int {
	// High priority for user-triggered actions
	if payload.TriggerType == TriggerUserButton {
		return 1
	}

	// Normal priority for file changes
	if payload.TriggerType == TriggerFileChange {
		return 2
	}

	// Low priority for scheduled tasks
	return 3
}

// healthChecker periodically checks n8n health
func (n8n *N8nIntegrationService) healthChecker() {
	ticker := time.NewTicker(n8n.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := n8n.checkN8nHealth(); err != nil {
				log.Printf("N8n health check failed: %v", err)
			}
		case <-n8n.stopChan:
			return
		}
	}
}

// checkN8nHealth checks if n8n is responding
func (n8n *N8nIntegrationService) checkN8nHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try multiple endpoints to check health
	endpoints := []string{
		"/healthz",          // Standard health endpoint
		"/api/v1/workflows", // API endpoint that should exist on cloud instances
		"/",                 // Root endpoint as fallback
	}

	var lastErr error
	for _, endpoint := range endpoints {
		healthURL := fmt.Sprintf("%s%s", n8n.config.BaseURL, endpoint)
		req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create health check request: %w", err)
			continue
		}

		// Add auth headers for API endpoints
		if endpoint != "/" {
			n8n.addAuthHeaders(req)
		}

		resp, err := n8n.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("health check request failed: %w", err)
			continue
		}
		resp.Body.Close()

		// Accept various success status codes
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			// 200 = success, 401/403 = server is responding but needs auth (which is fine for health check)
			return nil
		}

		lastErr = fmt.Errorf("n8n health check failed with status %d", resp.StatusCode)
	}

	return lastErr
}

// IsHealthy checks if the n8n integration service is healthy
func (n8n *N8nIntegrationService) IsHealthy(ctx context.Context) error {
	if !n8n.isRunning {
		return fmt.Errorf("n8n integration service not running")
	}

	// Check n8n health
	if err := n8n.checkN8nHealth(); err != nil {
		return fmt.Errorf("n8n health check failed: %w", err)
	}

	// Check queue size
	queueSize := len(n8n.requestQueue)
	if queueSize > 800 { // 80% of queue capacity
		return fmt.Errorf("request queue is nearly full: %d/1000", queueSize)
	}

	return nil
}

// GetStats returns service statistics
func (n8n *N8nIntegrationService) GetStats() map[string]interface{} {
	n8n.mu.RLock()
	defer n8n.mu.RUnlock()

	return map[string]interface{}{
		"isRunning":         n8n.isRunning,
		"queueSize":         len(n8n.requestQueue),
		"activeRequests":    len(n8n.activeRequests),
		"maxConcurrentJobs": n8n.config.MaxConcurrentJobs,
		"batchSize":         n8n.config.BatchSize,
	}
}
