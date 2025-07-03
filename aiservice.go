package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AIProvider represents different AI service providers
type AIProvider string

const (
	ProviderOpenAI  AIProvider = "openai"
	ProviderClaude  AIProvider = "claude"
	ProviderDefault AIProvider = "default"
)

// AIServiceInterface defines the contract for AI providers
type AIServiceInterface interface {
	// ClassifyDocument analyzes and classifies a document
	ClassifyDocument(ctx context.Context, content string, metadata map[string]interface{}) (*AIClassificationResult, error)

	// ExtractFinancialData extracts financial information from content
	ExtractFinancialData(ctx context.Context, content string) (*FinancialAnalysis, error)

	// AnalyzeRisks performs risk assessment on document content
	AnalyzeRisks(ctx context.Context, content string, docType string) (*RiskAnalysis, error)

	// GenerateInsights creates insights about the document
	GenerateInsights(ctx context.Context, content string, docType string) (*DocumentInsights, error)

	// ExtractEntities extracts named entities from content
	ExtractEntities(ctx context.Context, content string) (*EntityExtraction, error)

	// NEW METHODS FOR ENHANCED TEMPLATE PROCESSING

	// ExtractDocumentFields extracts structured field data from documents for template mapping
	ExtractDocumentFields(ctx context.Context, content string, documentType string, templateContext map[string]interface{}) (*DocumentFieldExtraction, error)

	// MapFieldsToTemplate maps extracted document fields to template field requirements
	MapFieldsToTemplate(ctx context.Context, extractedFields map[string]interface{}, templateFields []TemplateField, mappingContext map[string]interface{}) (*FieldMappingResult, error)

	// FormatFieldValue formats a raw field value according to template requirements (currency, dates, etc.)
	FormatFieldValue(ctx context.Context, rawValue interface{}, fieldType string, formatRequirements map[string]interface{}) (*FormattedFieldValue, error)

	// ValidateTemplateData validates that mapped data meets template requirements
	ValidateTemplateData(ctx context.Context, templateData map[string]interface{}, validationRules []ValidationRule) (*ValidationResult, error)

	// GetProvider returns the provider name
	GetProvider() AIProvider

	// IsAvailable checks if the service is configured and ready
	IsAvailable() bool

	// GetUsage returns current usage statistics
	GetUsage() *AIUsageStats
}

// AIService is the main AI service that manages multiple providers
type AIService struct {
	providers       map[AIProvider]AIServiceInterface
	primaryProvider AIProvider
	fallbackOrder   []AIProvider
	cache           *AICache
	rateLimiter     *RateLimiter
	mu              sync.RWMutex
}

// NewAIService creates a new AI service with configured providers
func NewAIService(config *AIConfig) *AIService {
	service := &AIService{
		providers:     make(map[AIProvider]AIServiceInterface),
		cache:         NewAICache(config.CacheTTL),
		rateLimiter:   NewRateLimiter(config.RateLimit),
		fallbackOrder: []AIProvider{},
	}

	// Initialize providers based on config
	if config.OpenAIKey != "" {
		service.providers[ProviderOpenAI] = NewOpenAIProvider(config.OpenAIKey, config.OpenAIModel)
		service.fallbackOrder = append(service.fallbackOrder, ProviderOpenAI)
	}

	if config.ClaudeKey != "" {
		service.providers[ProviderClaude] = NewClaudeProvider(config.ClaudeKey, config.ClaudeModel)
		service.fallbackOrder = append(service.fallbackOrder, ProviderClaude)
	}

	// Always add default provider as last fallback
	service.providers[ProviderDefault] = NewDefaultProvider()
	service.fallbackOrder = append(service.fallbackOrder, ProviderDefault)

	// Set primary provider
	if len(service.fallbackOrder) > 0 {
		service.primaryProvider = service.fallbackOrder[0]
	}

	return service
}

// AIConfig holds configuration for AI services
type AIConfig struct {
	OpenAIKey   string        `json:"openai_key"`
	OpenAIModel string        `json:"openai_model"`
	ClaudeKey   string        `json:"claude_key"`
	ClaudeModel string        `json:"claude_model"`
	CacheTTL    time.Duration `json:"cache_ttl"`
	RateLimit   int           `json:"rate_limit"` // requests per minute
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`

	// New fields
	EnableCache       bool             `json:"enable_cache"`
	EnableFallback    bool             `json:"enable_fallback"`
	PreferredProvider string           `json:"preferred_provider"`
	EnabledProviders  []string         `json:"enabled_providers,omitempty"`
	PromptSettings    PromptSettings   `json:"prompt_settings"`
	AnalysisSettings  AnalysisSettings `json:"analysis_settings"`
	SecuritySettings  SecuritySettings `json:"security_settings"`
}

// AIClassificationResult represents the classification result from AI
type AIClassificationResult struct {
	DocumentType string                 `json:"documentType"`
	Confidence   float64                `json:"confidence"`
	Keywords     []string               `json:"keywords"`
	Categories   []string               `json:"categories"`
	Language     string                 `json:"language"`
	Summary      string                 `json:"summary"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// FinancialAnalysis represents extracted financial data
type FinancialAnalysis struct {
	Revenue          float64            `json:"revenue"`
	EBITDA           float64            `json:"ebitda"`
	NetIncome        float64            `json:"netIncome"`
	TotalAssets      float64            `json:"totalAssets"`
	TotalLiabilities float64            `json:"totalLiabilities"`
	CashFlow         float64            `json:"cashFlow"`
	GrossMargin      float64            `json:"grossMargin"`
	OperatingMargin  float64            `json:"operatingMargin"`
	Confidence       float64            `json:"confidence"`
	Period           string             `json:"period"`
	Currency         string             `json:"currency"`
	DataPoints       map[string]float64 `json:"dataPoints"`
	Warnings         []string           `json:"warnings"`
}

// RiskAnalysis represents risk assessment results
type RiskAnalysis struct {
	OverallRiskScore float64    `json:"overallRiskScore"`
	RiskCategories   []RiskItem `json:"riskCategories"`
	Recommendations  []string   `json:"recommendations"`
	CriticalIssues   []string   `json:"criticalIssues"`
	Confidence       float64    `json:"confidence"`
}

// RiskItem represents a specific risk
type RiskItem struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // low, medium, high, critical
	Score       float64 `json:"score"`
	Mitigation  string  `json:"mitigation"`
}

// DocumentInsights represents AI-generated insights
type DocumentInsights struct {
	KeyPoints       []string               `json:"keyPoints"`
	Opportunities   []string               `json:"opportunities"`
	Concerns        []string               `json:"concerns"`
	ActionItems     []string               `json:"actionItems"`
	MarketContext   string                 `json:"marketContext"`
	CompetitiveInfo map[string]interface{} `json:"competitiveInfo"`
	Confidence      float64                `json:"confidence"`
}

// EntityExtraction represents extracted entities
type EntityExtraction struct {
	People         []Entity `json:"people"`
	Organizations  []Entity `json:"organizations"`
	Locations      []Entity `json:"locations"`
	Dates          []Entity `json:"dates"`
	MonetaryValues []Entity `json:"monetaryValues"`
	Percentages    []Entity `json:"percentages"`
	Products       []Entity `json:"products"`
}

// Entity represents an extracted entity
type Entity struct {
	Text       string                 `json:"text"`
	Type       string                 `json:"type"`
	Confidence float64                `json:"confidence"`
	Context    string                 `json:"context"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NEW TYPES FOR ENHANCED TEMPLATE PROCESSING

// DocumentFieldExtraction represents structured field data extracted from documents
type DocumentFieldExtraction struct {
	Fields     map[string]interface{} `json:"fields"`     // Raw extracted field values
	Confidence float64                `json:"confidence"` // Overall extraction confidence
	FieldTypes map[string]string      `json:"fieldTypes"` // Detected field types (currency, date, text, etc.)
	Metadata   map[string]interface{} `json:"metadata"`   // Additional extraction metadata
	Warnings   []string               `json:"warnings"`   // Extraction warnings or issues
	Source     string                 `json:"source"`     // Source document information
}

// FieldMappingResult represents the result of mapping document fields to template fields
type FieldMappingResult struct {
	Mappings       []FieldMapping         `json:"mappings"`       // Individual field mappings
	UnmappedFields []string               `json:"unmappedFields"` // Document fields that couldn't be mapped
	MissingFields  []string               `json:"missingFields"`  // Required template fields not found
	Confidence     float64                `json:"confidence"`     // Overall mapping confidence
	Suggestions    []MappingSuggestion    `json:"suggestions"`    // Alternative mapping suggestions
	Metadata       map[string]interface{} `json:"metadata"`       // Additional mapping metadata
}

// FieldMapping represents a single field mapping from document to template
type FieldMapping struct {
	DocumentField    string      `json:"documentField"`    // Source field from document
	TemplateField    string      `json:"templateField"`    // Target field in template
	Value            interface{} `json:"value"`            // Mapped value
	Confidence       float64     `json:"confidence"`       // Mapping confidence
	TransformApplied string      `json:"transformApplied"` // Any transformation applied
}

// MappingSuggestion represents an alternative mapping suggestion
type MappingSuggestion struct {
	DocumentField string  `json:"documentField"`
	TemplateField string  `json:"templateField"`
	Confidence    float64 `json:"confidence"`
	Reason        string  `json:"reason"`
}

// FormattedFieldValue represents a formatted field value
type FormattedFieldValue struct {
	FormattedValue string                 `json:"formattedValue"` // The formatted value
	OriginalValue  interface{}            `json:"originalValue"`  // Original raw value
	FormatApplied  string                 `json:"formatApplied"`  // Format that was applied
	Confidence     float64                `json:"confidence"`     // Formatting confidence
	Warnings       []string               `json:"warnings"`       // Any formatting warnings
	Metadata       map[string]interface{} `json:"metadata"`       // Additional formatting metadata
}

// AIUsageStats tracks AI service usage
type AIUsageStats struct {
	TotalRequests   int64            `json:"totalRequests"`
	SuccessfulCalls int64            `json:"successfulCalls"`
	FailedCalls     int64            `json:"failedCalls"`
	CacheHits       int64            `json:"cacheHits"`
	TotalTokens     int64            `json:"totalTokens"`
	ProviderStats   map[string]int64 `json:"providerStats"`
	LastReset       time.Time        `json:"lastReset"`
}

// ClassifyDocument uses AI to classify a document with fallback support
func (as *AIService) ClassifyDocument(ctx context.Context, content string, metadata map[string]interface{}) (*AIClassificationResult, error) {
	// Check cache first
	cacheKey := as.cache.GenerateKey("classify", content, metadata)
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*AIClassificationResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try each provider in fallback order
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.ClassifyDocument(ctx, content, metadata)
			if err == nil {
				// Cache successful result
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("all AI providers failed: %w", lastError)
}

// ExtractFinancialData extracts financial information with provider fallback
func (as *AIService) ExtractFinancialData(ctx context.Context, content string) (*FinancialAnalysis, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("financial", content, nil)
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*FinancialAnalysis); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.ExtractFinancialData(ctx, content)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("financial extraction failed: %w", lastError)
}

// SetPrimaryProvider changes the primary AI provider
func (as *AIService) SetPrimaryProvider(provider AIProvider) error {
	if _, exists := as.providers[provider]; !exists {
		return fmt.Errorf("provider %s not configured", provider)
	}

	as.primaryProvider = provider

	// Reorder fallback list to put primary first
	newOrder := []AIProvider{provider}
	for _, p := range as.fallbackOrder {
		if p != provider {
			newOrder = append(newOrder, p)
		}
	}
	as.fallbackOrder = newOrder

	return nil
}

// GetAvailableProviders returns list of configured providers
func (as *AIService) GetAvailableProviders() []AIProvider {
	providers := []AIProvider{}
	for provider, p := range as.providers {
		if p.IsAvailable() {
			providers = append(providers, provider)
		}
	}
	return providers
}

// IsAvailable checks if at least one AI provider is available
func (as *AIService) IsAvailable() bool {
	return len(as.GetAvailableProviders()) > 0
}

// AnalyzeRisks performs risk assessment with caching and fallback
func (as *AIService) AnalyzeRisks(ctx context.Context, content string, docType string) (*RiskAnalysis, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("risk", content, map[string]interface{}{"docType": docType})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*RiskAnalysis); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.AnalyzeRisks(ctx, content, docType)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("risk analysis failed: %w", lastError)
}

func (as *AIService) GenerateInsights(ctx context.Context, content string, docType string) (*DocumentInsights, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("insights", content, map[string]interface{}{"docType": docType})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*DocumentInsights); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.GenerateInsights(ctx, content, docType)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("insight generation failed: %w", lastError)
}

func (as *AIService) ExtractEntities(ctx context.Context, content string) (*EntityExtraction, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("entities", content, nil)
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*EntityExtraction); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.ExtractEntities(ctx, content)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("entity extraction failed: %w", lastError)
}

// NEW TEMPLATE PROCESSING METHODS WITH FALLBACK SUPPORT

// ExtractDocumentFields extracts structured field data with provider fallback
func (as *AIService) ExtractDocumentFields(ctx context.Context, content string, documentType string, templateContext map[string]interface{}) (*DocumentFieldExtraction, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("extract_fields", content, map[string]interface{}{
		"documentType": documentType,
		"context":      templateContext,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*DocumentFieldExtraction); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.ExtractDocumentFields(ctx, content, documentType, templateContext)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("document field extraction failed: %w", lastError)
}

// MapFieldsToTemplate maps fields to template with provider fallback
func (as *AIService) MapFieldsToTemplate(ctx context.Context, extractedFields map[string]interface{}, templateFields []TemplateField, mappingContext map[string]interface{}) (*FieldMappingResult, error) {
	// Check cache
	extractedFieldsStr := fmt.Sprintf("%v", extractedFields)
	cacheKey := as.cache.GenerateKey("map_fields", extractedFieldsStr, map[string]interface{}{
		"templateFields": templateFields,
		"context":        mappingContext,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*FieldMappingResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.MapFieldsToTemplate(ctx, extractedFields, templateFields, mappingContext)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("field mapping failed: %w", lastError)
}

// FormatFieldValue formats field values with provider fallback
func (as *AIService) FormatFieldValue(ctx context.Context, rawValue interface{}, fieldType string, formatRequirements map[string]interface{}) (*FormattedFieldValue, error) {
	// Check cache
	rawValueStr := fmt.Sprintf("%v", rawValue)
	cacheKey := as.cache.GenerateKey("format_field", rawValueStr, map[string]interface{}{
		"fieldType":          fieldType,
		"formatRequirements": formatRequirements,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*FormattedFieldValue); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.FormatFieldValue(ctx, rawValue, fieldType, formatRequirements)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("field formatting failed: %w", lastError)
}

// ValidateTemplateData validates template data with provider fallback
func (as *AIService) ValidateTemplateData(ctx context.Context, templateData map[string]interface{}, validationRules []ValidationRule) (*ValidationResult, error) {
	// Check cache
	templateDataStr := fmt.Sprintf("%v", templateData)
	cacheKey := as.cache.GenerateKey("validate_data", templateDataStr, map[string]interface{}{
		"validationRules": validationRules,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*ValidationResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			result, err := p.ValidateTemplateData(ctx, templateData, validationRules)
			if err == nil {
				as.cache.Set(cacheKey, result)
				return result, nil
			}
			lastError = err
		}
	}

	return nil, fmt.Errorf("template validation failed: %w", lastError)
}

// GetConfiguration returns the current AI service configuration
func (s *AIService) GetConfiguration() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config := map[string]interface{}{
		"provider":    s.primaryProvider,
		"enableCache": s.cache != nil,
		"cacheExpiry": 3600, // Default cache expiry
		"maxTokens":   4000,
		"temperature": 0.7,
	}

	// Add provider-specific config
	if s.primaryProvider == "openai" {
		config["modelName"] = "gpt-4"
	} else if s.primaryProvider == "claude" {
		config["modelName"] = "claude-3-opus"
	}

	return config
}

// NEW ENHANCED ENTITY EXTRACTION METHODS FOR TASK 1.3

// ExtractCompanyAndDealNames extracts company names and deal names with enhanced confidence scoring
func (as *AIService) ExtractCompanyAndDealNames(ctx context.Context, content string, documentType string) (*CompanyDealExtraction, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("company_deal_extract", content, map[string]interface{}{
		"documentType": documentType,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*CompanyDealExtraction); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if enhancedProvider, ok := p.(EnhancedEntityExtractorInterface); ok {
				result, err := enhancedProvider.ExtractCompanyAndDealNames(ctx, content, documentType)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("company and deal name extraction failed: %w", lastError)
}

// ExtractFinancialMetrics extracts financial metrics with enhanced validation
func (as *AIService) ExtractFinancialMetrics(ctx context.Context, content string, documentType string) (*FinancialMetricsExtraction, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("financial_metrics", content, map[string]interface{}{
		"documentType": documentType,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*FinancialMetricsExtraction); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if enhancedProvider, ok := p.(EnhancedEntityExtractorInterface); ok {
				result, err := enhancedProvider.ExtractFinancialMetrics(ctx, content, documentType)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("financial metrics extraction failed: %w", lastError)
}

// ExtractPersonnelAndRoles extracts personnel information and roles
func (as *AIService) ExtractPersonnelAndRoles(ctx context.Context, content string, documentType string) (*PersonnelRoleExtraction, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("personnel_roles", content, map[string]interface{}{
		"documentType": documentType,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*PersonnelRoleExtraction); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if enhancedProvider, ok := p.(EnhancedEntityExtractorInterface); ok {
				result, err := enhancedProvider.ExtractPersonnelAndRoles(ctx, content, documentType)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("personnel and roles extraction failed: %w", lastError)
}

// ValidateEntitiesAcrossDocuments validates and resolves conflicts between entities from multiple documents
func (as *AIService) ValidateEntitiesAcrossDocuments(ctx context.Context, documentExtractions []DocumentEntityExtraction) (*CrossDocumentValidation, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("cross_doc_validation", "", map[string]interface{}{
		"documentCount": len(documentExtractions),
		"extractionIds": extractDocumentIds(documentExtractions),
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*CrossDocumentValidation); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if enhancedProvider, ok := p.(EnhancedEntityExtractorInterface); ok {
				result, err := enhancedProvider.ValidateEntitiesAcrossDocuments(ctx, documentExtractions)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("cross-document entity validation failed: %w", lastError)
}

// Helper function to extract document IDs for cache key generation
func extractDocumentIds(extractions []DocumentEntityExtraction) []string {
	ids := make([]string, len(extractions))
	for i, extraction := range extractions {
		ids[i] = extraction.DocumentID
	}
	return ids
}

// NEW ENHANCED ENTITY EXTRACTION INTERFACE
type EnhancedEntityExtractorInterface interface {
	// ExtractCompanyAndDealNames extracts company names and deal names with enhanced confidence scoring
	ExtractCompanyAndDealNames(ctx context.Context, content string, documentType string) (*CompanyDealExtraction, error)

	// ExtractFinancialMetrics extracts financial metrics with enhanced validation
	ExtractFinancialMetrics(ctx context.Context, content string, documentType string) (*FinancialMetricsExtraction, error)

	// ExtractPersonnelAndRoles extracts personnel information and roles
	ExtractPersonnelAndRoles(ctx context.Context, content string, documentType string) (*PersonnelRoleExtraction, error)

	// ValidateEntitiesAcrossDocuments validates and resolves conflicts between entities from multiple documents
	ValidateEntitiesAcrossDocuments(ctx context.Context, documentExtractions []DocumentEntityExtraction) (*CrossDocumentValidation, error)
}

// CompanyDealExtraction represents extracted company and deal information
type CompanyDealExtraction struct {
	Companies  []CompanyEntity        `json:"companies"`
	DealNames  []DealEntity           `json:"dealNames"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
	Warnings   []string               `json:"warnings"`
}

// CompanyEntity represents a company with enhanced metadata
type CompanyEntity struct {
	Name       string                 `json:"name"`
	Role       string                 `json:"role"` // buyer, seller, target, advisor
	Confidence float64                `json:"confidence"`
	Context    string                 `json:"context"`
	Industry   string                 `json:"industry"`
	Location   string                 `json:"location"`
	Metadata   map[string]interface{} `json:"metadata"`
	Validated  bool                   `json:"validated"` // validated against known databases
}

// DealEntity represents deal information
type DealEntity struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`   // acquisition, merger, investment
	Status     string                 `json:"status"` // proposed, pending, completed
	Confidence float64                `json:"confidence"`
	Context    string                 `json:"context"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// FinancialMetricsExtraction represents extracted financial metrics with validation
type FinancialMetricsExtraction struct {
	Revenue          FinancialMetric            `json:"revenue"`
	EBITDA           FinancialMetric            `json:"ebitda"`
	NetIncome        FinancialMetric            `json:"netIncome"`
	TotalAssets      FinancialMetric            `json:"totalAssets"`
	TotalLiabilities FinancialMetric            `json:"totalLiabilities"`
	CashFlow         FinancialMetric            `json:"cashFlow"`
	DealValue        FinancialMetric            `json:"dealValue"`
	Multiples        map[string]FinancialMetric `json:"multiples"`
	Ratios           map[string]FinancialMetric `json:"ratios"`
	Currency         string                     `json:"currency"`
	Period           string                     `json:"period"`
	Confidence       float64                    `json:"confidence"`
	Validated        bool                       `json:"validated"`
	Warnings         []string                   `json:"warnings"`
	Metadata         map[string]interface{}     `json:"metadata"`
}

// FinancialMetric represents a financial value with metadata
type FinancialMetric struct {
	Value      float64 `json:"value"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Context    string  `json:"context"`
	Unit       string  `json:"unit"` // millions, thousands, etc.
	Validated  bool    `json:"validated"`
}

// PersonnelRoleExtraction represents extracted personnel and role information
type PersonnelRoleExtraction struct {
	Personnel  []PersonEntity         `json:"personnel"`
	Contacts   []ContactEntity        `json:"contacts"`
	Hierarchy  []HierarchyRelation    `json:"hierarchy"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
	Warnings   []string               `json:"warnings"`
}

// PersonEntity represents a person with role information
type PersonEntity struct {
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Company    string                 `json:"company"`
	Role       string                 `json:"role"` // decision_maker, advisor, contact
	Department string                 `json:"department"`
	Confidence float64                `json:"confidence"`
	Context    string                 `json:"context"`
	Contact    ContactInfo            `json:"contact"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ContactEntity represents contact information
type ContactEntity struct {
	Email      string                 `json:"email"`
	Phone      string                 `json:"phone"`
	Address    string                 `json:"address"`
	Company    string                 `json:"company"`
	Confidence float64                `json:"confidence"`
	Context    string                 `json:"context"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// HierarchyRelation represents organizational hierarchy
type HierarchyRelation struct {
	Superior    string  `json:"superior"`
	Subordinate string  `json:"subordinate"`
	Confidence  float64 `json:"confidence"`
	Context     string  `json:"context"`
}

// ContactInfo represents contact details
type ContactInfo struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// DocumentEntityExtraction represents entity extraction from a single document
type DocumentEntityExtraction struct {
	DocumentID   string                      `json:"documentId"`
	DocumentType string                      `json:"documentType"`
	Companies    *CompanyDealExtraction      `json:"companies"`
	Financial    *FinancialMetricsExtraction `json:"financial"`
	Personnel    *PersonnelRoleExtraction    `json:"personnel"`
	Confidence   float64                     `json:"confidence"`
	Metadata     map[string]interface{}      `json:"metadata"`
}

// CrossDocumentValidation represents validation results across multiple documents
type CrossDocumentValidation struct {
	ConsolidatedEntities ConsolidatedEntities   `json:"consolidatedEntities"`
	Conflicts            []EntityConflict       `json:"conflicts"`
	Resolutions          []ConflictResolution   `json:"resolutions"`
	Confidence           float64                `json:"confidence"`
	Summary              ValidationSummary      `json:"summary"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// ConsolidatedEntities represents validated entities across documents
type ConsolidatedEntities struct {
	Companies []CompanyEntity            `json:"companies"`
	Financial FinancialMetricsExtraction `json:"financial"`
	Personnel []PersonEntity             `json:"personnel"`
	Deals     []DealEntity               `json:"deals"`
}

// EntityConflict represents a conflict between entities from different documents
type EntityConflict struct {
	Type        string                 `json:"type"` // company, financial, personnel
	Field       string                 `json:"field"`
	Values      []ConflictValue        `json:"values"`
	Severity    string                 `json:"severity"` // low, medium, high
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ConflictValue represents a conflicting value
type ConflictValue struct {
	Value      interface{} `json:"value"`
	Source     string      `json:"source"`
	Confidence float64     `json:"confidence"`
	DocumentID string      `json:"documentId"`
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	ConflictID  string                 `json:"conflictId"`
	Resolution  string                 `json:"resolution"` // auto, manual, precedence
	ChosenValue interface{}            `json:"chosenValue"`
	Reasoning   string                 `json:"reasoning"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationSummary provides overall validation statistics
type ValidationSummary struct {
	TotalEntities     int     `json:"totalEntities"`
	ValidatedEntities int     `json:"validatedEntities"`
	ConflictsFound    int     `json:"conflictsFound"`
	ConflictsResolved int     `json:"conflictsResolved"`
	OverallConfidence float64 `json:"overallConfidence"`
}

// SEMANTIC FIELD MAPPING ENGINE INTERFACE FOR TASK 2.1
type SemanticFieldMappingInterface interface {
	// AnalyzeFieldSemantics analyzes field meaning and context for semantic understanding
	AnalyzeFieldSemantics(ctx context.Context, fieldName string, fieldValue interface{}, documentContext string) (*FieldSemanticAnalysis, error)

	// CreateSemanticMapping creates intelligent field mappings based on semantic understanding
	CreateSemanticMapping(ctx context.Context, sourceFields map[string]interface{}, templateFields []string, documentType string) (*SemanticMappingResult, error)

	// ResolveFieldConflicts resolves conflicts when multiple sources provide different values for the same field
	ResolveFieldConflicts(ctx context.Context, conflicts []FieldConflict, resolutionContext *ConflictResolutionContext) (*ConflictResolutionResult, error)

	// AnalyzeTemplateStructure analyzes template structure and field requirements
	AnalyzeTemplateStructure(ctx context.Context, templatePath string, templateContent []byte) (*TemplateStructureAnalysis, error)

	// ValidateFieldMapping validates the logical consistency and business rule compliance of field mappings
	ValidateFieldMapping(ctx context.Context, mapping *FieldMapping, validationRules []ValidationRule) (*MappingValidationResult, error)
}

// SEMANTIC FIELD MAPPING DATA STRUCTURES FOR TASK 2.1

// FieldSemanticAnalysis represents the semantic analysis of a field
type FieldSemanticAnalysis struct {
	FieldName        string                 `json:"field_name"`
	SemanticType     string                 `json:"semantic_type"`     // e.g., "currency", "date", "company_name", "percentage"
	BusinessCategory string                 `json:"business_category"` // e.g., "financial", "entity", "legal", "operational"
	DataType         string                 `json:"data_type"`         // e.g., "number", "string", "date", "boolean"
	ExpectedFormat   string                 `json:"expected_format"`   // e.g., "$#,##0.00", "MM/dd/yyyy", "Title Case"
	ConfidenceScore  float64                `json:"confidence_score"`  // 0.0 to 1.0
	Context          string                 `json:"context"`           // Surrounding context that influenced the analysis
	Metadata         map[string]interface{} `json:"metadata"`
	Suggestions      []string               `json:"suggestions"`    // Alternative interpretations
	BusinessRules    []string               `json:"business_rules"` // Applicable business rules
}

// SemanticMappingResult represents the result of semantic field mapping
type SemanticMappingResult struct {
	Mappings          []SemanticFieldMapping `json:"mappings"`
	UnmappedSource    []string               `json:"unmapped_source"`   // Source fields that couldn't be mapped
	UnmappedTemplate  []string               `json:"unmapped_template"` // Template fields that couldn't be filled
	OverallConfidence float64                `json:"overall_confidence"`
	MappingStrategy   string                 `json:"mapping_strategy"` // Strategy used for mapping
	Metadata          map[string]interface{} `json:"metadata"`
	Warnings          []string               `json:"warnings"`
	Recommendations   []string               `json:"recommendations"`
}

// SemanticFieldMapping represents a semantic mapping between source and template fields
type SemanticFieldMapping struct {
	SourceField           string               `json:"source_field"`
	TemplateField         string               `json:"template_field"`
	MappingType           string               `json:"mapping_type"` // "direct", "transformation", "aggregation", "calculation"
	Confidence            float64              `json:"confidence"`
	Transformation        *FieldTransformation `json:"transformation,omitempty"`
	BusinessJustification string               `json:"business_justification"`
	AlternativeMappings   []AlternativeMapping `json:"alternative_mappings,omitempty"`
}

// FieldTransformation represents a transformation applied to field data
type FieldTransformation struct {
	Type        string                 `json:"type"`        // "format", "calculate", "lookup", "aggregate"
	Function    string                 `json:"function"`    // Specific transformation function
	Parameters  map[string]interface{} `json:"parameters"`  // Transformation parameters
	Description string                 `json:"description"` // Human-readable description
}

// AlternativeMapping represents alternative mapping options
type AlternativeMapping struct {
	TemplateField string  `json:"template_field"`
	Confidence    float64 `json:"confidence"`
	Justification string  `json:"justification"`
}

// FieldConflict represents a conflict between multiple field values
type FieldConflict struct {
	FieldName    string                 `json:"field_name"`
	Values       []ConflictingValue     `json:"values"`
	ConflictType string                 `json:"conflict_type"` // "value_mismatch", "format_difference", "data_type_conflict"
	Severity     string                 `json:"severity"`      // "low", "medium", "high", "critical"
	Context      map[string]interface{} `json:"context"`
}

// ConflictingValue is defined in types.go

// ConflictResolutionContext provides context for resolving conflicts
type ConflictResolutionContext struct {
	DocumentTypes   []string               `json:"document_types"`
	BusinessRules   []BusinessRule         `json:"business_rules"`
	UserPreferences map[string]interface{} `json:"user_preferences"`
	HistoricalData  map[string]interface{} `json:"historical_data"`
	QualityMetrics  map[string]float64     `json:"quality_metrics"`
}

// ConflictResolutionResult represents the result of conflict resolution
type ConflictResolutionResult struct {
	ResolvedValues    map[string]interface{} `json:"resolved_values"`
	ResolutionMethod  string                 `json:"resolution_method"` // "confidence_based", "rule_based", "user_preference", "manual_review"
	Confidence        float64                `json:"confidence"`
	Justification     string                 `json:"justification"`
	RequiresReview    bool                   `json:"requires_review"`
	AlternativeValues []interface{}          `json:"alternative_values,omitempty"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// BusinessRule represents a business rule for validation
type BusinessRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"` // "validation", "transformation", "precedence"
	Condition   string                 `json:"condition"` // Rule condition
	Action      string                 `json:"action"`    // Action to take
	Priority    int                    `json:"priority"`  // Higher numbers = higher priority
	Metadata    map[string]interface{} `json:"metadata"`
}

// TemplateStructureAnalysis represents analysis of template structure
type TemplateStructureAnalysis struct {
	TemplateName       string                 `json:"template_name"`
	TemplateType       string                 `json:"template_type"` // "excel", "word", "pdf"
	Fields             []TemplateField        `json:"fields"`
	Sections           []TemplateSection      `json:"sections"`
	Relationships      []FieldRelationship    `json:"relationships"`
	RequiredFields     []string               `json:"required_fields"`
	OptionalFields     []string               `json:"optional_fields"`
	CalculatedFields   []CalculatedField      `json:"calculated_fields"`
	ValidationRules    []ValidationRule       `json:"validation_rules"`
	Complexity         string                 `json:"complexity"` // "simple", "moderate", "complex"
	CompatibilityScore float64                `json:"compatibility_score"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// TemplateField is defined in templatediscovery.go

// TemplateSection represents a logical section in a template
type TemplateSection struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`     // "header", "data", "summary", "footer"
	Fields      []string `json:"fields"`   // Field names in this section
	Location    string   `json:"location"` // Range or position
	Description string   `json:"description"`
}

// FieldRelationship represents relationships between fields
type FieldRelationship struct {
	SourceField      string                 `json:"source_field"`
	TargetField      string                 `json:"target_field"`
	RelationshipType string                 `json:"relationship_type"` // "depends_on", "calculates_from", "validates_against"
	Description      string                 `json:"description"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CalculatedField represents a field that is calculated from other fields
type CalculatedField struct {
	Name        string                 `json:"name"`
	Formula     string                 `json:"formula"`
	InputFields []string               `json:"input_fields"`
	OutputType  string                 `json:"output_type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationRule is defined in correctionprocessor.go

// MappingValidationResult represents the result of mapping validation
type MappingValidationResult struct {
	IsValid           bool                    `json:"is_valid"`
	OverallScore      float64                 `json:"overall_score"` // 0.0 to 1.0
	ValidationResults []FieldValidationResult `json:"validation_results"`
	Errors            []ValidationError       `json:"errors"`
	Warnings          []ValidationWarning     `json:"warnings"`
	Recommendations   []string                `json:"recommendations"`
	AuditTrail        []AuditEntry            `json:"audit_trail"`
	Metadata          map[string]interface{}  `json:"metadata"`
}

// FieldValidationResult represents validation result for a specific field
type FieldValidationResult struct {
	FieldName      string                 `json:"field_name"`
	IsValid        bool                   `json:"is_valid"`
	Score          float64                `json:"score"`
	AppliedRules   []string               `json:"applied_rules"`
	ValidationTime string                 `json:"validation_time"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// AuditEntry represents an entry in the audit trail
type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Action    string                 `json:"action"`
	User      string                 `json:"user,omitempty"`
	Details   string                 `json:"details"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SEMANTIC FIELD MAPPING ENGINE IMPLEMENTATION METHODS FOR TASK 2.1

// AnalyzeFieldSemantics analyzes field meaning and context for semantic understanding
func (as *AIService) AnalyzeFieldSemantics(ctx context.Context, fieldName string, fieldValue interface{}, documentContext string) (*FieldSemanticAnalysis, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("field_semantics", fieldName, map[string]interface{}{
		"fieldValue":      fieldValue,
		"documentContext": documentContext,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*FieldSemanticAnalysis); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if semanticProvider, ok := p.(SemanticFieldMappingInterface); ok {
				result, err := semanticProvider.AnalyzeFieldSemantics(ctx, fieldName, fieldValue, documentContext)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("field semantic analysis failed: %w", lastError)
}

// CreateSemanticMapping creates intelligent field mappings based on semantic understanding
func (as *AIService) CreateSemanticMapping(ctx context.Context, sourceFields map[string]interface{}, templateFields []string, documentType string) (*SemanticMappingResult, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("semantic_mapping", "", map[string]interface{}{
		"sourceFieldCount":   len(sourceFields),
		"templateFieldCount": len(templateFields),
		"documentType":       documentType,
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*SemanticMappingResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if semanticProvider, ok := p.(SemanticFieldMappingInterface); ok {
				result, err := semanticProvider.CreateSemanticMapping(ctx, sourceFields, templateFields, documentType)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("semantic mapping creation failed: %w", lastError)
}

// ResolveFieldConflicts resolves conflicts when multiple sources provide different values for the same field
func (as *AIService) ResolveFieldConflicts(ctx context.Context, conflicts []FieldConflict, resolutionContext *ConflictResolutionContext) (*ConflictResolutionResult, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("conflict_resolution", "", map[string]interface{}{
		"conflictCount": len(conflicts),
		"contextHash":   generateContextHash(resolutionContext),
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*ConflictResolutionResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if semanticProvider, ok := p.(SemanticFieldMappingInterface); ok {
				result, err := semanticProvider.ResolveFieldConflicts(ctx, conflicts, resolutionContext)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("field conflict resolution failed: %w", lastError)
}

// AnalyzeTemplateStructure analyzes template structure and field requirements
func (as *AIService) AnalyzeTemplateStructure(ctx context.Context, templatePath string, templateContent []byte) (*TemplateStructureAnalysis, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("template_structure", templatePath, map[string]interface{}{
		"contentSize": len(templateContent),
		"contentHash": generateContentHash(templateContent),
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*TemplateStructureAnalysis); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if semanticProvider, ok := p.(SemanticFieldMappingInterface); ok {
				result, err := semanticProvider.AnalyzeTemplateStructure(ctx, templatePath, templateContent)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("template structure analysis failed: %w", lastError)
}

// ValidateFieldMapping validates the logical consistency and business rule compliance of field mappings
func (as *AIService) ValidateFieldMapping(ctx context.Context, mapping *FieldMapping, validationRules []ValidationRule) (*MappingValidationResult, error) {
	// Check cache
	cacheKey := as.cache.GenerateKey("mapping_validation", "", map[string]interface{}{
		"mappingHash": generateMappingHash(mapping),
		"rulesCount":  len(validationRules),
	})
	if cached := as.cache.Get(cacheKey); cached != nil {
		if result, ok := cached.(*MappingValidationResult); ok {
			return result, nil
		}
	}

	// Rate limiting
	if err := as.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Try providers with fallback
	var lastError error
	for _, provider := range as.fallbackOrder {
		if p, exists := as.providers[provider]; exists && p.IsAvailable() {
			if semanticProvider, ok := p.(SemanticFieldMappingInterface); ok {
				result, err := semanticProvider.ValidateFieldMapping(ctx, mapping, validationRules)
				if err == nil {
					as.cache.Set(cacheKey, result)
					return result, nil
				}
				lastError = err
			}
		}
	}

	return nil, fmt.Errorf("field mapping validation failed: %w", lastError)
}

// Helper functions for cache key generation
func generateContextHash(context *ConflictResolutionContext) string {
	if context == nil {
		return "nil"
	}
	return fmt.Sprintf("dt_%d_br_%d_up_%d", len(context.DocumentTypes), len(context.BusinessRules), len(context.UserPreferences))
}

func generateContentHash(content []byte) string {
	if len(content) == 0 {
		return "empty"
	}
	// Simple hash based on content size and first/last bytes
	if len(content) < 10 {
		return fmt.Sprintf("small_%x", content)
	}
	return fmt.Sprintf("hash_%d_%x_%x", len(content), content[:5], content[len(content)-5:])
}

func generateMappingHash(mapping *FieldMapping) string {
	if mapping == nil {
		return "nil"
	}
	return fmt.Sprintf("map_%s_%s_%.2f", mapping.DocumentField, mapping.TemplateField, mapping.Confidence)
}
