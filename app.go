package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// App struct
type App struct {
	ctx               context.Context
	configService     *ConfigService
	folderManager     *FolderManager
	templateManager   *TemplateManager
	documentRouter    *DocumentRouter
	documentProcessor *DocumentProcessor
	aiConfigManager   *AIConfigManager
	aiService         *AIService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config service
	configService, err := NewConfigService()
	if err != nil {
		// Log error but continue - will show in UI
		fmt.Printf("Error initializing config: %v\n", err)
		return
	}
	a.configService = configService

	// Initialize folder manager
	a.folderManager = NewFolderManager(configService)

	// Initialize template manager
	a.templateManager = NewTemplateManager(configService)

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

	// Initialize document processor and router
	a.documentProcessor = NewDocumentProcessor(aiService)
	a.documentRouter = NewDocumentRouter(a.folderManager, a.documentProcessor)
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
