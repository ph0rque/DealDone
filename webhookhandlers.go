package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// WebhookHandlers manages HTTP handlers for webhook communications
type WebhookHandlers struct {
	app            *App
	webhookService *WebhookService
	resultChannel  chan *WebhookResultPayload
	mu             sync.RWMutex
	isRunning      bool
}

// NewWebhookHandlers creates a new webhook handlers instance
func NewWebhookHandlers(app *App, webhookService *WebhookService) *WebhookHandlers {
	return &WebhookHandlers{
		app:            app,
		webhookService: webhookService,
		resultChannel:  make(chan *WebhookResultPayload, 100), // Buffered channel for results
		isRunning:      false,
	}
}

// HandleProcessingResults handles incoming processing results from n8n workflows
func (wh *WebhookHandlers) HandleProcessingResults(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for cross-origin requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Signature, X-Timestamp")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Use the webhook service to receive and validate the payload
	resultPayload, err := wh.webhookService.ReceiveProcessingResults(w, r)
	if err != nil {
		// Error response already sent by webhook service
		log.Printf("Error receiving processing results: %v", err)
		return
	}

	// Process the results asynchronously
	go wh.processWebhookResults(resultPayload)

	// Send the result to the channel for any listeners
	select {
	case wh.resultChannel <- resultPayload:
		// Successfully queued
	default:
		// Channel full, log warning but don't block
		log.Printf("Warning: Result channel full, dropping result for job %s", resultPayload.JobID)
	}
}

// HandleStatusQuery handles status queries from the frontend or n8n
func (wh *WebhookHandlers) HandleStatusQuery(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Timestamp")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract job ID from URL path
	jobID := r.URL.Query().Get("jobId")
	if jobID == "" {
		http.Error(w, "Missing jobId parameter", http.StatusBadRequest)
		return
	}

	// Create status query
	query := &WebhookStatusQuery{
		JobID:     jobID,
		DealName:  r.URL.Query().Get("dealName"),
		Timestamp: time.Now().UnixMilli(),
	}

	// Query status from n8n
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	statusResponse, err := wh.webhookService.QueryJobStatus(ctx, query)
	if err != nil {
		log.Printf("Error querying job status: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query status: %v", err), http.StatusInternalServerError)
		return
	}

	// Return status response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(statusResponse)
}

// HandleHealthCheck handles health check requests
func (wh *WebhookHandlers) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Perform health checks
	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UnixMilli(),
		"checks":    make(map[string]interface{}),
	}

	checks := healthStatus["checks"].(map[string]interface{})

	// Check webhook service health
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := wh.webhookService.IsHealthy(ctx); err != nil {
		checks["webhook_service"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		healthStatus["status"] = "unhealthy"
	} else {
		checks["webhook_service"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check app services
	if wh.app != nil {
		checks["app_services"] = map[string]interface{}{
			"status":          "healthy",
			"document_router": wh.app.documentRouter != nil,
			"ai_service":      wh.app.aiService != nil,
			"folder_manager":  wh.app.folderManager != nil,
			"job_tracker":     wh.app.jobTracker != nil,
		}

		// Check job tracker health specifically
		if wh.app.jobTracker != nil {
			if err := wh.app.jobTracker.IsHealthy(ctx); err != nil {
				checks["job_tracker"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				healthStatus["status"] = "unhealthy"
			} else {
				checks["job_tracker"] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	// Set response status based on overall health
	statusCode := http.StatusOK
	if healthStatus["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthStatus)
}

// GetResultChannel returns the channel for listening to webhook results
func (wh *WebhookHandlers) GetResultChannel() <-chan *WebhookResultPayload {
	return wh.resultChannel
}

// StartListening starts listening for webhook results and processes them
func (wh *WebhookHandlers) StartListening() {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	if wh.isRunning {
		return // Already running
	}

	wh.isRunning = true
	go wh.resultProcessor()
}

// StopListening stops the webhook result processor
func (wh *WebhookHandlers) StopListening() {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	if !wh.isRunning {
		return // Not running
	}

	wh.isRunning = false
	close(wh.resultChannel)
}

// processWebhookResults processes webhook results and updates the application state
func (wh *WebhookHandlers) processWebhookResults(result *WebhookResultPayload) {
	log.Printf("Processing webhook result for job %s, deal %s", result.JobID, result.DealName)

	// Update job status in tracker
	if wh.app != nil && wh.app.jobTracker != nil {
		if result.Status == "completed" {
			if err := wh.app.jobTracker.CompleteJob(result.JobID, result); err != nil {
				log.Printf("Error completing job in tracker: %v", err)
			}
		} else if result.Status == "failed" {
			errorMsg := "Processing failed"
			if len(result.Errors) > 0 {
				errorMsg = result.Errors[0].Message
			}
			if err := wh.app.jobTracker.FailJob(result.JobID, errorMsg); err != nil {
				log.Printf("Error marking job as failed in tracker: %v", err)
			}
		} else {
			// Update progress and status
			updates := map[string]interface{}{
				"status":             result.Status,
				"progress":           result.AverageConfidence,
				"processedDocuments": result.ProcessedDocuments,
				"currentStep":        "Processing",
			}
			if len(result.Errors) > 0 {
				updates["errors"] = result.Errors
			}

			if err := wh.app.jobTracker.UpdateJob(result.JobID, updates); err != nil {
				log.Printf("Error updating job in tracker: %v", err)
			}
		}
	}

	// Update deal folder structure if templates were updated
	if len(result.TemplatesUpdated) > 0 {
		if err := wh.updateTemplateFiles(result); err != nil {
			log.Printf("Error updating template files: %v", err)
		}
	}

	// Log processing completion
	log.Printf("Completed processing for job %s: %d documents processed, %.2f average confidence",
		result.JobID, result.ProcessedDocuments, result.AverageConfidence)

	// Handle any errors
	if len(result.Errors) > 0 {
		log.Printf("Processing errors for job %s: %v", result.JobID, result.Errors)
	}
}

// updateTemplateFiles updates template files based on webhook results
func (wh *WebhookHandlers) updateTemplateFiles(result *WebhookResultPayload) error {
	if wh.app.folderManager == nil {
		return fmt.Errorf("folder manager not available")
	}

	// Get deal folder path
	dealFolderPath := wh.app.folderManager.GetDealPath(result.DealName)
	analysisFolderPath := filepath.Join(dealFolderPath, "analysis")

	// Check if analysis folder exists, create if not
	if err := wh.app.folderManager.EnsureFolderExists(analysisFolderPath); err != nil {
		return fmt.Errorf("failed to ensure analysis folder exists: %w", err)
	}

	// Process each updated template
	for _, templateName := range result.TemplatesUpdated {
		log.Printf("Template updated: %s in deal %s", templateName, result.DealName)

		// Additional processing can be added here:
		// - Validate template integrity
		// - Update template metadata
		// - Trigger frontend notifications
	}

	return nil
}

// resultProcessor continuously processes results from the channel
func (wh *WebhookHandlers) resultProcessor() {
	for {
		wh.mu.RLock()
		isRunning := wh.isRunning
		wh.mu.RUnlock()

		if !isRunning {
			break
		}

		select {
		case result, ok := <-wh.resultChannel:
			if !ok {
				// Channel closed, exit
				return
			}
			wh.processWebhookResults(result)
		case <-time.After(1 * time.Second):
			// Periodic check to see if we should continue
			continue
		}
	}
}

// RegisterHandlers registers all webhook handlers with an HTTP mux
func (wh *WebhookHandlers) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/webhook/results", wh.HandleProcessingResults)
	mux.HandleFunc("/webhook/status", wh.HandleStatusQuery)
	mux.HandleFunc("/webhook/health", wh.HandleHealthCheck)

	// Template analysis endpoints for n8n workflows
	mux.HandleFunc("/discover-templates", wh.HandleDiscoverTemplates)
	mux.HandleFunc("/extract-document-fields", wh.HandleExtractDocumentFields)
	mux.HandleFunc("/map-template-fields", wh.HandleMapTemplateFields)
	mux.HandleFunc("/populate-template", wh.HandlePopulateTemplate)

	// NEW ENHANCED N8N WEBHOOK ENDPOINTS FOR TASK 1.2.2
	mux.HandleFunc("/webhook/n8n/enhanced/extract-document-fields", wh.HandleEnhancedExtractDocumentFields)
	mux.HandleFunc("/webhook/n8n/enhanced/map-fields-to-template", wh.HandleEnhancedMapFieldsToTemplate)
	mux.HandleFunc("/webhook/n8n/enhanced/format-field-value", wh.HandleEnhancedFormatFieldValue)
	mux.HandleFunc("/webhook/n8n/enhanced/validate-template-data", wh.HandleEnhancedValidateTemplateData)
	mux.HandleFunc("/webhook/n8n/enhanced/analyze-document", wh.HandleEnhancedAnalyzeDocument)

	// ENHANCED ENTITY EXTRACTION WEBHOOK ENDPOINTS FOR TASK 1.3
	mux.HandleFunc("/webhook/entity-extraction/company-and-deal-names", wh.handleExtractCompanyAndDealNames)
	mux.HandleFunc("/webhook/entity-extraction/financial-metrics", wh.handleExtractFinancialMetrics)
	mux.HandleFunc("/webhook/entity-extraction/personnel-and-roles", wh.handleExtractPersonnelAndRoles)
	mux.HandleFunc("/webhook/entity-extraction/validate-entities-across-documents", wh.handleValidateEntitiesAcrossDocuments)
}

// CreateHTTPServer creates an HTTP server with webhook handlers
func (wh *WebhookHandlers) CreateHTTPServer(port int) *http.Server {
	mux := http.NewServeMux()
	wh.RegisterHandlers(mux)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// WebhookServerConfig holds configuration for the webhook HTTP server
type WebhookServerConfig struct {
	Port        int    `json:"port"`
	EnableHTTPS bool   `json:"enable_https"`
	CertFile    string `json:"cert_file,omitempty"`
	KeyFile     string `json:"key_file,omitempty"`
	AutoStart   bool   `json:"auto_start"`
}

// StartWebhookServer starts the webhook HTTP server
func (wh *WebhookHandlers) StartWebhookServer(config *WebhookServerConfig) error {
	if config == nil {
		config = &WebhookServerConfig{
			Port:      8080,
			AutoStart: true,
		}
	}

	server := wh.CreateHTTPServer(config.Port)

	log.Printf("Starting webhook server on port %d", config.Port)

	// Start the result processor
	wh.StartListening()

	// Start server in a goroutine
	go func() {
		var err error
		if config.EnableHTTPS && config.CertFile != "" && config.KeyFile != "" {
			err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	return nil
}

// HandleDiscoverTemplates handles template discovery requests from n8n
func (wh *WebhookHandlers) HandleDiscoverTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		DocumentType   string                 `json:"documentType"`
		DealName       string                 `json:"dealName"`
		DocumentPath   string                 `json:"documentPath"`
		JobID          string                 `json:"jobId"`
		Classification map[string]interface{} `json:"classification"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call the template discovery method
	result, err := wh.app.DiscoverTemplatesForN8n(request.DocumentType, request.DealName, request.DocumentPath, request.Classification)
	if err != nil {
		log.Printf("Template discovery error: %v", err)
		http.Error(w, fmt.Sprintf("Template discovery failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleExtractDocumentFields handles document field extraction requests from n8n
func (wh *WebhookHandlers) HandleExtractDocumentFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		MappingParams map[string]interface{} `json:"mappingParams"`
		DealName      string                 `json:"dealName"`
		JobID         string                 `json:"jobId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call the field extraction method
	result, err := wh.app.ExtractDocumentFields(request.MappingParams, request.DealName)
	if err != nil {
		log.Printf("Field extraction error: %v", err)
		http.Error(w, fmt.Sprintf("Field extraction failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleMapTemplateFields handles template field mapping requests from n8n
func (wh *WebhookHandlers) HandleMapTemplateFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		MappingParams   map[string]interface{} `json:"mappingParams"`
		ExtractedFields map[string]interface{} `json:"extractedFields"`
		JobID           string                 `json:"jobId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call the field mapping method
	result, err := wh.app.MapTemplateFields(request.MappingParams, request.ExtractedFields)
	if err != nil {
		log.Printf("Field mapping error: %v", err)
		http.Error(w, fmt.Sprintf("Field mapping failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandlePopulateTemplate handles template population requests from n8n
func (wh *WebhookHandlers) HandlePopulateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TemplateID       string                   `json:"templateId"`
		FieldMappings    []map[string]interface{} `json:"fieldMappings"`
		PreserveFormulas bool                     `json:"preserveFormulas"`
		DealName         string                   `json:"dealName"`
		JobID            string                   `json:"jobId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract deal name from context if not provided
	dealName := request.DealName
	if dealName == "" {
		// Try to extract from job context or use a default
		dealName = "unknown"
	}

	// Call the template population method
	result, err := wh.app.PopulateTemplateWithData(request.TemplateID, request.FieldMappings, request.PreserveFormulas, dealName)
	if err != nil {
		log.Printf("Template population error: %v", err)
		http.Error(w, fmt.Sprintf("Template population failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// NEW ENHANCED WEBHOOK HANDLERS FOR TASK 1.2.2

// HandleEnhancedExtractDocumentFields handles enhanced document field extraction using AI
func (wh *WebhookHandlers) HandleEnhancedExtractDocumentFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Content         string                 `json:"content"`
		DocumentType    string                 `json:"documentType"`
		TemplateContext map[string]interface{} `json:"templateContext"`
		JobID           string                 `json:"jobId"`
		DealName        string                 `json:"dealName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use AI service to extract document fields
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := wh.app.aiService.ExtractDocumentFields(ctx, request.Content, request.DocumentType, request.TemplateContext)
	if err != nil {
		log.Printf("Enhanced field extraction error: %v", err)
		http.Error(w, fmt.Sprintf("Enhanced field extraction failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleEnhancedMapFieldsToTemplate handles enhanced field mapping using AI
func (wh *WebhookHandlers) HandleEnhancedMapFieldsToTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ExtractedFields map[string]interface{} `json:"extractedFields"`
		TemplateFields  []TemplateField        `json:"templateFields"`
		MappingContext  map[string]interface{} `json:"mappingContext"`
		JobID           string                 `json:"jobId"`
		DealName        string                 `json:"dealName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use AI service to map fields to template
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := wh.app.aiService.MapFieldsToTemplate(ctx, request.ExtractedFields, request.TemplateFields, request.MappingContext)
	if err != nil {
		log.Printf("Enhanced field mapping error: %v", err)
		http.Error(w, fmt.Sprintf("Enhanced field mapping failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleEnhancedFormatFieldValue handles enhanced field value formatting using AI
func (wh *WebhookHandlers) HandleEnhancedFormatFieldValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		RawValue           interface{}            `json:"rawValue"`
		FieldType          string                 `json:"fieldType"`
		FormatRequirements map[string]interface{} `json:"formatRequirements"`
		JobID              string                 `json:"jobId"`
		DealName           string                 `json:"dealName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use AI service to format field value
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := wh.app.aiService.FormatFieldValue(ctx, request.RawValue, request.FieldType, request.FormatRequirements)
	if err != nil {
		log.Printf("Enhanced field formatting error: %v", err)
		http.Error(w, fmt.Sprintf("Enhanced field formatting failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleEnhancedValidateTemplateData handles enhanced template data validation using AI
func (wh *WebhookHandlers) HandleEnhancedValidateTemplateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TemplateData    map[string]interface{} `json:"templateData"`
		ValidationRules []ValidationRule       `json:"validationRules"`
		JobID           string                 `json:"jobId"`
		DealName        string                 `json:"dealName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use AI service to validate template data
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := wh.app.aiService.ValidateTemplateData(ctx, request.TemplateData, request.ValidationRules)
	if err != nil {
		log.Printf("Enhanced template validation error: %v", err)
		http.Error(w, fmt.Sprintf("Enhanced template validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleEnhancedAnalyzeDocument handles comprehensive document analysis using AI
func (wh *WebhookHandlers) HandleEnhancedAnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Content      string                 `json:"content"`
		DocumentType string                 `json:"documentType"`
		AnalysisType string                 `json:"analysisType"` // "classification", "financial", "risks", "insights"
		Context      map[string]interface{} `json:"context"`
		JobID        string                 `json:"jobId"`
		DealName     string                 `json:"dealName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var result interface{}
	var err error

	// Route to appropriate AI analysis method based on analysis type
	switch request.AnalysisType {
	case "classification":
		result, err = wh.app.aiService.ClassifyDocument(ctx, request.Content, request.Context)
	case "financial":
		result, err = wh.app.aiService.ExtractFinancialData(ctx, request.Content)
	case "risks":
		result, err = wh.app.aiService.AnalyzeRisks(ctx, request.Content, request.DocumentType)
	case "insights":
		result, err = wh.app.aiService.GenerateInsights(ctx, request.Content, request.DocumentType)
	case "entities":
		result, err = wh.app.aiService.ExtractEntities(ctx, request.Content)
	default:
		// Default to comprehensive classification
		result, err = wh.app.aiService.ClassifyDocument(ctx, request.Content, request.Context)
	}

	if err != nil {
		log.Printf("Enhanced document analysis error: %v", err)
		http.Error(w, fmt.Sprintf("Enhanced document analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ENHANCED ENTITY EXTRACTION WEBHOOK ENDPOINTS FOR TASK 1.3

// handleExtractCompanyAndDealNames handles company and deal name extraction requests
func (wh *WebhookHandlers) handleExtractCompanyAndDealNames(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var request struct {
		Content      string `json:"content"`
		DocumentType string `json:"documentType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Use AI service to extract company and deal names
	result, err := wh.app.aiService.ExtractCompanyAndDealNames(ctx, request.Content, request.DocumentType)
	if err != nil {
		log.Printf("Error extracting company and deal names: %v", err)
		http.Error(w, "Failed to extract company and deal names", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// handleExtractFinancialMetrics handles financial metrics extraction requests
func (wh *WebhookHandlers) handleExtractFinancialMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var request struct {
		Content      string `json:"content"`
		DocumentType string `json:"documentType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Use AI service to extract financial metrics
	result, err := wh.app.aiService.ExtractFinancialMetrics(ctx, request.Content, request.DocumentType)
	if err != nil {
		log.Printf("Error extracting financial metrics: %v", err)
		http.Error(w, "Failed to extract financial metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// handleExtractPersonnelAndRoles handles personnel and roles extraction requests
func (wh *WebhookHandlers) handleExtractPersonnelAndRoles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var request struct {
		Content      string `json:"content"`
		DocumentType string `json:"documentType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Use AI service to extract personnel and roles
	result, err := wh.app.aiService.ExtractPersonnelAndRoles(ctx, request.Content, request.DocumentType)
	if err != nil {
		log.Printf("Error extracting personnel and roles: %v", err)
		http.Error(w, "Failed to extract personnel and roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// handleValidateEntitiesAcrossDocuments handles cross-document entity validation requests
func (wh *WebhookHandlers) handleValidateEntitiesAcrossDocuments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var request struct {
		DocumentExtractions []DocumentEntityExtraction `json:"documentExtractions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(request.DocumentExtractions) == 0 {
		http.Error(w, "Document extractions are required", http.StatusBadRequest)
		return
	}

	// Use AI service to validate entities across documents
	result, err := wh.app.aiService.ValidateEntitiesAcrossDocuments(ctx, request.DocumentExtractions)
	if err != nil {
		log.Printf("Error validating entities across documents: %v", err)
		http.Error(w, "Failed to validate entities across documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
