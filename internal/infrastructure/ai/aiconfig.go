package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Create DealDone specific config directory
	dealDoneConfigDir := filepath.Join(configDir, "DealDone")
	if err := os.MkdirAll(dealDoneConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return dealDoneConfigDir, nil
}

// AIConfigManager manages AI service configurations
type AIConfigManager struct {
	mu         sync.RWMutex
	config     *AIConfig
	configPath string
	providers  map[Provider]interface{}
}

// NewAIConfigManager creates a new AI configuration manager
func NewAIConfigManager() (*AIConfigManager, error) {
	manager := &AIConfigManager{
		providers: make(map[Provider]interface{}),
	}

	// Set config path
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	manager.configPath = filepath.Join(configDir, "ai_config.json")

	// Load or create default config
	if err := manager.Load(); err != nil {
		// Create default config if load fails
		manager.config = manager.createDefaultConfig()
		if err := manager.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return manager, nil
}

// createDefaultConfig creates default AI configuration
func (acm *AIConfigManager) createDefaultConfig() *AIConfig {
	return &AIConfig{
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIModel: "gpt-4-turbo-preview",
		ClaudeKey:   os.Getenv("CLAUDE_API_KEY"),
		ClaudeModel: "claude-3-opus-20240229",
		CacheTTL:    time.Minute * 30,
		RateLimit:   60, // requests per minute
		MaxRetries:  3,
		RetryDelay:  time.Second * 2,

		// Additional configuration
		EnableCache:       true,
		EnableFallback:    true,
		PreferredProvider: "",
		PromptSettings: PromptSettings{
			Temperature: 0.3,
			MaxTokens:   4096,
			TopP:        0.9,
		},
		AnalysisSettings: AnalysisSettings{
			MaxDocumentSize:         10000, // characters
			ExtractKeywords:         true,
			ExtractEntities:         true,
			GenerateSummary:         true,
			MinConfidenceScore:      0.7,
			EnableRiskAnalysis:      true,
			EnableFinancialAnalysis: true,
		},
		SecuritySettings: SecuritySettings{
			RedactPII:         true,
			AllowedDomains:    []string{},
			BlockedKeywords:   []string{},
			MaxRequestsPerDoc: 10,
			EnableAuditLog:    true,
		},
	}
}

// Extended AIConfig with additional settings
type PromptSettings struct {
	Temperature   float64  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
	TopP          float64  `json:"top_p"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

type AnalysisSettings struct {
	MaxDocumentSize         int     `json:"max_document_size"`
	ExtractKeywords         bool    `json:"extract_keywords"`
	ExtractEntities         bool    `json:"extract_entities"`
	GenerateSummary         bool    `json:"generate_summary"`
	MinConfidenceScore      float64 `json:"min_confidence_score"`
	EnableRiskAnalysis      bool    `json:"enable_risk_analysis"`
	EnableFinancialAnalysis bool    `json:"enable_financial_analysis"`
}

type SecuritySettings struct {
	RedactPII         bool     `json:"redact_pii"`
	AllowedDomains    []string `json:"allowed_domains"`
	BlockedKeywords   []string `json:"blocked_keywords"`
	MaxRequestsPerDoc int      `json:"max_requests_per_doc"`
	EnableAuditLog    bool     `json:"enable_audit_log"`
}

// Update AIConfig to include new settings
func (acm *AIConfigManager) enhanceAIConfig() {
	// This would be called during migration to add new fields
	acm.mu.Lock()
	defer acm.mu.Unlock()

	if acm.config.PromptSettings.Temperature == 0 {
		acm.config.PromptSettings = PromptSettings{
			Temperature: 0.3,
			MaxTokens:   4096,
			TopP:        0.9,
		}
	}

	if acm.config.AnalysisSettings.MaxDocumentSize == 0 {
		acm.config.AnalysisSettings = AnalysisSettings{
			MaxDocumentSize:         10000,
			ExtractKeywords:         true,
			ExtractEntities:         true,
			GenerateSummary:         true,
			MinConfidenceScore:      0.7,
			EnableRiskAnalysis:      true,
			EnableFinancialAnalysis: true,
		}
	}

	if acm.config.SecuritySettings.MaxRequestsPerDoc == 0 {
		acm.config.SecuritySettings = SecuritySettings{
			RedactPII:         true,
			MaxRequestsPerDoc: 10,
			EnableAuditLog:    true,
		}
	}
}

// GetConfig returns current configuration
func (acm *AIConfigManager) GetConfig() *AIConfig {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *acm.config
	return &configCopy
}

// UpdateConfig updates configuration with validation
func (acm *AIConfigManager) UpdateConfig(updates map[string]interface{}) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	// Validate updates
	if err := acm.validateUpdates(updates); err != nil {
		return err
	}

	// Apply updates
	if err := acm.applyUpdates(updates); err != nil {
		return err
	}

	// Save to disk
	return acm.Save()
}

// validateUpdates validates configuration updates
func (acm *AIConfigManager) validateUpdates(updates map[string]interface{}) error {
	// Validate rate limit
	if rateLimit, ok := updates["rate_limit"].(float64); ok {
		if rateLimit < 1 || rateLimit > 1000 {
			return fmt.Errorf("rate_limit must be between 1 and 1000")
		}
	}

	// Validate cache TTL
	if cacheTTL, ok := updates["cache_ttl"].(string); ok {
		if _, err := time.ParseDuration(cacheTTL); err != nil {
			return fmt.Errorf("invalid cache_ttl duration: %w", err)
		}
	}

	// Validate temperature
	if temp, ok := updates["temperature"].(float64); ok {
		if temp < 0 || temp > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	// Validate confidence score
	if confidence, ok := updates["min_confidence_score"].(float64); ok {
		if confidence < 0 || confidence > 1 {
			return fmt.Errorf("min_confidence_score must be between 0 and 1")
		}
	}

	return nil
}

// applyUpdates applies validated updates to configuration
func (acm *AIConfigManager) applyUpdates(updates map[string]interface{}) error {
	// Convert updates to JSON and back to apply to config
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	// Create temporary config for partial update
	var partialConfig AIConfig
	if err := json.Unmarshal(jsonData, &partialConfig); err != nil {
		return err
	}

	// Apply non-zero values
	if partialConfig.OpenAIKey != "" {
		acm.config.OpenAIKey = partialConfig.OpenAIKey
	}
	if partialConfig.ClaudeKey != "" {
		acm.config.ClaudeKey = partialConfig.ClaudeKey
	}
	if partialConfig.RateLimit > 0 {
		acm.config.RateLimit = partialConfig.RateLimit
	}
	// ... apply other fields as needed

	return nil
}

// Load loads configuration from disk
func (acm *AIConfigManager) Load() error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	data, err := os.ReadFile(acm.configPath)
	if err != nil {
		return err
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	acm.config = &config

	// Enhance config with new fields if needed
	acm.enhanceAIConfig()

	return nil
}

// Save saves configuration to disk
func (acm *AIConfigManager) Save() error {
	data, err := json.MarshalIndent(acm.config, "", "  ")
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(acm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(acm.configPath, data, 0600)
}

// SetAPIKey sets an API key for a provider
func (acm *AIConfigManager) SetAPIKey(provider Provider, apiKey string) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	switch provider {
	case ProviderOpenAI:
		acm.config.OpenAIKey = apiKey
	case ProviderClaude:
		acm.config.ClaudeKey = apiKey
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	return acm.Save()
}

// SetPreferredProvider sets the preferred AI provider
func (acm *AIConfigManager) SetPreferredProvider(provider Provider) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	acm.config.PreferredProvider = string(provider)
	return acm.Save()
}

// IsProviderConfigured checks if a provider has valid configuration
func (acm *AIConfigManager) IsProviderConfigured(provider Provider) bool {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	switch provider {
	case ProviderOpenAI:
		return acm.config.OpenAIKey != ""
	case ProviderClaude:
		return acm.config.ClaudeKey != ""
	case ProviderDefault:
		return true
	default:
		return false
	}
}

// GetProviderStatus returns status of all providers
func (acm *AIConfigManager) GetProviderStatus() map[Provider]ProviderStatus {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	status := make(map[Provider]ProviderStatus)

	// OpenAI status
	status[ProviderOpenAI] = ProviderStatus{
		Configured: acm.config.OpenAIKey != "",
		Model:      acm.config.OpenAIModel,
		Enabled:    acm.config.EnabledProviders == nil || contains(acm.config.EnabledProviders, string(ProviderOpenAI)),
	}

	// Claude status
	status[ProviderClaude] = ProviderStatus{
		Configured: acm.config.ClaudeKey != "",
		Model:      acm.config.ClaudeModel,
		Enabled:    acm.config.EnabledProviders == nil || contains(acm.config.EnabledProviders, string(ProviderClaude)),
	}

	// Default provider always available
	status[ProviderDefault] = ProviderStatus{
		Configured: true,
		Model:      "rule-based",
		Enabled:    true,
	}

	return status
}

// ProviderStatus represents the status of an AI provider
type ProviderStatus struct {
	Configured bool   `json:"configured"`
	Model      string `json:"model"`
	Enabled    bool   `json:"enabled"`
}

// Export exports configuration (without sensitive data)
func (acm *AIConfigManager) Export() (map[string]interface{}, error) {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	export := make(map[string]interface{})

	// Export non-sensitive configuration
	export["rate_limit"] = acm.config.RateLimit
	export["cache_ttl"] = acm.config.CacheTTL.String()
	export["max_retries"] = acm.config.MaxRetries
	export["retry_delay"] = acm.config.RetryDelay.String()
	export["preferred_provider"] = acm.config.PreferredProvider
	export["prompt_settings"] = acm.config.PromptSettings
	export["analysis_settings"] = acm.config.AnalysisSettings
	export["security_settings"] = acm.config.SecuritySettings

	// Include provider status (but not keys)
	export["providers"] = acm.GetProviderStatus()

	return export, nil
}

// Import imports configuration
func (acm *AIConfigManager) Import(data map[string]interface{}) error {
	return acm.UpdateConfig(data)
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
