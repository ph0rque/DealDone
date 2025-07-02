package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestNewWebhookHandlers(t *testing.T) {
	app := &App{}
	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)
	if handlers == nil {
		t.Errorf("expected webhook handlers but got nil")
	}
	if handlers.app != app {
		t.Errorf("expected app reference to be set")
	}
	if handlers.webhookService != webhookService {
		t.Errorf("expected webhook service reference to be set")
	}
	if handlers.resultChannel == nil {
		t.Errorf("expected result channel to be initialized")
	}
}

func TestWebhookHandlers_HandleProcessingResults(t *testing.T) {
	// Create test app with services
	tempDir := t.TempDir()
	configService := &ConfigService{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	configService.SetDealDoneRoot(tempDir)

	app := &App{
		configService: configService,
		folderManager: NewFolderManager(configService),
		jobTracker:    NewJobTracker(configService),
	}

	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
			AuthConfig: WebhookAuthConfig{
				APIKey:       "test-key",
				SharedSecret: "test-secret",
				EnableHMAC:   true,
			},
			ValidatePayload: true,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupAuth      bool
		expectedStatus int
	}{
		{
			name:   "valid processing result",
			method: "POST",
			body: &WebhookResultPayload{
				JobID:              "job-123",
				DealName:           "TestDeal",
				Status:             "completed",
				ProcessedDocuments: 3,
				AverageConfidence:  0.85,
				TemplatesUpdated:   []string{"template1.xlsx"},
				ProcessingTime:     25500,
				Timestamp:          time.Now().UnixMilli(),
			},
			setupAuth:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS request (CORS preflight)",
			method:         "OPTIONS",
			setupAuth:      false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET method not allowed",
			method:         "GET",
			setupAuth:      false,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON body",
			method:         "POST",
			body:           "invalid json",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyBytes = []byte(str)
				} else {
					bodyBytes, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/webhook/results", bytes.NewBuffer(bodyBytes))
			if tt.setupAuth {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-API-Key", "test-key")

				// Generate HMAC signature
				timestamp := strconv.FormatInt(time.Now().Unix(), 10)
				req.Header.Set("X-Timestamp", timestamp)

				if len(bodyBytes) > 0 {
					message := fmt.Sprintf("%s|%s|%s|%s", tt.method, "/webhook/results", timestamp, string(bodyBytes))
					mac := hmac.New(sha256.New, []byte("test-secret"))
					mac.Write([]byte(message))
					signature := hex.EncodeToString(mac.Sum(nil))
					req.Header.Set("X-Signature", signature)
				}
			}

			w := httptest.NewRecorder()
			handlers.HandleProcessingResults(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("expected CORS origin *, got %s", origin)
			}
		})
	}
}

func TestWebhookHandlers_HandleStatusQuery(t *testing.T) {
	// Create test app with job tracker
	tempDir := t.TempDir()
	configService := &ConfigService{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	configService.SetDealDoneRoot(tempDir)

	jobTracker := NewJobTracker(configService)

	// Add a test job
	jobInfo := jobTracker.CreateJob("job-123", "TestDeal", TriggerFileChange, []string{"/test/doc.pdf"})
	jobInfo.Status = JobStatusProcessing
	jobInfo.Progress = 0.75

	app := &App{
		configService: configService,
		jobTracker:    jobTracker,
	}

	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	tests := []struct {
		name           string
		method         string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "valid status query",
			method:         "GET",
			queryParams:    "?jobId=job-123&dealName=TestDeal",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing job ID",
			method:         "GET",
			queryParams:    "?dealName=TestDeal",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent job",
			method:         "GET",
			queryParams:    "?jobId=job-404&dealName=TestDeal",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/webhook/status" + tt.queryParams
			req := httptest.NewRequest(tt.method, url, nil)
			w := httptest.NewRecorder()

			handlers.HandleStatusQuery(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("expected CORS origin *, got %s", origin)
			}

			// For successful requests, check response format
			if tt.expectedStatus == http.StatusOK && tt.method == "GET" {
				var response WebhookStatusResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestWebhookHandlers_HandleHealthCheck(t *testing.T) {
	// Create minimal app for health check
	app := &App{
		folderManager: &FolderManager{},
		aiService:     &AIService{},
		jobTracker:    &JobTracker{},
	}

	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET health check",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/webhook/health", nil)
			w := httptest.NewRecorder()

			handlers.HandleHealthCheck(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("expected CORS origin *, got %s", origin)
			}

			if tt.checkResponse && tt.method == "GET" {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode health response: %v", err)
				}

				if status, ok := response["status"]; !ok || status == "" {
					t.Errorf("expected status field in health response")
				}

				if timestamp, ok := response["timestamp"]; !ok || timestamp == nil {
					t.Errorf("expected timestamp field in health response")
				}

				if checks, ok := response["checks"]; !ok || checks == nil {
					t.Errorf("expected checks field in health response")
				}
			}
		})
	}
}

func TestWebhookHandlers_ResultProcessor(t *testing.T) {
	// Create test app with all services
	tempDir := t.TempDir()
	configService := &ConfigService{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	configService.SetDealDoneRoot(tempDir)

	app := &App{
		configService: configService,
		folderManager: NewFolderManager(configService),
		jobTracker:    NewJobTracker(configService),
	}

	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	// Create a job to process results for
	jobInfo := app.jobTracker.CreateJob("job-123", "TestDeal", TriggerFileChange, []string{"/test/doc.pdf"})
	jobInfo.Status = JobStatusProcessing

	// Start the result processor
	handlers.StartListening()
	defer handlers.StopListening()

	// Test processing results
	tests := []struct {
		name   string
		result *WebhookResultPayload
	}{
		{
			name: "completed job",
			result: &WebhookResultPayload{
				JobID:              "job-123",
				DealName:           "TestDeal",
				Status:             "completed",
				ProcessedDocuments: 5,
				AverageConfidence:  0.85,
				TemplatesUpdated:   []string{"template1.xlsx"},
				ProcessingTime:     30500,
				Timestamp:          time.Now().UnixMilli(),
			},
		},
		{
			name: "failed job",
			result: &WebhookResultPayload{
				JobID:    "job-123",
				DealName: "TestDeal",
				Status:   "failed",
				Errors: []ProcessingError{
					{
						Code:        "PROCESSING_ERROR",
						Message:     "Failed to process document",
						Level:       "error",
						Source:      "test",
						Timestamp:   time.Now().UnixMilli(),
						Recoverable: false,
					},
				},
				Timestamp: time.Now().UnixMilli(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send result to channel
			select {
			case handlers.resultChannel <- tt.result:
				// Give processor time to handle the result
				time.Sleep(100 * time.Millisecond)
			case <-time.After(1 * time.Second):
				t.Errorf("failed to send result to channel")
			}

			// Verify job status was updated
			job, err := app.jobTracker.GetJob(tt.result.JobID)
			if err != nil {
				t.Errorf("failed to get job after processing: %v", err)
			}

			if tt.result.Status == "completed" {
				if job.Status != JobStatusCompleted {
					t.Errorf("expected job status to be completed, got %s", job.Status)
				}
			} else if tt.result.Status == "failed" {
				if job.Status != JobStatusFailed {
					t.Errorf("expected job status to be failed, got %s", job.Status)
				}
			}
		})
	}
}

func TestWebhookHandlers_CreateHTTPServer(t *testing.T) {
	app := &App{}
	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:     "http://localhost:5678",
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryDelayMs:   1000,
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	server := handlers.CreateHTTPServer(8080)
	if server == nil {
		t.Errorf("expected HTTP server but got nil")
	}

	if server.Addr != ":8080" {
		t.Errorf("expected server address :8080, got %s", server.Addr)
	}

	if server.ReadTimeout != 30*time.Second {
		t.Errorf("expected read timeout 30s, got %v", server.ReadTimeout)
	}

	if server.WriteTimeout != 30*time.Second {
		t.Errorf("expected write timeout 30s, got %v", server.WriteTimeout)
	}

	if server.IdleTimeout != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", server.IdleTimeout)
	}
}

func TestWebhookHandlers_UpdateTemplateFiles(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()
	configService := &ConfigService{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	configService.SetDealDoneRoot(tempDir)

	folderManager := NewFolderManager(configService)
	folderManager.InitializeFolderStructure()

	app := &App{
		configService: configService,
		folderManager: folderManager,
	}

	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL: "http://localhost:5678",
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	// Create test deal folder
	dealName := "TestDeal"
	dealPath, err := folderManager.CreateDealFolder(dealName)
	if err != nil {
		t.Fatalf("failed to create deal folder: %v", err)
	}

	result := &WebhookResultPayload{
		JobID:            "job-123",
		DealName:         dealName,
		Status:           "completed",
		TemplatesUpdated: []string{"template1.xlsx", "template2.xlsx"},
	}

	err = handlers.updateTemplateFiles(result)
	if err != nil {
		t.Errorf("unexpected error updating template files: %v", err)
	}

	// Verify analysis folder was created
	analysisPath := filepath.Join(dealPath, "analysis")
	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		t.Errorf("analysis folder was not created")
	}
}

func TestWebhookHandlers_StartStopListening(t *testing.T) {
	app := &App{}
	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL: "http://localhost:5678",
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	// Test starting listening
	handlers.StartListening()
	if !handlers.isRunning {
		t.Errorf("expected handlers to be running after StartListening")
	}

	// Test starting again (should not crash)
	handlers.StartListening()

	// Test stopping listening
	handlers.StopListening()
	if handlers.isRunning {
		t.Errorf("expected handlers to not be running after StopListening")
	}

	// Test stopping again (should not crash)
	handlers.StopListening()
}

func TestWebhookHandlers_GetResultChannel(t *testing.T) {
	app := &App{}
	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL: "http://localhost:5678",
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	channel := handlers.GetResultChannel()
	if channel == nil {
		t.Errorf("expected result channel but got nil")
	}

	// Test that we can receive from the channel
	go func() {
		handlers.resultChannel <- &WebhookResultPayload{
			JobID: "test-job",
		}
	}()

	select {
	case result := <-channel:
		if result.JobID != "test-job" {
			t.Errorf("expected job ID 'test-job', got %s", result.JobID)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("timeout waiting for result from channel")
	}
}

// Benchmark tests
func BenchmarkWebhookHandlers_HandleProcessingResults(b *testing.B) {
	app := &App{}
	webhookService := &WebhookService{
		config: &WebhookConfig{
			N8NBaseURL:      "http://localhost:5678",
			ValidatePayload: false, // Disable validation for performance
		},
	}

	handlers := NewWebhookHandlers(app, webhookService)

	payload := &WebhookResultPayload{
		JobID:              "bench-job",
		DealName:           "BenchDeal",
		Status:             "completed",
		ProcessedDocuments: 5,
		AverageConfidence:  0.85,
	}

	bodyBytes, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/webhook/results", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.HandleProcessingResults(w, req)
	}
}
