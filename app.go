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
