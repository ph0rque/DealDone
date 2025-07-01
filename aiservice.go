package main

import (
	"context"
	"fmt"
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
