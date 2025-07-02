package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WebhookService handles webhook communications between DealDone and n8n
type WebhookService struct {
	config *WebhookConfig
	client *http.Client
	mu     sync.RWMutex
}

// WebhookConfig holds configuration for webhook service
type WebhookConfig struct {
	N8NBaseURL      string            `json:"n8n_base_url"`
	AuthConfig      WebhookAuthConfig `json:"auth_config"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	MaxRetries      int               `json:"max_retries"`
	RetryDelayMs    int               `json:"retry_delay_ms"`
	EnableLogging   bool              `json:"enable_logging"`
	ValidatePayload bool              `json:"validate_payload"`
}

// NewWebhookService creates a new webhook service with configuration
func NewWebhookService(config *WebhookConfig) (*WebhookService, error) {
	if config == nil {
		return nil, fmt.Errorf("webhook config cannot be nil")
	}

	// Validate configuration
	if err := validateWebhookConfig(config); err != nil {
		return nil, fmt.Errorf("invalid webhook config: %w", err)
	}

	// Set defaults
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelayMs == 0 {
		config.RetryDelayMs = 1000
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	service := &WebhookService{
		config: config,
		client: client,
	}

	return service, nil
}

// SendDocumentAnalysisRequest sends a document analysis request to n8n workflow
func (ws *WebhookService) SendDocumentAnalysisRequest(ctx context.Context, payload *DocumentWebhookPayload) error {
	if payload == nil {
		return fmt.Errorf("payload cannot be nil")
	}

	// Validate payload
	if ws.config.ValidatePayload {
		if err := ws.ValidatePayload(payload); err != nil {
			return fmt.Errorf("payload validation failed: %w", err)
		}
	}

	// Set timestamp if not provided
	if payload.Timestamp == 0 {
		payload.Timestamp = time.Now().UnixMilli()
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build webhook URL
	webhookURL := fmt.Sprintf("%s/webhook/dealdone/document-analysis", ws.config.N8NBaseURL)

	// Create request with retry logic
	return ws.sendWithRetry(ctx, "POST", webhookURL, jsonData)
}

// ReceiveProcessingResults handles incoming processing results from n8n
func (ws *WebhookService) ReceiveProcessingResults(w http.ResponseWriter, r *http.Request) (*WebhookResultPayload, error) {
	// Verify request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("invalid method: %s", r.Method)
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	// Verify HMAC signature if enabled
	if ws.config.AuthConfig.EnableHMAC {
		if err := ws.verifyHMACSignature(r, body); err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return nil, fmt.Errorf("HMAC verification failed: %w", err)
		}
	}

	// Verify API key
	if err := ws.verifyAPIKey(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, fmt.Errorf("API key verification failed: %w", err)
	}

	// Parse JSON payload
	var resultPayload WebhookResultPayload
	if err := json.Unmarshal(body, &resultPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validate payload
	if ws.config.ValidatePayload {
		if err := validateWebhookResultPayload(&resultPayload); err != nil {
			http.Error(w, "Payload validation failed", http.StatusBadRequest)
			return nil, fmt.Errorf("payload validation failed: %w", err)
		}
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "received",
		"timestamp": time.Now().UnixMilli(),
	})

	return &resultPayload, nil
}

// QueryJobStatus queries the status of a processing job
func (ws *WebhookService) QueryJobStatus(ctx context.Context, query *WebhookStatusQuery) (*WebhookStatusResponse, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	// Validate query
	if ws.config.ValidatePayload {
		if err := validateWebhookStatusQuery(query); err != nil {
			return nil, fmt.Errorf("query validation failed: %w", err)
		}
	}

	// Set timestamp if not provided
	if query.Timestamp == 0 {
		query.Timestamp = time.Now().UnixMilli()
	}

	// Build status query URL
	statusURL := fmt.Sprintf("%s/webhook/dealdone/status/%s", ws.config.N8NBaseURL, query.JobID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	ws.addAuthHeaders(req, nil)

	// Send request
	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var statusResponse WebhookStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &statusResponse, nil
}

// ValidatePayload validates a document webhook payload
func (ws *WebhookService) ValidatePayload(payload *DocumentWebhookPayload) error {
	return validateDocumentWebhookPayload(payload)
}

// GetValidationResult returns detailed validation results
func (ws *WebhookService) GetValidationResult(payload *DocumentWebhookPayload) *WebhookValidationResult {
	result := &WebhookValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	if err := validateDocumentWebhookPayload(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}

// sendWithRetry sends HTTP request with exponential backoff retry logic
func (ws *WebhookService) sendWithRetry(ctx context.Context, method, url string, body []byte) error {
	var lastErr error

	for attempt := 0; attempt <= ws.config.MaxRetries; attempt++ {
		// Create request
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		ws.addAuthHeaders(req, body)

		// Send request
		resp, err := ws.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < ws.config.MaxRetries {
				time.Sleep(time.Duration(ws.config.RetryDelayMs*(attempt+1)) * time.Millisecond)
				continue
			}
			break
		}

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return nil
		}

		// Read error response
		errorBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(errorBody))

		// Retry on server errors
		if resp.StatusCode >= 500 && attempt < ws.config.MaxRetries {
			time.Sleep(time.Duration(ws.config.RetryDelayMs*(attempt+1)) * time.Millisecond)
			continue
		}

		break
	}

	return fmt.Errorf("failed after %d attempts: %w", ws.config.MaxRetries+1, lastErr)
}

// addAuthHeaders adds authentication headers to HTTP request
func (ws *WebhookService) addAuthHeaders(req *http.Request, body []byte) {
	// Add API key header
	if ws.config.AuthConfig.APIKey != "" {
		req.Header.Set("X-API-Key", ws.config.AuthConfig.APIKey)
	}

	// Add HMAC signature if enabled
	if ws.config.AuthConfig.EnableHMAC && ws.config.AuthConfig.SharedSecret != "" && body != nil {
		signature := ws.generateHMACSignature(body)
		req.Header.Set("X-Signature", "sha256="+signature)
	}

	// Add timestamp
	req.Header.Set("X-Timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
}

// verifyAPIKey verifies the API key from request headers
func (ws *WebhookService) verifyAPIKey(r *http.Request) error {
	if ws.config.AuthConfig.APIKey == "" {
		return nil // API key not configured, skip verification
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		return fmt.Errorf("missing API key header")
	}

	if apiKey != ws.config.AuthConfig.APIKey {
		return fmt.Errorf("invalid API key")
	}

	return nil
}

// verifyHMACSignature verifies HMAC signature from request
func (ws *WebhookService) verifyHMACSignature(r *http.Request, body []byte) error {
	if ws.config.AuthConfig.SharedSecret == "" {
		return fmt.Errorf("HMAC enabled but no shared secret configured")
	}

	signature := r.Header.Get("X-Signature")
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	// Remove 'sha256=' prefix if present
	if strings.HasPrefix(signature, "sha256=") {
		signature = signature[7:]
	}

	expectedSignature := ws.generateHMACSignature(body)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// generateHMACSignature generates HMAC-SHA256 signature for payload
func (ws *WebhookService) generateHMACSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(ws.config.AuthConfig.SharedSecret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// UpdateConfig updates the webhook service configuration
func (ws *WebhookService) UpdateConfig(config *WebhookConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate new configuration
	if err := validateWebhookConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.config = config

	// Update HTTP client timeout if changed
	ws.client.Timeout = time.Duration(config.TimeoutSeconds) * time.Second

	return nil
}

// GetConfig returns a copy of the current configuration
func (ws *WebhookService) GetConfig() *WebhookConfig {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *ws.config
	return &configCopy
}

// IsHealthy performs a health check on the webhook service
func (ws *WebhookService) IsHealthy(ctx context.Context) error {
	// Test connection to n8n
	healthURL := fmt.Sprintf("%s/healthz", ws.config.N8NBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("n8n health check returned status %d", resp.StatusCode)
	}

	return nil
}

// Manual validation functions

// validateWebhookConfig validates webhook configuration
func validateWebhookConfig(config *WebhookConfig) error {
	if config.N8NBaseURL == "" {
		return fmt.Errorf("n8n_base_url is required")
	}

	// Validate URL format
	if _, err := url.Parse(config.N8NBaseURL); err != nil {
		return fmt.Errorf("invalid n8n_base_url format: %w", err)
	}

	if config.TimeoutSeconds < 0 {
		return fmt.Errorf("timeout_seconds must be non-negative")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	if config.RetryDelayMs < 0 {
		return fmt.Errorf("retry_delay_ms must be non-negative")
	}

	return nil
}

// validateDocumentWebhookPayload validates document webhook payload
func validateDocumentWebhookPayload(payload *DocumentWebhookPayload) error {
	if payload.DealName == "" {
		return fmt.Errorf("dealName is required")
	}

	if len(payload.FilePaths) == 0 {
		return fmt.Errorf("filePaths cannot be empty")
	}

	if payload.TriggerType == "" {
		return fmt.Errorf("triggerType is required")
	}

	if payload.JobID == "" {
		return fmt.Errorf("jobId is required")
	}

	if payload.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// validateWebhookResultPayload validates webhook result payload
func validateWebhookResultPayload(payload *WebhookResultPayload) error {
	if payload.JobID == "" {
		return fmt.Errorf("jobId is required")
	}

	if payload.DealName == "" {
		return fmt.Errorf("dealName is required")
	}

	if payload.Status == "" {
		return fmt.Errorf("status is required")
	}

	if payload.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// validateWebhookStatusQuery validates webhook status query
func validateWebhookStatusQuery(query *WebhookStatusQuery) error {
	if query.JobID == "" {
		return fmt.Errorf("jobId is required")
	}

	if query.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}
