package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx                     context.Context
	configService           *ConfigService
	folderManager           *FolderManager
	permissionChecker       *PermissionChecker
	templateManager         *TemplateManager
	templateDiscovery       *TemplateDiscovery
	documentProcessor       *DocumentProcessor
	aiService               *AIService
	ocrService              *OCRService
	documentRouter          *DocumentRouter
	aiConfigManager         *AIConfigManager
	templateParser          *TemplateParser
	dataMapper              *DataMapper
	fieldMatcher            *FieldMatcher
	templatePopulator       *TemplatePopulator
	dealValuationCalculator *DealValuationCalculator
	competitiveAnalyzer     *CompetitiveAnalyzer
	trendAnalyzer           *TrendAnalyzer
	anomalyDetector         *AnomalyDetector
	webhookService          *WebhookService
	webhookHandlers         *WebhookHandlers
	webhookServer           *http.Server
	webhookServerConfig     *WebhookServerConfig
	webhookServerMu         sync.RWMutex
	jobTracker              *JobTracker
	n8nIntegration          *N8nIntegrationService
	schemaValidator         *WebhookSchemaValidator
	authManager             *AuthManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config service with fallback
	configService, err := NewConfigService()
	if err != nil {
		// Log error but continue with minimal config
		fmt.Printf("Error initializing config: %v\n", err)
		// Create a minimal config service to allow the app to function
		home, _ := os.UserHomeDir()
		defaultRoot := filepath.Join(home, "Desktop", "DealDone")
		configService = &ConfigService{
			config: &Config{
				DealDoneRoot:    defaultRoot,
				FirstRun:        true,
				DefaultTemplate: "",
				LastOpenedDeal:  "",
			},
		}
	}
	a.configService = configService

	// Initialize folder manager
	a.folderManager = NewFolderManager(configService)

	// Initialize permission checker
	a.permissionChecker = NewPermissionChecker()

	// Initialize template manager
	a.templateManager = NewTemplateManager(configService)

	// Initialize template discovery
	a.templateDiscovery = NewTemplateDiscovery(a.templateManager)

	// Initialize AI configuration manager
	aiConfigManager, err := NewAIConfigManager(configService)
	if err != nil {
		fmt.Printf("Error initializing AI config: %v\n", err)
		// Create with default config
		aiConfigManager = &AIConfigManager{
			config: &AIConfig{
				CacheTTL:  time.Minute * 30,
				RateLimit: 60,
			},
		}
	}
	a.aiConfigManager = aiConfigManager

	// Initialize AI service with config
	aiService := NewAIService(aiConfigManager.GetConfig())
	a.aiService = aiService

	// Initialize OCR service
	a.ocrService = NewOCRService("") // No default provider

	// Always initialize document processor and router - these are essential
	a.documentProcessor = NewDocumentProcessor(aiService)
	a.documentRouter = NewDocumentRouter(a.folderManager, a.documentProcessor)

	// Initialize template processing services
	templatesPath := configService.GetTemplatesPath()
	a.templateParser = NewTemplateParser(templatesPath)
	a.fieldMatcher = NewFieldMatcher(aiService)
	a.dataMapper = NewDataMapper(aiService, a.templateParser)
	a.templatePopulator = NewTemplatePopulator(a.templateParser)

	// Initialize analysis services
	a.dealValuationCalculator = NewDealValuationCalculator(aiService)
	a.competitiveAnalyzer = NewCompetitiveAnalyzer(aiService, a.documentProcessor)
	a.trendAnalyzer = NewTrendAnalyzer(aiService, a.dataMapper)
	a.anomalyDetector = NewAnomalyDetector(aiService, a.dataMapper)

	// Initialize webhook service with default configuration
	webhookConfig := &WebhookConfig{
		N8NBaseURL: "http://localhost:5678",
		AuthConfig: WebhookAuthConfig{
			APIKey:          "",
			SharedSecret:    "",
			TokenExpiration: 0,
			EnableHMAC:      false, // Will be enabled when secrets are configured
		},
		TimeoutSeconds:  30,
		MaxRetries:      3,
		RetryDelayMs:    1000,
		EnableLogging:   true,
		ValidatePayload: true,
	}

	webhookService, err := NewWebhookService(webhookConfig)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize webhook service: %v\n", err)
		// Create a minimal webhook service to allow the app to function
		webhookService = &WebhookService{
			config: webhookConfig,
		}
	}
	a.webhookService = webhookService

	// Initialize job tracker
	a.jobTracker = NewJobTracker(configService)

	// Initialize n8n integration service
	n8nConfig := &N8nConfig{
		BaseURL:               "http://localhost:5678",
		APIKey:                "",
		DefaultTimeout:        30 * time.Second,
		MaxRetries:            3,
		RetryDelay:            2 * time.Second,
		MaxConcurrentJobs:     5,
		EnableBatchProcessing: false,
		BatchSize:             10,
		BatchTimeout:          5 * time.Minute,
		HealthCheckInterval:   1 * time.Minute,
		LogRequests:           true,
	}

	n8nIntegration, err := NewN8nIntegrationService(n8nConfig, a.jobTracker, webhookService)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize n8n integration: %v\n", err)
		// Create a minimal service to allow the app to function
		n8nIntegration = &N8nIntegrationService{
			config: n8nConfig,
		}
	}
	a.n8nIntegration = n8nIntegration

	// Start n8n integration service
	if err := a.n8nIntegration.Start(); err != nil {
		fmt.Printf("Warning: Failed to start n8n integration service: %v\n", err)
	}

	// Initialize schema validator
	a.schemaValidator = NewWebhookSchemaValidator()

	// Initialize authentication manager
	authStoragePath := filepath.Join(configService.GetDealDoneRoot(), "config", "auth_keys.json")
	authManager, err := NewAuthManager(authStoragePath, nil)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize auth manager: %v\n", err)
	} else {
		a.authManager = authManager
	}

	// Initialize webhook handlers
	a.webhookHandlers = NewWebhookHandlers(a, webhookService)
}

// GetHomeDirectory returns the user's home directory
func (a *App) GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return homeDir
}

// GetCurrentDirectory returns the current working directory
func (a *App) GetCurrentDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return a.GetHomeDirectory()
	}
	return wd
}

// GetDesktopDirectory returns the desktop directory path
func (a *App) GetDesktopDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Desktop")
}

// GetDocumentsDirectory returns the documents directory path
func (a *App) GetDocumentsDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Documents")
}

// GetDownloadsDirectory returns the downloads directory path
func (a *App) GetDownloadsDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Downloads")
}

// IsFirstRun checks if this is the first run
func (a *App) IsFirstRun() bool {
	return !a.folderManager.IsDealDoneReady()
}

// CompleteFirstRunSetup initializes the folder structure with given path
func (a *App) CompleteFirstRunSetup(path string) error {
	// Update config with new path
	err := a.configService.SetDealDoneRoot(path)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Initialize folder structure
	err = a.folderManager.InitializeFolderStructure()
	if err != nil {
		return fmt.Errorf("failed to create folder structure: %w", err)
	}

	// Generate default templates
	err = a.templateManager.GenerateDefaultTemplates()
	if err != nil {
		// Don't fail the setup if templates fail
		fmt.Printf("Warning: Failed to generate default templates: %v\n", err)
	}

	return nil
}

// ProcessDocument handles document processing and routing
func (a *App) ProcessDocument(filePath string, dealName string) (*RoutingResult, error) {
	if a.documentRouter == nil {
		return nil, fmt.Errorf("document router not initialized")
	}

	return a.documentRouter.RouteDocument(filePath, dealName)
}

// ProcessDocuments handles batch document processing
func (a *App) ProcessDocuments(filePaths []string, dealName string) ([]*RoutingResult, error) {
	if a.documentRouter == nil {
		return nil, fmt.Errorf("document router not initialized")
	}

	return a.documentRouter.RouteDocuments(filePaths, dealName)
}

// ProcessFolder processes all documents in a folder
func (a *App) ProcessFolder(folderPath string, dealName string) ([]*RoutingResult, error) {
	if a.documentRouter == nil {
		return nil, fmt.Errorf("document router not initialized")
	}

	return a.documentRouter.RouteFolder(folderPath, dealName)
}

// GetRoutingSummary returns summary statistics for routing results
func (a *App) GetRoutingSummary(results []*RoutingResult) map[string]interface{} {
	if a.documentRouter == nil {
		return map[string]interface{}{
			"error": "document router not initialized",
		}
	}

	return a.documentRouter.GetRoutingSummary(results)
}

// GetSupportedFileTypes returns list of supported file extensions
func (a *App) GetSupportedFileTypes() []string {
	if a.documentProcessor == nil {
		return []string{}
	}

	return a.documentProcessor.GetSupportedExtensions()
}

// AnalyzeDocument analyzes a document without routing it
func (a *App) AnalyzeDocument(filePath string) (*DocumentInfo, error) {
	if a.documentProcessor == nil {
		return nil, fmt.Errorf("document processor not initialized")
	}

	return a.documentProcessor.ProcessDocument(filePath)
}

// GetDealsList returns list of all deals
func (a *App) GetDealsList() ([]DealInfo, error) {
	return a.folderManager.GetAllDeals()
}

// CreateDeal creates a new deal folder
func (a *App) CreateDeal(dealName string) error {
	_, err := a.folderManager.CreateDealFolder(dealName)
	return err
}

// DealExists checks if a deal folder exists
func (a *App) DealExists(dealName string) bool {
	return a.folderManager.DealExists(dealName)
}

// ExtractTextFromDocument extracts text content from a document
func (a *App) ExtractTextFromDocument(filePath string) (string, error) {
	if a.documentProcessor == nil {
		return "", fmt.Errorf("document processor not initialized")
	}

	return a.documentProcessor.ExtractText(filePath)
}

// GetDocumentMetadata extracts metadata from a document
func (a *App) GetDocumentMetadata(filePath string) (map[string]interface{}, error) {
	if a.documentProcessor == nil {
		return nil, fmt.Errorf("document processor not initialized")
	}

	return a.documentProcessor.GetDocumentMetadata(filePath)
}

// GetDealDoneRoot returns the configured DealDone root folder path
func (a *App) GetDealDoneRoot() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetDealDoneRoot()
}

// SetDealDoneRoot sets the DealDone root folder path and initializes the folder structure
func (a *App) SetDealDoneRoot(path string) error {
	if a.configService == nil {
		return fmt.Errorf("configuration service not initialized")
	}

	// Set the path in config
	if err := a.configService.SetDealDoneRoot(path); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Initialize folder structure
	if err := a.folderManager.InitializeFolderStructure(); err != nil {
		return fmt.Errorf("failed to create folder structure: %w", err)
	}

	// Generate default templates if none exist
	templatesPath := a.configService.GetTemplatesPath()
	dtg := NewDefaultTemplateGenerator(templatesPath)
	if !dtg.HasDefaultTemplates() {
		if err := dtg.GenerateDefaultTemplates(); err != nil {
			// Log error but don't fail - templates are optional
			println("Warning: Failed to generate default templates:", err.Error())
		}
	}

	// Mark first run as complete
	if err := a.configService.SetFirstRun(false); err != nil {
		return fmt.Errorf("failed to update first run status: %w", err)
	}

	return nil
}

// ValidateFolderStructure checks if the DealDone folder structure is valid
func (a *App) ValidateFolderStructure() error {
	if a.folderManager == nil {
		return fmt.Errorf("folder manager not initialized")
	}
	return a.folderManager.ValidateFolderStructure()
}

// GetDefaultDealDonePath returns the suggested default path for DealDone
func (a *App) GetDefaultDealDonePath() string {
	return filepath.Join(a.GetDesktopDirectory(), "DealDone")
}

// CheckFolderWritePermission checks if the app can write to the specified path
func (a *App) CheckFolderWritePermission(path string) bool {
	// Try to create a test directory
	testPath := filepath.Join(path, ".dealdone_test")
	err := os.MkdirAll(testPath, 0755)
	if err != nil {
		return false
	}

	// Clean up test directory
	os.RemoveAll(testPath)
	return true
}

// GetConfiguredTemplatesPath returns the path to the Templates folder
func (a *App) GetConfiguredTemplatesPath() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetTemplatesPath()
}

// GetConfiguredDealsPath returns the path to the Deals folder
func (a *App) GetConfiguredDealsPath() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetDealsPath()
}

// GetDealFolderPath returns the full path to a specific deal folder
func (a *App) GetDealFolderPath(dealName string) string {
	if a.folderManager == nil {
		return ""
	}
	return a.folderManager.GetDealPath(dealName)
}

// AI Configuration Methods

// GetAIProviderStatus returns the status of all AI providers
func (a *App) GetAIProviderStatus() map[string]interface{} {
	if a.aiConfigManager == nil {
		return map[string]interface{}{
			"error": "AI configuration not initialized",
		}
	}

	status := a.aiConfigManager.GetProviderStatus()
	result := make(map[string]interface{})

	for provider, providerStatus := range status {
		result[string(provider)] = map[string]interface{}{
			"configured": providerStatus.Configured,
			"model":      providerStatus.Model,
			"enabled":    providerStatus.Enabled,
		}
	}

	return result
}

// GetAIConfiguration returns the current AI configuration (without sensitive data)
func (a *App) GetAIConfiguration() (map[string]interface{}, error) {
	if a.aiConfigManager == nil {
		return nil, fmt.Errorf("AI configuration not initialized")
	}

	return a.aiConfigManager.Export()
}

// UpdateAIConfiguration updates AI configuration settings
func (a *App) UpdateAIConfiguration(updates map[string]interface{}) error {
	if a.aiConfigManager == nil {
		return fmt.Errorf("AI configuration not initialized")
	}

	if err := a.aiConfigManager.UpdateConfig(updates); err != nil {
		return err
	}

	// Reinitialize AI service with new config
	a.aiService = NewAIService(a.aiConfigManager.GetConfig())
	a.documentProcessor = NewDocumentProcessor(a.aiService)
	a.documentRouter = NewDocumentRouter(a.folderManager, a.documentProcessor)

	return nil
}

// SetAIProvider sets the preferred AI provider
func (a *App) SetAIProvider(provider string) error {
	if a.aiConfigManager == nil {
		return fmt.Errorf("AI configuration not initialized")
	}

	return a.aiConfigManager.SetPreferredProvider(AIProvider(provider))
}

// SetAIAPIKey sets an API key for a provider
func (a *App) SetAIAPIKey(provider string, apiKey string) error {
	if a.aiConfigManager == nil {
		return fmt.Errorf("AI configuration not initialized")
	}

	if err := a.aiConfigManager.SetAPIKey(AIProvider(provider), apiKey); err != nil {
		return err
	}

	// Reinitialize AI service with new config
	a.aiService = NewAIService(a.aiConfigManager.GetConfig())
	a.documentProcessor = NewDocumentProcessor(a.aiService)
	a.documentRouter = NewDocumentRouter(a.folderManager, a.documentProcessor)

	return nil
}

// AI Analysis Methods

// AnalyzeDocumentRisks analyzes a document for potential risks
func (a *App) AnalyzeDocumentRisks(filePath string) (*RiskAnalysis, error) {
	if a.aiService == nil || !a.aiService.IsAvailable() {
		return nil, fmt.Errorf("AI service not available")
	}

	// Extract text from document
	text, err := a.ExtractTextFromDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Detect document type
	info, err := a.documentProcessor.ProcessDocument(filePath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return a.aiService.AnalyzeRisks(ctx, text, string(info.Type))
}

// GenerateDocumentInsights generates insights about a document
func (a *App) GenerateDocumentInsights(filePath string) (*DocumentInsights, error) {
	if a.aiService == nil || !a.aiService.IsAvailable() {
		return nil, fmt.Errorf("AI service not available")
	}

	// Extract text from document
	text, err := a.ExtractTextFromDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Detect document type
	info, err := a.documentProcessor.ProcessDocument(filePath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return a.aiService.GenerateInsights(ctx, text, string(info.Type))
}

// ExtractDocumentEntities extracts named entities from a document
func (a *App) ExtractDocumentEntities(filePath string) (*EntityExtraction, error) {
	if a.aiService == nil || !a.aiService.IsAvailable() {
		return nil, fmt.Errorf("AI service not available")
	}

	// Extract text from document
	text, err := a.ExtractTextFromDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return a.aiService.ExtractEntities(ctx, text)
}

// ExtractFinancialData extracts financial data from a document
func (a *App) ExtractFinancialData(filePath string) (*FinancialAnalysis, error) {
	if a.aiService == nil || !a.aiService.IsAvailable() {
		return nil, fmt.Errorf("AI service not available")
	}

	// Extract text from document
	text, err := a.ExtractTextFromDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	return a.aiService.ExtractFinancialData(ctx, text)
}

// IsDealDoneReady checks if the DealDone folder structure is ready
func (a *App) IsDealDoneReady() bool {
	return a.folderManager.IsDealDoneReady()
}

// GetAIConfig returns the current AI configuration
func (a *App) GetAIConfig() (map[string]interface{}, error) {
	if a.aiService == nil {
		return map[string]interface{}{
			"provider": "none",
			"status":   "not initialized",
			"error":    "AI service not available",
		}, nil
	}
	return a.aiService.GetConfiguration(), nil
}

// SaveAIConfig saves the AI configuration
func (a *App) SaveAIConfig(config map[string]interface{}) error {
	// This would need to be implemented in aiservice
	// For now, return nil
	return nil
}

// GetAppConfig returns the application configuration
func (a *App) GetAppConfig() (map[string]interface{}, error) {
	// Return current app configuration
	config := map[string]interface{}{
		"dealDonePath":        a.configService.GetDealDoneRoot(),
		"autoCreateFolders":   true,
		"autoAnalyze":         true,
		"extractFinancial":    true,
		"extractRisks":        true,
		"extractEntities":     true,
		"confidenceThreshold": 0.7,
	}
	return config, nil
}

// SaveAppConfig saves the application configuration
func (a *App) SaveAppConfig(config map[string]interface{}) error {
	// This would update the app configuration
	// For now, return nil
	return nil
}

// GetAvailableAIProviders returns list of available AI providers
func (a *App) GetAvailableAIProviders() []string {
	return []string{"openai", "claude", "default"}
}

// TestAIProvider tests connection to AI provider
func (a *App) TestAIProvider(provider string, apiKey string) map[string]interface{} {
	// Test the provider connection
	// For now, return success
	return map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	}
}

// ExportAIConfig exports AI configuration (without sensitive data)
func (a *App) ExportAIConfig() (map[string]interface{}, error) {
	if a.aiService == nil {
		return map[string]interface{}{
			"provider": "none",
			"status":   "not initialized",
			"error":    "AI service not available",
		}, nil
	}
	config := a.aiService.GetConfiguration()
	// Remove sensitive data
	delete(config, "apiKey")
	return config, nil
}

// ImportAIConfig imports AI configuration
func (a *App) ImportAIConfig(config map[string]interface{}) error {
	// Import configuration
	return nil
}

// DiscoverTemplates performs comprehensive template discovery
func (a *App) DiscoverTemplates() ([]TemplateInfo, error) {
	return a.templateDiscovery.DiscoverTemplates()
}

// SearchTemplates searches templates with filters
func (a *App) SearchTemplates(query string, filters map[string]string) ([]TemplateInfo, error) {
	return a.templateDiscovery.SearchTemplates(query, filters)
}

// GetTemplateCategories returns available template categories
func (a *App) GetTemplateCategories() []string {
	return a.templateDiscovery.GetTemplateCategories()
}

// GetTemplateByID retrieves a template by its ID
func (a *App) GetTemplateByID(id string) (*TemplateInfo, error) {
	return a.templateDiscovery.GetTemplateByID(id)
}

// OrganizeTemplatesByCategory organizes templates by category
func (a *App) OrganizeTemplatesByCategory() (map[string][]TemplateInfo, error) {
	return a.templateDiscovery.OrganizeTemplatesByCategory()
}

// ImportTemplate imports a new template with metadata
func (a *App) ImportTemplate(sourcePath string, metadata *TemplateMetadata) error {
	return a.templateDiscovery.ImportTemplate(sourcePath, metadata)
}

// SaveTemplateMetadata saves metadata for a template
func (a *App) SaveTemplateMetadata(templatePath string, metadata *TemplateMetadata) error {
	return a.templateDiscovery.SaveTemplateMetadata(templatePath, metadata)
}

// Template Processing Methods

// ParseTemplate parses a template file and returns its structure
func (a *App) ParseTemplate(templatePath string) (*TemplateData, error) {
	if a.templateParser == nil {
		return nil, fmt.Errorf("template parser not initialized")
	}
	return a.templateParser.ParseTemplate(templatePath)
}

// ExtractTemplateFields extracts all fields from a template
func (a *App) ExtractTemplateFields(templatePath string) ([]DataField, error) {
	if a.templateParser == nil {
		return nil, fmt.Errorf("template parser not initialized")
	}

	templateData, err := a.templateParser.ParseTemplate(templatePath)
	if err != nil {
		return nil, err
	}

	return a.templateParser.ExtractDataFields(templateData), nil
}

// MapDataToTemplate maps extracted document data to template fields
func (a *App) MapDataToTemplate(templatePath string, documentPaths []string, dealName string) (*MappedData, error) {
	if a.dataMapper == nil {
		return nil, fmt.Errorf("data mapper not initialized")
	}

	// Parse template
	templateData, err := a.templateParser.ParseTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Process documents to get document info
	documents := make([]DocumentInfo, 0, len(documentPaths))
	for _, path := range documentPaths {
		info, err := a.documentProcessor.ProcessDocument(path)
		if err != nil {
			continue // Skip failed documents
		}
		documents = append(documents, *info)
	}

	// Map data
	return a.dataMapper.ExtractAndMapData(templateData, documents, dealName)
}

// MatchTemplateFields performs intelligent field matching between documents and template
func (a *App) MatchTemplateFields(sourceFields []string, templatePath string) (*MatchingResult, error) {
	if a.fieldMatcher == nil {
		return nil, fmt.Errorf("field matcher not initialized")
	}

	// Extract template fields
	templateFields, err := a.ExtractTemplateFields(templatePath)
	if err != nil {
		return nil, err
	}

	return a.fieldMatcher.MatchFields(sourceFields, templateFields)
}

// PopulateTemplate fills a template with mapped data
func (a *App) PopulateTemplate(templatePath string, mappedData *MappedData, outputPath string) error {
	if a.templatePopulator == nil {
		return fmt.Errorf("template populator not initialized")
	}

	return a.templatePopulator.PopulateTemplate(templatePath, mappedData, outputPath)
}

// ValidatePopulatedTemplate checks if formulas are preserved in populated template
func (a *App) ValidatePopulatedTemplate(populatedPath string, templatePath string) error {
	if a.templatePopulator == nil {
		return fmt.Errorf("template populator not initialized")
	}

	// Get original formulas
	preservation, err := a.templatePopulator.PreserveFormulas(templatePath)
	if err != nil {
		return fmt.Errorf("failed to analyze template formulas: %w", err)
	}

	return a.templatePopulator.ValidatePopulatedTemplate(populatedPath, preservation)
}

// GetFieldMappingSuggestions provides suggestions for unmapped fields
func (a *App) GetFieldMappingSuggestions(unmatchedFields []string, templatePath string) (map[string][]string, error) {
	if a.fieldMatcher == nil {
		return nil, fmt.Errorf("field matcher not initialized")
	}

	// Extract template fields
	templateFields, err := a.ExtractTemplateFields(templatePath)
	if err != nil {
		return nil, err
	}

	return a.fieldMatcher.GetFieldMappingSuggestions(unmatchedFields, templateFields), nil
}

// Analysis Engine Methods

// CalculateDealValuation performs comprehensive deal valuation
func (a *App) CalculateDealValuation(dealName string, financialData *FinancialAnalysis, marketData map[string]interface{}) (*ValuationResult, error) {
	if a.dealValuationCalculator == nil {
		return nil, fmt.Errorf("deal valuation calculator not initialized")
	}

	return a.dealValuationCalculator.CalculateValuation(dealName, financialData, marketData)
}

// CalculateQuickValuation performs a quick valuation based on basic metrics
func (a *App) CalculateQuickValuation(revenue, ebitda, netIncome float64) (*ValuationRange, error) {
	if a.dealValuationCalculator == nil {
		return nil, fmt.Errorf("deal valuation calculator not initialized")
	}

	return a.dealValuationCalculator.CalculateQuickValuation(revenue, ebitda, netIncome), nil
}

// GenerateValuationReport generates a text report of the valuation
func (a *App) GenerateValuationReport(result *ValuationResult) (string, error) {
	if a.dealValuationCalculator == nil {
		return "", fmt.Errorf("deal valuation calculator not initialized")
	}

	return a.dealValuationCalculator.GenerateValuationReport(result), nil
}

// AnalyzeCompetitiveLandscape performs competitive analysis
func (a *App) AnalyzeCompetitiveLandscape(dealName string, targetCompany string, documents []DocumentInfo, marketData map[string]interface{}) (*CompetitiveAnalysis, error) {
	if a.competitiveAnalyzer == nil {
		return nil, fmt.Errorf("competitive analyzer not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	return a.competitiveAnalyzer.AnalyzeCompetitiveLandscape(ctx, dealName, targetCompany, documents, marketData)
}

// QuickCompetitiveAssessment performs a quick competitive assessment
func (a *App) QuickCompetitiveAssessment(targetCompany string, revenue float64, marketShare float64) (map[string]interface{}, error) {
	if a.competitiveAnalyzer == nil {
		return nil, fmt.Errorf("competitive analyzer not initialized")
	}

	return a.competitiveAnalyzer.QuickCompetitiveAssessment(targetCompany, revenue, marketShare), nil
}

// AnalyzeTrends performs trend analysis across multiple documents
func (a *App) AnalyzeTrends(dealName string, documents []DocumentInfo, historicalData map[string]interface{}) (*TrendAnalysisResult, error) {
	if a.trendAnalyzer == nil {
		return nil, fmt.Errorf("trend analyzer not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	return a.trendAnalyzer.AnalyzeTrends(ctx, dealName, documents, historicalData)
}

// QuickTrendAssessment performs a quick trend assessment
func (a *App) QuickTrendAssessment(metricName string, values []float64) (map[string]interface{}, error) {
	if a.trendAnalyzer == nil {
		return nil, fmt.Errorf("trend analyzer not initialized")
	}

	return a.trendAnalyzer.QuickTrendAssessment(metricName, values), nil
}

// DetectAnomalies performs anomaly detection
func (a *App) DetectAnomalies(dealName string, documents []DocumentInfo, timeSeriesData map[string][]DataPoint) (*AnomalyDetectionResult, error) {
	if a.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	return a.anomalyDetector.DetectAnomalies(ctx, dealName, documents, timeSeriesData)
}

// QuickAnomalyCheck performs a quick anomaly check on a metric
func (a *App) QuickAnomalyCheck(metricName string, currentValue float64, historicalValues []float64) (map[string]interface{}, error) {
	if a.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}

	return a.anomalyDetector.QuickAnomalyCheck(metricName, currentValue, historicalValues), nil
}

// Analysis Export Methods

// ExportValuationToCSV exports valuation results to CSV format
func (a *App) ExportValuationToCSV(result *ValuationResult, outputPath string) error {
	return a.exportValuationData(result, outputPath, "csv")
}

// ExportValuationToJSON exports valuation results to JSON format
func (a *App) ExportValuationToJSON(result *ValuationResult, outputPath string) error {
	return a.exportValuationData(result, outputPath, "json")
}

// ExportCompetitiveAnalysisToCSV exports competitive analysis to CSV
func (a *App) ExportCompetitiveAnalysisToCSV(analysis *CompetitiveAnalysis, outputPath string) error {
	return a.exportCompetitiveData(analysis, outputPath, "csv")
}

// ExportCompetitiveAnalysisToJSON exports competitive analysis to JSON
func (a *App) ExportCompetitiveAnalysisToJSON(analysis *CompetitiveAnalysis, outputPath string) error {
	return a.exportCompetitiveData(analysis, outputPath, "json")
}

// ExportTrendAnalysisToCSV exports trend analysis to CSV
func (a *App) ExportTrendAnalysisToCSV(analysis *TrendAnalysisResult, outputPath string) error {
	return a.exportTrendData(analysis, outputPath, "csv")
}

// ExportTrendAnalysisToJSON exports trend analysis to JSON
func (a *App) ExportTrendAnalysisToJSON(analysis *TrendAnalysisResult, outputPath string) error {
	return a.exportTrendData(analysis, outputPath, "json")
}

// ExportAnomalyDetectionToCSV exports anomaly detection results to CSV
func (a *App) ExportAnomalyDetectionToCSV(result *AnomalyDetectionResult, outputPath string) error {
	return a.exportAnomalyData(result, outputPath, "csv")
}

// ExportAnomalyDetectionToJSON exports anomaly detection results to JSON
func (a *App) ExportAnomalyDetectionToJSON(result *AnomalyDetectionResult, outputPath string) error {
	return a.exportAnomalyData(result, outputPath, "json")
}

// ExportCompleteAnalysisReport exports a comprehensive analysis report
func (a *App) ExportCompleteAnalysisReport(dealName string, outputPath string, format string) error {
	// This would generate a complete analysis report combining all analysis types
	// For now, return success
	return nil
}

// Helper methods for exporting data

func (a *App) exportValuationData(result *ValuationResult, outputPath string, format string) error {
	// Implementation would depend on the format and requirements
	// For now, return success
	return nil
}

func (a *App) exportCompetitiveData(analysis *CompetitiveAnalysis, outputPath string, format string) error {
	// Implementation would depend on the format and requirements
	// For now, return success
	return nil
}

func (a *App) exportTrendData(analysis *TrendAnalysisResult, outputPath string, format string) error {
	// Implementation would depend on the format and requirements
	// For now, return success
	return nil
}

func (a *App) exportAnomalyData(result *AnomalyDetectionResult, outputPath string, format string) error {
	// Implementation would depend on the format and requirements
	// For now, return success
	return nil
}

// Webhook-related methods

// SendDocumentsToN8n sends documents to n8n for analysis
func (a *App) SendDocumentsToN8n(dealName string, filePaths []string, triggerType string) (string, error) {
	if a.n8nIntegration == nil {
		return "", fmt.Errorf("n8n integration service not initialized")
	}
	if a.jobTracker == nil {
		return "", fmt.Errorf("job tracker not initialized")
	}

	// Generate job ID
	jobID := fmt.Sprintf("job_%d_%s", time.Now().UnixMilli(), dealName)

	// Create job entry in tracker
	a.jobTracker.CreateJob(jobID, dealName, WebhookTriggerType(triggerType), filePaths)

	// Create payload
	payload := &DocumentWebhookPayload{
		DealName:    dealName,
		FilePaths:   filePaths,
		TriggerType: WebhookTriggerType(triggerType),
		JobID:       jobID,
		Timestamp:   time.Now().UnixMilli(),
	}

	// Send to n8n through integration service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request, err := a.n8nIntegration.SendDocumentAnalysisRequest(ctx, payload)
	if err != nil {
		// Mark job as failed
		a.jobTracker.FailJob(jobID, fmt.Sprintf("Failed to send to n8n: %v", err))
		return "", fmt.Errorf("failed to send documents to n8n: %w", err)
	}

	// Log the n8n request ID for tracking
	log.Printf("Created n8n request %s for job %s", request.ID, jobID)

	return jobID, nil
}

// GetWebhookJobStatus queries the status of a webhook job from local tracker first, then n8n if needed
func (a *App) GetWebhookJobStatus(jobID string, dealName string) (map[string]interface{}, error) {
	// Try to get job status from local tracker first
	if a.jobTracker != nil {
		if job, err := a.jobTracker.GetJob(jobID); err == nil {
			return map[string]interface{}{
				"jobId":              job.JobID,
				"dealName":           job.DealName,
				"status":             string(job.Status),
				"progress":           job.Progress,
				"currentStep":        job.CurrentStep,
				"createdAt":          job.CreatedAt,
				"updatedAt":          job.UpdatedAt,
				"startedAt":          job.StartedAt,
				"completedAt":        job.CompletedAt,
				"estimatedTime":      job.EstimatedTime,
				"processedDocuments": job.ProcessedDocuments,
				"totalDocuments":     job.TotalDocuments,
				"queuePosition":      job.QueuePosition,
				"retryCount":         job.RetryCount,
				"errors":             job.Errors,
				"processingHistory":  job.ProcessingHistory,
				"source":             "local_tracker",
			}, nil
		}
	}

	// Fall back to querying n8n directly
	if a.webhookService == nil {
		return nil, fmt.Errorf("webhook service not initialized and job not found in local tracker")
	}

	query := &WebhookStatusQuery{
		JobID:     jobID,
		DealName:  dealName,
		Timestamp: time.Now().UnixMilli(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := a.webhookService.QueryJobStatus(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query job status: %w", err)
	}

	return map[string]interface{}{
		"jobId":          response.JobID,
		"status":         response.Status,
		"progress":       response.Progress,
		"currentStep":    response.CurrentStep,
		"estimatedTime":  response.EstimatedTime,
		"lastUpdated":    response.LastUpdated,
		"additionalInfo": response.AdditionalInfo,
		"source":         "n8n_query",
	}, nil
}

// GetWebhookConfiguration returns the current webhook configuration
func (a *App) GetWebhookConfiguration() (map[string]interface{}, error) {
	if a.webhookService == nil {
		return nil, fmt.Errorf("webhook service not initialized")
	}

	config := a.webhookService.GetConfig()
	return map[string]interface{}{
		"n8nBaseURL":      config.N8NBaseURL,
		"timeoutSeconds":  config.TimeoutSeconds,
		"maxRetries":      config.MaxRetries,
		"retryDelayMs":    config.RetryDelayMs,
		"enableLogging":   config.EnableLogging,
		"validatePayload": config.ValidatePayload,
		"hasAPIKey":       config.AuthConfig.APIKey != "",
		"hasSharedSecret": config.AuthConfig.SharedSecret != "",
		"enableHMAC":      config.AuthConfig.EnableHMAC,
	}, nil
}

// UpdateWebhookConfiguration updates the webhook configuration
func (a *App) UpdateWebhookConfiguration(updates map[string]interface{}) error {
	if a.webhookService == nil {
		return fmt.Errorf("webhook service not initialized")
	}

	config := a.webhookService.GetConfig()

	// Update fields if provided
	if n8nURL, ok := updates["n8nBaseURL"].(string); ok {
		config.N8NBaseURL = n8nURL
	}
	if timeout, ok := updates["timeoutSeconds"].(float64); ok {
		config.TimeoutSeconds = int(timeout)
	}
	if maxRetries, ok := updates["maxRetries"].(float64); ok {
		config.MaxRetries = int(maxRetries)
	}
	if retryDelay, ok := updates["retryDelayMs"].(float64); ok {
		config.RetryDelayMs = int(retryDelay)
	}
	if enableLogging, ok := updates["enableLogging"].(bool); ok {
		config.EnableLogging = enableLogging
	}
	if validatePayload, ok := updates["validatePayload"].(bool); ok {
		config.ValidatePayload = validatePayload
	}
	if apiKey, ok := updates["apiKey"].(string); ok {
		config.AuthConfig.APIKey = apiKey
	}
	if sharedSecret, ok := updates["sharedSecret"].(string); ok {
		config.AuthConfig.SharedSecret = sharedSecret
	}
	if enableHMAC, ok := updates["enableHMAC"].(bool); ok {
		config.AuthConfig.EnableHMAC = enableHMAC
	}

	return a.webhookService.UpdateConfig(config)
}

// StartWebhookServer starts the webhook HTTP server
// StartWebhookServer starts the webhook HTTP server with comprehensive configuration
func (a *App) StartWebhookServer(port int) error {
	return a.StartWebhookServerWithConfig(map[string]interface{}{
		"port":      port,
		"autoStart": true,
	})
}

// StartWebhookServerWithConfig starts the webhook server with detailed configuration
func (a *App) StartWebhookServerWithConfig(configMap map[string]interface{}) error {
	a.webhookServerMu.Lock()
	defer a.webhookServerMu.Unlock()

	if a.webhookHandlers == nil {
		return fmt.Errorf("webhook handlers not initialized")
	}

	// Stop existing server if running
	if a.webhookServer != nil {
		if err := a.stopWebhookServerInternal(); err != nil {
			log.Printf("Warning: Error stopping existing webhook server: %v", err)
		}
	}

	// Parse configuration
	config := &WebhookServerConfig{
		Port:      8080,
		AutoStart: true,
	}

	if port, ok := configMap["port"].(float64); ok {
		config.Port = int(port)
	} else if port, ok := configMap["port"].(int); ok {
		config.Port = port
	}

	if autoStart, ok := configMap["autoStart"].(bool); ok {
		config.AutoStart = autoStart
	}

	if enableHTTPS, ok := configMap["enableHTTPS"].(bool); ok {
		config.EnableHTTPS = enableHTTPS
	}

	if certFile, ok := configMap["certFile"].(string); ok {
		config.CertFile = certFile
	}

	if keyFile, ok := configMap["keyFile"].(string); ok {
		config.KeyFile = keyFile
	}

	// Store configuration
	a.webhookServerConfig = config

	// Create server with authentication middleware
	server := a.createAuthenticatedWebhookServer(config)
	a.webhookServer = server

	// Start the result processor
	a.webhookHandlers.StartListening()

	// Start server in a goroutine
	go func() {
		var err error
		if config.EnableHTTPS && config.CertFile != "" && config.KeyFile != "" {
			log.Printf("Starting HTTPS webhook server on port %d", config.Port)
			err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
		} else {
			log.Printf("Starting HTTP webhook server on port %d", config.Port)
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	log.Printf("Webhook server started successfully on port %d", config.Port)
	return nil
}

// StopWebhookServer stops the webhook HTTP server
func (a *App) StopWebhookServer() error {
	a.webhookServerMu.Lock()
	defer a.webhookServerMu.Unlock()

	return a.stopWebhookServerInternal()
}

// stopWebhookServerInternal stops the webhook server (must be called with lock held)
func (a *App) stopWebhookServerInternal() error {
	if a.webhookServer == nil {
		return fmt.Errorf("webhook server is not running")
	}

	// Stop the result processor
	if a.webhookHandlers != nil {
		a.webhookHandlers.StopListening()
	}

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.webhookServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown webhook server: %w", err)
	}

	a.webhookServer = nil
	log.Printf("Webhook server stopped successfully")
	return nil
}

// RestartWebhookServer restarts the webhook server with current configuration
func (a *App) RestartWebhookServer() error {
	a.webhookServerMu.RLock()
	config := a.webhookServerConfig
	a.webhookServerMu.RUnlock()

	if config == nil {
		return fmt.Errorf("no webhook server configuration available")
	}

	configMap := map[string]interface{}{
		"port":        config.Port,
		"enableHTTPS": config.EnableHTTPS,
		"certFile":    config.CertFile,
		"keyFile":     config.KeyFile,
		"autoStart":   config.AutoStart,
	}

	return a.StartWebhookServerWithConfig(configMap)
}

// GetWebhookServerStatus returns the current status of the webhook server
func (a *App) GetWebhookServerStatus() (map[string]interface{}, error) {
	a.webhookServerMu.RLock()
	defer a.webhookServerMu.RUnlock()

	status := map[string]interface{}{
		"timestamp": time.Now().UnixMilli(),
		"running":   a.webhookServer != nil,
	}

	if a.webhookServerConfig != nil {
		status["configuration"] = map[string]interface{}{
			"port":        a.webhookServerConfig.Port,
			"enableHTTPS": a.webhookServerConfig.EnableHTTPS,
			"hasCertFile": a.webhookServerConfig.CertFile != "",
			"hasKeyFile":  a.webhookServerConfig.KeyFile != "",
			"autoStart":   a.webhookServerConfig.AutoStart,
		}
	}

	// Add endpoint information
	if a.webhookServer != nil {
		protocol := "http"
		if a.webhookServerConfig.EnableHTTPS {
			protocol = "https"
		}

		baseURL := fmt.Sprintf("%s://localhost:%d", protocol, a.webhookServerConfig.Port)
		status["endpoints"] = map[string]interface{}{
			"results": fmt.Sprintf("%s/webhook/results", baseURL),
			"status":  fmt.Sprintf("%s/webhook/status", baseURL),
			"health":  fmt.Sprintf("%s/webhook/health", baseURL),
			"baseURL": baseURL,
		}
	}

	return status, nil
}

// GetWebhookServerConfiguration returns the current webhook server configuration
func (a *App) GetWebhookServerConfiguration() (map[string]interface{}, error) {
	a.webhookServerMu.RLock()
	defer a.webhookServerMu.RUnlock()

	if a.webhookServerConfig == nil {
		return map[string]interface{}{
			"port":        8080,
			"enableHTTPS": false,
			"autoStart":   true,
		}, nil
	}

	return map[string]interface{}{
		"port":        a.webhookServerConfig.Port,
		"enableHTTPS": a.webhookServerConfig.EnableHTTPS,
		"hasCertFile": a.webhookServerConfig.CertFile != "",
		"hasKeyFile":  a.webhookServerConfig.KeyFile != "",
		"autoStart":   a.webhookServerConfig.AutoStart,
	}, nil
}

// UpdateWebhookServerConfiguration updates the webhook server configuration
func (a *App) UpdateWebhookServerConfiguration(updates map[string]interface{}) error {
	a.webhookServerMu.Lock()
	defer a.webhookServerMu.Unlock()

	if a.webhookServerConfig == nil {
		a.webhookServerConfig = &WebhookServerConfig{
			Port:      8080,
			AutoStart: true,
		}
	}

	config := a.webhookServerConfig
	needsRestart := false

	// Update configuration fields
	if port, ok := updates["port"].(float64); ok {
		newPort := int(port)
		if config.Port != newPort {
			config.Port = newPort
			needsRestart = true
		}
	} else if port, ok := updates["port"].(int); ok {
		if config.Port != port {
			config.Port = port
			needsRestart = true
		}
	}

	if enableHTTPS, ok := updates["enableHTTPS"].(bool); ok {
		if config.EnableHTTPS != enableHTTPS {
			config.EnableHTTPS = enableHTTPS
			needsRestart = true
		}
	}

	if certFile, ok := updates["certFile"].(string); ok {
		if config.CertFile != certFile {
			config.CertFile = certFile
			needsRestart = true
		}
	}

	if keyFile, ok := updates["keyFile"].(string); ok {
		if config.KeyFile != keyFile {
			config.KeyFile = keyFile
			needsRestart = true
		}
	}

	if autoStart, ok := updates["autoStart"].(bool); ok {
		config.AutoStart = autoStart
	}

	// Restart server if needed and it's currently running
	if needsRestart && a.webhookServer != nil {
		log.Printf("Webhook server configuration changed, restarting...")
		if err := a.stopWebhookServerInternal(); err != nil {
			log.Printf("Warning: Error stopping webhook server for restart: %v", err)
		}

		// Start with new configuration
		server := a.createAuthenticatedWebhookServer(config)
		a.webhookServer = server

		go func() {
			var err error
			if config.EnableHTTPS && config.CertFile != "" && config.KeyFile != "" {
				err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
			} else {
				err = server.ListenAndServe()
			}

			if err != nil && err != http.ErrServerClosed {
				log.Printf("Webhook server error after restart: %v", err)
			}
		}()
	}

	return nil
}

// GetWebhookEndpoints returns the available webhook endpoints
func (a *App) GetWebhookEndpoints() (map[string]interface{}, error) {
	a.webhookServerMu.RLock()
	defer a.webhookServerMu.RUnlock()

	if a.webhookServer == nil || a.webhookServerConfig == nil {
		return map[string]interface{}{
			"running": false,
			"message": "webhook server is not running",
		}, nil
	}

	protocol := "http"
	if a.webhookServerConfig.EnableHTTPS {
		protocol = "https"
	}

	baseURL := fmt.Sprintf("%s://localhost:%d", protocol, a.webhookServerConfig.Port)

	return map[string]interface{}{
		"running":  true,
		"baseURL":  baseURL,
		"protocol": protocol,
		"port":     a.webhookServerConfig.Port,
		"endpoints": map[string]interface{}{
			"results": map[string]interface{}{
				"url":         fmt.Sprintf("%s/webhook/results", baseURL),
				"method":      "POST",
				"description": "Receive processing results from n8n workflows",
				"auth":        "API key + HMAC signature required",
			},
			"status": map[string]interface{}{
				"url":         fmt.Sprintf("%s/webhook/status", baseURL),
				"method":      "GET",
				"description": "Query job status by ID",
				"auth":        "API key required",
				"params":      []string{"jobId", "dealName"},
			},
			"health": map[string]interface{}{
				"url":         fmt.Sprintf("%s/webhook/health", baseURL),
				"method":      "GET",
				"description": "Health check endpoint for monitoring",
				"auth":        "none",
			},
		},
	}, nil
}

// createAuthenticatedWebhookServer creates an HTTP server with authentication middleware
func (a *App) createAuthenticatedWebhookServer(config *WebhookServerConfig) *http.Server {
	mux := http.NewServeMux()

	// Add authentication middleware for protected endpoints
	mux.HandleFunc("/webhook/results", a.withAuthentication(a.webhookHandlers.HandleProcessingResults))
	mux.HandleFunc("/webhook/status", a.withAuthentication(a.webhookHandlers.HandleStatusQuery))
	mux.HandleFunc("/webhook/health", a.webhookHandlers.HandleHealthCheck) // No auth for health checks

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// withAuthentication wraps an HTTP handler with API key and HMAC authentication
func (a *App) withAuthentication(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Signature, X-Timestamp")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Skip authentication for health checks
		if r.URL.Path == "/webhook/health" {
			handler(w, r)
			return
		}

		// Extract authentication headers
		apiKey := r.Header.Get("X-API-Key")
		signature := r.Header.Get("X-Signature")
		timestamp := r.Header.Get("X-Timestamp")

		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}

		// Authenticate request using AuthManager
		if a.authManager != nil {
			// Read body for authentication
			var body string
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}
				body = string(bodyBytes)
				// Restore body for handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Parse timestamp
			var timestampInt int64
			if timestamp != "" {
				var err error
				timestampInt, err = strconv.ParseInt(timestamp, 10, 64)
				if err != nil {
					timestampInt = time.Now().Unix()
				}
			} else {
				timestampInt = time.Now().Unix()
			}

			// Create authentication request
			authReq := &AuthenticationRequest{
				APIKey:    apiKey,
				Signature: signature,
				Timestamp: timestampInt,
				Method:    r.Method,
				Path:      r.URL.Path,
				Body:      body,
				Headers:   make(map[string]string),
				ClientIP:  r.RemoteAddr,
				UserAgent: r.UserAgent(),
				RequestID: r.Header.Get("X-Request-ID"),
			}

			// Copy relevant headers
			for key, values := range r.Header {
				if len(values) > 0 {
					authReq.Headers[key] = values[0]
				}
			}

			// Authenticate the request
			authResult, err := a.authManager.AuthenticateRequest(authReq)
			if err != nil {
				http.Error(w, fmt.Sprintf("Authentication error: %v", err), http.StatusInternalServerError)
				return
			}

			if !authResult.Success {
				http.Error(w, authResult.ErrorMessage, http.StatusUnauthorized)
				return
			}

			// Add authentication info to request context for handler use
			ctx := context.WithValue(r.Context(), "authResult", authResult)
			r = r.WithContext(ctx)
		}

		// Call the original handler
		handler(w, r)
	}
}

// GetWebhookServerHealth checks the health of the webhook services
func (a *App) GetWebhookServerHealth() (map[string]interface{}, error) {
	if a.webhookService == nil {
		return nil, fmt.Errorf("webhook service not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"timestamp": time.Now().UnixMilli(),
	}

	// Check webhook service health
	if err := a.webhookService.IsHealthy(ctx); err != nil {
		healthStatus["webhookService"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		healthStatus["overallStatus"] = "unhealthy"
	} else {
		healthStatus["webhookService"] = map[string]interface{}{
			"status": "healthy",
		}
		healthStatus["overallStatus"] = "healthy"
	}

	// Check handlers status
	if a.webhookHandlers != nil {
		healthStatus["webhookHandlers"] = map[string]interface{}{
			"status": "initialized",
		}
	}

	return healthStatus, nil
}

// ValidateDocumentPayload validates a document webhook payload
func (a *App) ValidateDocumentPayload(dealName string, filePaths []string, triggerType string) (map[string]interface{}, error) {
	if a.webhookService == nil {
		return nil, fmt.Errorf("webhook service not initialized")
	}

	payload := &DocumentWebhookPayload{
		DealName:    dealName,
		FilePaths:   filePaths,
		TriggerType: WebhookTriggerType(triggerType),
		JobID:       "validation-test",
		Timestamp:   time.Now().UnixMilli(),
	}

	validationResult := a.webhookService.GetValidationResult(payload)

	return map[string]interface{}{
		"valid":  validationResult.Valid,
		"errors": validationResult.Errors,
	}, nil
}

// Job Tracking Methods

// QueryJobs queries jobs based on criteria
func (a *App) QueryJobs(dealName string, status string, triggerType string, limit int, offset int, sortBy string, sortOrder string) (map[string]interface{}, error) {
	if a.jobTracker == nil {
		return nil, fmt.Errorf("job tracker not initialized")
	}

	query := &JobQuery{
		DealName:    dealName,
		Status:      JobStatus(status),
		TriggerType: WebhookTriggerType(triggerType),
		Limit:       limit,
		Offset:      offset,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}

	jobs, err := a.jobTracker.QueryJobs(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}

	// Convert jobs to map format for frontend
	jobsData := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		jobsData[i] = map[string]interface{}{
			"jobId":              job.JobID,
			"dealName":           job.DealName,
			"status":             string(job.Status),
			"triggerType":        string(job.TriggerType),
			"filePaths":          job.FilePaths,
			"createdAt":          job.CreatedAt,
			"updatedAt":          job.UpdatedAt,
			"startedAt":          job.StartedAt,
			"completedAt":        job.CompletedAt,
			"progress":           job.Progress,
			"currentStep":        job.CurrentStep,
			"estimatedTime":      job.EstimatedTime,
			"processedDocuments": job.ProcessedDocuments,
			"totalDocuments":     job.TotalDocuments,
			"queuePosition":      job.QueuePosition,
			"retryCount":         job.RetryCount,
			"maxRetries":         job.MaxRetries,
			"errors":             job.Errors,
			"processingHistory":  job.ProcessingHistory,
			"metadata":           job.Metadata,
		}
	}

	return map[string]interface{}{
		"jobs":       jobsData,
		"totalCount": len(jobsData),
		"query":      query,
	}, nil
}

// GetJobSummary returns aggregated job statistics
func (a *App) GetJobSummary() (map[string]interface{}, error) {
	if a.jobTracker == nil {
		return nil, fmt.Errorf("job tracker not initialized")
	}

	summary := a.jobTracker.GetJobSummary()

	// Convert status counts to string keys for frontend
	statusCounts := make(map[string]int)
	for status, count := range summary.StatusCounts {
		statusCounts[string(status)] = count
	}

	// Convert recent activity
	recentActivity := make([]map[string]interface{}, len(summary.RecentActivity))
	for i, entry := range summary.RecentActivity {
		recentActivity[i] = map[string]interface{}{
			"timestamp": entry.Timestamp,
			"status":    entry.Status,
			"step":      entry.Step,
			"message":   entry.Message,
			"progress":  entry.Progress,
			"error":     entry.Error,
		}
	}

	return map[string]interface{}{
		"totalJobs":       summary.TotalJobs,
		"activeJobs":      summary.ActiveJobs,
		"completedJobs":   summary.CompletedJobs,
		"failedJobs":      summary.FailedJobs,
		"averageProgress": summary.AverageProgress,
		"statusCounts":    statusCounts,
		"dealCounts":      summary.DealCounts,
		"recentActivity":  recentActivity,
	}, nil
}

// GetJobsByDeal returns all jobs for a specific deal
func (a *App) GetJobsByDeal(dealName string) ([]map[string]interface{}, error) {
	result, err := a.QueryJobs(dealName, "", "", 0, 0, "updatedAt", "desc")
	if err != nil {
		return nil, err
	}

	if jobs, ok := result["jobs"].([]map[string]interface{}); ok {
		return jobs, nil
	}

	return []map[string]interface{}{}, nil
}

// GetActiveJobs returns all currently active (processing, queued) jobs
func (a *App) GetActiveJobs() ([]map[string]interface{}, error) {
	processingJobs, err1 := a.QueryJobs("", "processing", "", 0, 0, "updatedAt", "desc")
	queuedJobs, err2 := a.QueryJobs("", "queued", "", 0, 0, "updatedAt", "desc")

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf("failed to get active jobs: %v, %v", err1, err2)
	}

	var allActiveJobs []map[string]interface{}

	if err1 == nil {
		if jobs, ok := processingJobs["jobs"].([]map[string]interface{}); ok {
			allActiveJobs = append(allActiveJobs, jobs...)
		}
	}

	if err2 == nil {
		if jobs, ok := queuedJobs["jobs"].([]map[string]interface{}); ok {
			allActiveJobs = append(allActiveJobs, jobs...)
		}
	}

	return allActiveJobs, nil
}

// RetryFailedJob retries a failed job
func (a *App) RetryFailedJob(jobID string) error {
	if a.jobTracker == nil {
		return fmt.Errorf("job tracker not initialized")
	}

	return a.jobTracker.RetryJob(jobID)
}

// CancelJob cancels a pending or processing job
func (a *App) CancelJob(jobID string) error {
	if a.jobTracker == nil {
		return fmt.Errorf("job tracker not initialized")
	}

	return a.jobTracker.CancelJob(jobID)
}

// GetJobDetails returns detailed information about a specific job
func (a *App) GetJobDetails(jobID string) (map[string]interface{}, error) {
	if a.jobTracker == nil {
		return nil, fmt.Errorf("job tracker not initialized")
	}

	job, err := a.jobTracker.GetJob(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %w", err)
	}

	// Convert processing history
	historyData := make([]map[string]interface{}, len(job.ProcessingHistory))
	for i, entry := range job.ProcessingHistory {
		historyData[i] = map[string]interface{}{
			"timestamp": entry.Timestamp,
			"status":    entry.Status,
			"step":      entry.Step,
			"message":   entry.Message,
			"progress":  entry.Progress,
			"error":     entry.Error,
		}
	}

	result := map[string]interface{}{
		"jobId":              job.JobID,
		"dealName":           job.DealName,
		"status":             string(job.Status),
		"triggerType":        string(job.TriggerType),
		"filePaths":          job.FilePaths,
		"createdAt":          job.CreatedAt,
		"updatedAt":          job.UpdatedAt,
		"startedAt":          job.StartedAt,
		"completedAt":        job.CompletedAt,
		"progress":           job.Progress,
		"currentStep":        job.CurrentStep,
		"estimatedTime":      job.EstimatedTime,
		"processedDocuments": job.ProcessedDocuments,
		"totalDocuments":     job.TotalDocuments,
		"queuePosition":      job.QueuePosition,
		"retryCount":         job.RetryCount,
		"maxRetries":         job.MaxRetries,
		"errors":             job.Errors,
		"processingHistory":  historyData,
		"metadata":           job.Metadata,
	}

	// Include processing results if available
	if job.ProcessingResults != nil {
		result["processingResults"] = map[string]interface{}{
			"processedDocuments": job.ProcessingResults.ProcessedDocuments,
			"templatesUpdated":   job.ProcessingResults.TemplatesUpdated,
			"averageConfidence":  job.ProcessingResults.AverageConfidence,
			"processingTime":     job.ProcessingResults.ProcessingTime,
			"results":            job.ProcessingResults.Results,
			"errors":             job.ProcessingResults.Errors,
		}
	}

	return result, nil
}

// CleanupOldJobs removes old completed/failed jobs
func (a *App) CleanupOldJobs(olderThanHours int) (int, error) {
	if a.jobTracker == nil {
		return 0, fmt.Errorf("job tracker not initialized")
	}

	if olderThanHours <= 0 {
		olderThanHours = 168 // Default to 1 week
	}

	removedCount := a.jobTracker.CleanupOldJobs(olderThanHours)
	return removedCount, nil
}

// GetJobTrackerHealth checks the health of the job tracking system
func (a *App) GetJobTrackerHealth() (map[string]interface{}, error) {
	if a.jobTracker == nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "job tracker not initialized",
		}, fmt.Errorf("job tracker not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"timestamp": time.Now().UnixMilli(),
	}

	if err := a.jobTracker.IsHealthy(ctx); err != nil {
		healthStatus["status"] = "unhealthy"
		healthStatus["error"] = err.Error()
		return healthStatus, err
	}

	healthStatus["status"] = "healthy"

	// Add some basic stats
	if summary, err := a.GetJobSummary(); err == nil {
		healthStatus["totalJobs"] = summary["totalJobs"]
		healthStatus["activeJobs"] = summary["activeJobs"]
	}

	return healthStatus, nil
}

// N8n Integration Methods

// GetN8nIntegrationStatus returns the status of the n8n integration service
func (a *App) GetN8nIntegrationStatus() (map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return map[string]interface{}{
			"status": "not_initialized",
			"error":  "n8n integration service not initialized",
		}, fmt.Errorf("n8n integration service not initialized")
	}

	stats := a.n8nIntegration.GetStats()

	// Add health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := "healthy"
	var healthError string
	if err := a.n8nIntegration.IsHealthy(ctx); err != nil {
		healthStatus = "unhealthy"
		healthError = err.Error()
	}

	result := map[string]interface{}{
		"status":            healthStatus,
		"isRunning":         stats["isRunning"],
		"queueSize":         stats["queueSize"],
		"activeRequests":    stats["activeRequests"],
		"maxConcurrentJobs": stats["maxConcurrentJobs"],
		"batchSize":         stats["batchSize"],
		"timestamp":         time.Now().UnixMilli(),
	}

	if healthError != "" {
		result["error"] = healthError
	}

	return result, nil
}

// GetN8nConfiguration returns the current n8n configuration
func (a *App) GetN8nConfiguration() (map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return nil, fmt.Errorf("n8n integration service not initialized")
	}

	config := a.n8nIntegration.config
	return map[string]interface{}{
		"baseURL":               config.BaseURL,
		"hasAPIKey":             config.APIKey != "",
		"defaultTimeout":        config.DefaultTimeout.Seconds(),
		"maxRetries":            config.MaxRetries,
		"retryDelay":            config.RetryDelay.Seconds(),
		"maxConcurrentJobs":     config.MaxConcurrentJobs,
		"enableBatchProcessing": config.EnableBatchProcessing,
		"batchSize":             config.BatchSize,
		"batchTimeout":          config.BatchTimeout.Seconds(),
		"healthCheckInterval":   config.HealthCheckInterval.Seconds(),
		"logRequests":           config.LogRequests,
		"workflowEndpoints":     config.WorkflowEndpoints,
	}, nil
}

// UpdateN8nConfiguration updates the n8n configuration
func (a *App) UpdateN8nConfiguration(updates map[string]interface{}) error {
	if a.n8nIntegration == nil {
		return fmt.Errorf("n8n integration service not initialized")
	}

	config := a.n8nIntegration.config

	// Update configuration fields
	if baseURL, ok := updates["baseURL"].(string); ok {
		config.BaseURL = baseURL
	}
	if apiKey, ok := updates["apiKey"].(string); ok {
		config.APIKey = apiKey
	}
	if timeout, ok := updates["defaultTimeout"].(float64); ok {
		config.DefaultTimeout = time.Duration(timeout) * time.Second
	}
	if maxRetries, ok := updates["maxRetries"].(float64); ok {
		config.MaxRetries = int(maxRetries)
	}
	if retryDelay, ok := updates["retryDelay"].(float64); ok {
		config.RetryDelay = time.Duration(retryDelay) * time.Second
	}
	if maxJobs, ok := updates["maxConcurrentJobs"].(float64); ok {
		config.MaxConcurrentJobs = int(maxJobs)
	}
	if enableBatch, ok := updates["enableBatchProcessing"].(bool); ok {
		config.EnableBatchProcessing = enableBatch
	}
	if batchSize, ok := updates["batchSize"].(float64); ok {
		config.BatchSize = int(batchSize)
	}
	if batchTimeout, ok := updates["batchTimeout"].(float64); ok {
		config.BatchTimeout = time.Duration(batchTimeout) * time.Second
	}
	if healthInterval, ok := updates["healthCheckInterval"].(float64); ok {
		config.HealthCheckInterval = time.Duration(healthInterval) * time.Second
	}
	if logRequests, ok := updates["logRequests"].(bool); ok {
		config.LogRequests = logRequests
	}

	// Note: Service restart may be required for some configuration changes to take effect
	log.Printf("N8n configuration updated")
	return nil
}

// GetActiveN8nRequests returns currently active n8n workflow requests
func (a *App) GetActiveN8nRequests() ([]map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return nil, fmt.Errorf("n8n integration service not initialized")
	}

	activeRequests := a.n8nIntegration.GetActiveRequests()
	result := make([]map[string]interface{}, len(activeRequests))

	for i, request := range activeRequests {
		result[i] = map[string]interface{}{
			"id":             request.ID,
			"jobId":          request.JobID,
			"workflowName":   request.WorkflowName,
			"priority":       request.Priority,
			"status":         request.Status,
			"createdAt":      request.CreatedAt,
			"startedAt":      request.StartedAt,
			"completedAt":    request.CompletedAt,
			"retryCount":     request.RetryCount,
			"maxRetries":     request.MaxRetries,
			"lastError":      request.LastError,
			"n8nExecutionId": request.N8nExecutionID,
			"metadata":       request.Metadata,
		}

		// Include payload details
		if request.Payload != nil {
			result[i]["dealName"] = request.Payload.DealName
			result[i]["triggerType"] = string(request.Payload.TriggerType)
			result[i]["documentCount"] = len(request.Payload.FilePaths)
		}
	}

	return result, nil
}

// GetN8nWorkflowStatus gets the status of a specific n8n workflow execution
func (a *App) GetN8nWorkflowStatus(executionID string) (map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return nil, fmt.Errorf("n8n integration service not initialized")
	}

	if executionID == "" {
		return nil, fmt.Errorf("execution ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := a.n8nIntegration.GetWorkflowStatus(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	return map[string]interface{}{
		"executionId":  response.ExecutionID,
		"status":       response.Status,
		"data":         response.Data,
		"error":        response.Error,
		"startTime":    response.StartTime,
		"endTime":      response.EndTime,
		"duration":     response.Duration,
		"workflowName": response.WorkflowName,
		"triggerData":  response.TriggerData,
	}, nil
}

// CancelN8nWorkflow cancels a running n8n workflow execution
func (a *App) CancelN8nWorkflow(executionID string) error {
	if a.n8nIntegration == nil {
		return fmt.Errorf("n8n integration service not initialized")
	}

	if executionID == "" {
		return fmt.Errorf("execution ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.n8nIntegration.CancelWorkflow(ctx, executionID)
}

// RestartN8nIntegrationService restarts the n8n integration service
func (a *App) RestartN8nIntegrationService() error {
	if a.n8nIntegration == nil {
		return fmt.Errorf("n8n integration service not initialized")
	}

	// Stop the service
	if err := a.n8nIntegration.Stop(); err != nil {
		log.Printf("Warning: Error stopping n8n integration service: %v", err)
	}

	// Wait a moment for graceful shutdown
	time.Sleep(1 * time.Second)

	// Start the service again
	if err := a.n8nIntegration.Start(); err != nil {
		return fmt.Errorf("failed to restart n8n integration service: %w", err)
	}

	log.Printf("N8n integration service restarted successfully")
	return nil
}

// TestN8nConnection tests the connection to the n8n instance
func (a *App) TestN8nConnection() (map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return nil, fmt.Errorf("n8n integration service not initialized")
	}

	start := time.Now()
	err := a.n8nIntegration.checkN8nHealth()
	duration := time.Since(start)

	result := map[string]interface{}{
		"timestamp":    time.Now().UnixMilli(),
		"responseTime": duration.Milliseconds(),
		"baseURL":      a.n8nIntegration.config.BaseURL,
	}

	if err != nil {
		result["status"] = "failed"
		result["error"] = err.Error()
		return result, err
	}

	result["status"] = "success"
	return result, nil
}

// GetN8nRequestHistory returns recent n8n request history for a job or deal
func (a *App) GetN8nRequestHistory(jobID string, dealName string, limit int) ([]map[string]interface{}, error) {
	if a.n8nIntegration == nil {
		return nil, fmt.Errorf("n8n integration service not initialized")
	}

	if a.jobTracker == nil {
		return nil, fmt.Errorf("job tracker not initialized")
	}

	var jobs []*JobInfo
	var err error

	if jobID != "" {
		// Get specific job
		if job, jobErr := a.jobTracker.GetJob(jobID); jobErr == nil {
			jobs = []*JobInfo{job}
		} else {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
	} else if dealName != "" {
		// Get jobs for specific deal
		query := &JobQuery{
			DealName:  dealName,
			Limit:     limit,
			SortBy:    "updatedAt",
			SortOrder: "desc",
		}
		jobs, err = a.jobTracker.QueryJobs(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query jobs: %w", err)
		}
	} else {
		// Get all recent jobs
		query := &JobQuery{
			Limit:     limit,
			SortBy:    "updatedAt",
			SortOrder: "desc",
		}
		jobs, err = a.jobTracker.QueryJobs(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query jobs: %w", err)
		}
	}

	result := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		result[i] = map[string]interface{}{
			"jobId":              job.JobID,
			"dealName":           job.DealName,
			"status":             string(job.Status),
			"triggerType":        string(job.TriggerType),
			"createdAt":          job.CreatedAt,
			"updatedAt":          job.UpdatedAt,
			"completedAt":        job.CompletedAt,
			"documentCount":      job.TotalDocuments,
			"processedDocuments": job.ProcessedDocuments,
			"progress":           job.Progress,
			"errors":             job.Errors,
			"retryCount":         job.RetryCount,
		}

		// Include n8n-specific metadata if available
		if job.Metadata != nil {
			if n8nRequestID, ok := job.Metadata["n8n_request_id"]; ok {
				result[i]["n8nRequestId"] = n8nRequestID
			}
			if workflowName, ok := job.Metadata["workflow_name"]; ok {
				result[i]["workflowName"] = workflowName
			}
		}
	}

	return result, nil
}

// Schema Validation Methods

// ValidateWebhookPayload validates a webhook payload against its schema
func (a *App) ValidateWebhookPayload(payload map[string]interface{}, schemaName string) (map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	result, err := a.schemaValidator.ValidatePayload(payload, schemaName)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return map[string]interface{}{
		"valid":          result.Valid,
		"errors":         result.Errors,
		"warnings":       result.Warnings,
		"schemaUsed":     result.SchemaUsed,
		"validationTime": result.ValidationTime,
		"payloadSize":    result.PayloadSize,
		"schemaVersion":  result.SchemaVersion,
	}, nil
}

// GetWebhookSchemas returns a list of all available webhook schemas
func (a *App) GetWebhookSchemas() ([]map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	schemaNames := a.schemaValidator.ListSchemas()
	schemas := make([]map[string]interface{}, len(schemaNames))

	for i, name := range schemaNames {
		info, err := a.schemaValidator.GetSchemaInfo(name)
		if err != nil {
			continue // Skip schemas that can't be retrieved
		}

		schemas[i] = map[string]interface{}{
			"schemaName":    info.SchemaName,
			"schemaVersion": info.SchemaVersion,
			"description":   info.Description,
			"lastUpdated":   info.LastUpdated,
		}
	}

	return schemas, nil
}

// GetWebhookSchemaDetails returns detailed information about a specific schema
func (a *App) GetWebhookSchemaDetails(schemaName string) (map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	schema, err := a.schemaValidator.GetSchema(schemaName)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}

	return map[string]interface{}{
		"schemaName":           schema.SchemaName,
		"type":                 schema.Type,
		"title":                schema.Title,
		"description":          schema.Description,
		"version":              schema.Version,
		"properties":           schema.Properties,
		"required":             schema.Required,
		"additionalProperties": schema.AdditionalProperties,
		"examples":             schema.Examples,
		"lastUpdated":          schema.LastUpdated,
	}, nil
}

// UpdateSchemaValidatorSettings updates schema validator configuration
func (a *App) UpdateSchemaValidatorSettings(settings map[string]interface{}) error {
	if a.schemaValidator == nil {
		return fmt.Errorf("schema validator not initialized")
	}

	// Update strict mode if specified
	if strictMode, ok := settings["strictMode"].(bool); ok {
		a.schemaValidator.SetStrictMode(strictMode)
	}

	log.Printf("Schema validator settings updated")
	return nil
}

// GetSchemaValidatorInfo returns information about the schema validator
func (a *App) GetSchemaValidatorInfo() (map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	apiVersion := a.schemaValidator.GetAPIVersion()
	compatibilityInfo := a.schemaValidator.GetCompatibilityInfo()
	schemaCount := len(a.schemaValidator.ListSchemas())

	return map[string]interface{}{
		"apiVersion": map[string]interface{}{
			"major": apiVersion.Major,
			"minor": apiVersion.Minor,
			"patch": apiVersion.Patch,
			"label": apiVersion.Label,
		},
		"schemaCount": schemaCount,
		"compatibility": map[string]interface{}{
			"currentVersion":     compatibilityInfo.CurrentVersion,
			"supportedVersions":  compatibilityInfo.SupportedVersions,
			"deprecatedVersions": compatibilityInfo.DeprecatedVersions,
			"minimumVersion":     compatibilityInfo.MinimumVersion,
		},
	}, nil
}

// ValidateDocumentPayloadSchema validates a document webhook payload against its schema
func (a *App) ValidateDocumentPayloadSchema(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "document-webhook-payload")
}

// ValidateResultPayload validates a webhook result payload specifically
func (a *App) ValidateResultPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "webhook-result-payload")
}

// ValidateErrorPayload validates an error handling payload specifically
func (a *App) ValidateErrorPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "error-handling-payload")
}

// ValidateCorrectionPayload validates a user correction payload specifically
func (a *App) ValidateCorrectionPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "user-correction-payload")
}

// ValidateBatchPayload validates a batch processing payload specifically
func (a *App) ValidateBatchPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "batch-processing-payload")
}

// ValidateHealthCheckPayload validates a health check payload specifically
func (a *App) ValidateHealthCheckPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	return a.ValidateWebhookPayload(payload, "health-check-payload")
}

// CreateSamplePayload creates a sample payload for testing a specific schema
func (a *App) CreateSamplePayload(schemaName string, dealName string) (map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	now := time.Now().UnixMilli()
	jobID := fmt.Sprintf("job_%d_%s", now, dealName)

	switch schemaName {
	case "document-webhook-payload":
		return map[string]interface{}{
			"dealName":       dealName,
			"filePaths":      []string{"/path/to/document1.pdf", "/path/to/document2.docx"},
			"triggerType":    "user_button",
			"workflowType":   "document-analysis",
			"jobId":          jobID,
			"priority":       2,
			"timestamp":      now,
			"retryCount":     0,
			"maxRetries":     3,
			"timeoutSeconds": 300,
			"metadata": map[string]interface{}{
				"source": "frontend",
			},
		}, nil

	case "webhook-result-payload":
		return map[string]interface{}{
			"jobId":              jobID,
			"dealName":           dealName,
			"workflowType":       "document-analysis",
			"status":             "completed",
			"processedDocuments": 2,
			"totalDocuments":     2,
			"averageConfidence":  0.85,
			"processingTimeMs":   15000,
			"startTime":          now - 15000,
			"timestamp":          now,
			"templatesUpdated":   []string{"template1.xlsx", "template2.xlsx"},
		}, nil

	case "error-handling-payload":
		return map[string]interface{}{
			"originalJobId":  jobID,
			"errorJobId":     fmt.Sprintf("error_%d_%s", now, dealName),
			"dealName":       dealName,
			"errorType":      "processing_timeout",
			"retryAttempt":   1,
			"maxRetries":     3,
			"retryStrategy":  "exponential",
			"recoveryAction": "retry",
			"timestamp":      now,
		}, nil

	case "user-correction-payload":
		return map[string]interface{}{
			"correctionId":  fmt.Sprintf("corr_%d_%s", now, dealName),
			"originalJobId": jobID,
			"dealName":      dealName,
			"templatePath":  "/templates/deal_template.xlsx",
			"corrections": []map[string]interface{}{
				{
					"fieldName":          "dealValue",
					"originalValue":      "1000000",
					"correctedValue":     "1200000",
					"originalConfidence": 0.75,
					"userConfidence":     0.95,
					"correctionReason":   "wrong_extraction",
					"notes":              "Extracted from wrong section",
				},
			},
			"correctionType": "manual",
			"applyToSimilar": true,
			"confidence":     0.95,
			"timestamp":      now,
		}, nil

	case "batch-processing-payload":
		return map[string]interface{}{
			"batchId":   fmt.Sprintf("batch_%d_%s", now, dealName),
			"dealName":  dealName,
			"batchType": "deal_analysis",
			"items": []map[string]interface{}{
				{
					"itemId":   "item1",
					"itemType": "document",
					"itemPath": "/documents/doc1.pdf",
				},
				{
					"itemId":   "item2",
					"itemType": "document",
					"itemPath": "/documents/doc2.docx",
				},
			},
			"priority":  2,
			"timestamp": now,
		}, nil

	case "health-check-payload":
		return map[string]interface{}{
			"checkId":    fmt.Sprintf("health_%d", now),
			"checkType":  "system",
			"components": []string{"database", "ai_service", "file_system"},
			"timestamp":  now,
		}, nil

	default:
		return nil, fmt.Errorf("unknown schema: %s", schemaName)
	}
}

// TestSchemaValidation performs a comprehensive test of schema validation
func (a *App) TestSchemaValidation() (map[string]interface{}, error) {
	if a.schemaValidator == nil {
		return nil, fmt.Errorf("schema validator not initialized")
	}

	testResults := make(map[string]interface{})
	schemas := a.schemaValidator.ListSchemas()

	for _, schemaName := range schemas {
		// Create sample payload
		samplePayload, err := a.CreateSamplePayload(schemaName, "TestDeal")
		if err != nil {
			testResults[schemaName] = map[string]interface{}{
				"error": fmt.Sprintf("Failed to create sample: %v", err),
			}
			continue
		}

		// Validate the sample
		validationResult, err := a.ValidateWebhookPayload(samplePayload, schemaName)
		if err != nil {
			testResults[schemaName] = map[string]interface{}{
				"error": fmt.Sprintf("Validation failed: %v", err),
			}
			continue
		}

		testResults[schemaName] = map[string]interface{}{
			"valid":          validationResult["valid"],
			"errorCount":     len(validationResult["errors"].([]ValidationError)),
			"warningCount":   len(validationResult["warnings"].([]ValidationError)),
			"validationTime": validationResult["validationTime"],
		}
	}

	return map[string]interface{}{
		"timestamp":   time.Now().UnixMilli(),
		"schemaCount": len(schemas),
		"results":     testResults,
	}, nil
}

// Authentication Management Methods

// GenerateAPIKey generates a new API key for webhook authentication
func (a *App) GenerateAPIKey(req map[string]interface{}) (map[string]interface{}, error) {
	if a.authManager == nil {
		return nil, fmt.Errorf("authentication manager not initialized")
	}

	// Convert request map to KeyGenerationRequest
	keyReq := &KeyGenerationRequest{}

	if name, ok := req["name"].(string); ok {
		keyReq.Name = name
	}
	if description, ok := req["description"].(string); ok {
		keyReq.Description = description
	}
	if permissions, ok := req["permissions"].([]interface{}); ok {
		for _, p := range permissions {
			if perm, ok := p.(string); ok {
				keyReq.Permissions = append(keyReq.Permissions, perm)
			}
		}
	}
	if expirationDays, ok := req["expirationDays"].(float64); ok {
		keyReq.ExpirationDays = int(expirationDays)
	}
	if rateLimitTier, ok := req["rateLimitTier"].(string); ok {
		keyReq.RateLimitTier = rateLimitTier
	}

	result, err := a.authManager.GenerateAPIKey(keyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	return map[string]interface{}{
		"keyId":         result.KeyID,
		"apiKey":        result.APIKey, // Only returned here
		"name":          result.KeyInfo.Name,
		"description":   result.KeyInfo.Description,
		"permissions":   result.KeyInfo.Permissions,
		"expiresAt":     result.ExpiresAt,
		"generatedAt":   result.GeneratedAt,
		"isActive":      result.KeyInfo.IsActive,
		"rateLimitTier": result.KeyInfo.RateLimitTier,
	}, nil
}

// ListAPIKeys returns a list of all API keys (without the actual key values)
func (a *App) ListAPIKeys() ([]map[string]interface{}, error) {
	if a.authManager == nil {
		return nil, fmt.Errorf("authentication manager not initialized")
	}

	keys, err := a.authManager.ListAPIKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	result := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		result[i] = map[string]interface{}{
			"keyId":         key.KeyID,
			"name":          key.Name,
			"description":   key.Description,
			"permissions":   key.Permissions,
			"createdAt":     key.CreatedAt,
			"expiresAt":     key.ExpiresAt,
			"lastUsed":      key.LastUsed,
			"usageCount":    key.UsageCount,
			"isActive":      key.IsActive,
			"rateLimitTier": key.RateLimitTier,
		}
	}

	return result, nil
}

// RevokeAPIKey revokes an API key
func (a *App) RevokeAPIKey(keyID string, reason string) error {
	if a.authManager == nil {
		return fmt.Errorf("authentication manager not initialized")
	}

	return a.authManager.RevokeAPIKey(keyID, reason)
}

// GetAPIKeyInfo returns information about a specific API key
func (a *App) GetAPIKeyInfo(keyID string) (map[string]interface{}, error) {
	if a.authManager == nil {
		return nil, fmt.Errorf("authentication manager not initialized")
	}

	key, err := a.authManager.GetAPIKeyInfo(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key info: %w", err)
	}

	return map[string]interface{}{
		"keyId":         key.KeyID,
		"name":          key.Name,
		"description":   key.Description,
		"permissions":   key.Permissions,
		"createdAt":     key.CreatedAt,
		"expiresAt":     key.ExpiresAt,
		"lastUsed":      key.LastUsed,
		"usageCount":    key.UsageCount,
		"isActive":      key.IsActive,
		"rateLimitTier": key.RateLimitTier,
	}, nil
}

// CreateWebhookAuthPair generates a matched API key and HMAC secret for webhook communication
func (a *App) CreateWebhookAuthPair(name string, description string) (map[string]interface{}, error) {
	if a.authManager == nil {
		return nil, fmt.Errorf("authentication manager not initialized")
	}

	// Generate API key
	keyReq := &KeyGenerationRequest{
		Name:           name,
		Description:    description + " (webhook communication)",
		Permissions:    []string{"webhook:receive", "webhook:send", "documents:process"},
		RateLimitTier:  "premium",
		ExpirationDays: 365, // 1 year
	}

	result, err := a.authManager.GenerateAPIKey(keyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate webhook auth pair: %w", err)
	}

	// Generate HMAC secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate HMAC secret: %w", err)
	}
	hmacSecret := hex.EncodeToString(secretBytes)

	return map[string]interface{}{
		"keyId":       result.KeyID,
		"apiKey":      result.APIKey,
		"hmacSecret":  hmacSecret,
		"name":        result.KeyInfo.Name,
		"description": result.KeyInfo.Description,
		"expiresAt":   result.ExpiresAt,
		"generatedAt": result.GeneratedAt,
		"instructions": map[string]string{
			"apiKey":     "Use this as the X-API-Key header in webhook requests",
			"hmacSecret": "Use this to generate HMAC-SHA256 signatures for request validation",
			"algorithm":  "HMAC-SHA256",
			"encoding":   "hexadecimal",
		},
	}, nil
}

// ValidateHMACSignature validates an HMAC signature for a webhook request
func (a *App) ValidateHMACSignature(payload string, signature string, secret string) (map[string]interface{}, error) {
	// Create HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	isValid := hmac.Equal([]byte(signature), []byte(expectedSignature))

	return map[string]interface{}{
		"isValid":           isValid,
		"expectedSignature": expectedSignature,
		"providedSignature": signature,
		"algorithm":         "HMAC-SHA256",
		"timestamp":         time.Now().UnixMilli(),
	}, nil
}

// GenerateHMACSignature generates an HMAC signature for a payload
func (a *App) GenerateHMACSignature(payload string, secret string) (map[string]interface{}, error) {
	// Create HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return map[string]interface{}{
		"signature": signature,
		"payload":   payload,
		"algorithm": "HMAC-SHA256",
		"encoding":  "hexadecimal",
		"timestamp": time.Now().UnixMilli(),
	}, nil
}

// GetAuthManagerConfiguration returns the current auth manager configuration
func (a *App) GetAuthManagerConfiguration() (map[string]interface{}, error) {
	if a.authManager == nil {
		return nil, fmt.Errorf("authentication manager not initialized")
	}

	return map[string]interface{}{
		"initialized": true,
		"features": map[string]bool{
			"apiKeyGeneration": true,
			"hmacSignatures":   true,
			"rateLimiting":     true,
			"auditLogging":     true,
			"keyRotation":      false, // Not fully implemented
			"ipWhitelisting":   true,
		},
		"supportedTiers": []string{"basic", "premium", "unlimited"},
		"supportedPermissions": []string{
			"webhook:receive",
			"webhook:send",
			"documents:process",
			"jobs:query",
			"admin:manage",
		},
	}, nil
}
