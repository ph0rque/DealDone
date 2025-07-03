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
