package ai

import (
	"time"
)

// Provider represents an AI service provider
type Provider string

const (
	ProviderOpenAI  Provider = "openai"
	ProviderClaude  Provider = "claude"
	ProviderDefault Provider = "default"
)

// Model represents an AI model
type Model string

const (
	ModelGPT4          Model = "gpt-4"
	ModelGPT4Turbo     Model = "gpt-4-turbo-preview"
	ModelGPT35Turbo    Model = "gpt-3.5-turbo"
	ModelClaude3Opus   Model = "claude-3-opus-20240229"
	ModelClaude3Sonnet Model = "claude-3-sonnet-20240229"
	ModelClaude3Haiku  Model = "claude-3-haiku-20240307"
)

// RequestType represents the type of AI request
type RequestType string

const (
	RequestTypeClassification   RequestType = "classification"
	RequestTypeExtraction       RequestType = "extraction"
	RequestTypeAnalysis         RequestType = "analysis"
	RequestTypeFieldMapping     RequestType = "field_mapping"
	RequestTypeEntityExtraction RequestType = "entity_extraction"
	RequestTypeRiskAnalysis     RequestType = "risk_analysis"
	RequestTypeInsights         RequestType = "insights"
)

// Priority represents the priority of an AI request
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

// Status represents the status of an AI operation
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// Request represents a generic AI request
type Request struct {
	ID          string                 `json:"id"`
	Type        RequestType            `json:"type"`
	Provider    Provider               `json:"provider"`
	Model       Model                  `json:"model"`
	Prompt      string                 `json:"prompt"`
	Context     string                 `json:"context,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Priority    Priority               `json:"priority"`
	MaxTokens   int                    `json:"maxTokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// Response represents a generic AI response
type Response struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"requestId"`
	Provider       Provider               `json:"provider"`
	Model          Model                  `json:"model"`
	Content        string                 `json:"content"`
	StructuredData interface{}            `json:"structuredData,omitempty"`
	TokensUsed     int                    `json:"tokensUsed"`
	Cost           float64                `json:"cost"`
	Latency        time.Duration          `json:"latency"`
	Status         Status                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CompletedAt    time.Time              `json:"completedAt"`
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Name          Provider               `json:"name"`
	APIKey        string                 `json:"apiKey"`
	APIEndpoint   string                 `json:"apiEndpoint,omitempty"`
	DefaultModel  Model                  `json:"defaultModel"`
	MaxRetries    int                    `json:"maxRetries"`
	RetryDelay    time.Duration          `json:"retryDelay"`
	Timeout       time.Duration          `json:"timeout"`
	RateLimit     int                    `json:"rateLimit"`
	CustomHeaders map[string]string      `json:"customHeaders,omitempty"`
	Features      []string               `json:"features,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled         bool          `json:"enabled"`
	TTL             time.Duration `json:"ttl"`
	MaxSize         int           `json:"maxSize"`
	CleanupInterval time.Duration `json:"cleanupInterval"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requestsPerMinute"`
	RequestsPerHour   int `json:"requestsPerHour"`
	RequestsPerDay    int `json:"requestsPerDay"`
	BurstSize         int `json:"burstSize"`
}

// UsageMetrics represents AI usage metrics
type UsageMetrics struct {
	Provider           Provider      `json:"provider"`
	Model              Model         `json:"model"`
	TotalRequests      int64         `json:"totalRequests"`
	SuccessfulRequests int64         `json:"successfulRequests"`
	FailedRequests     int64         `json:"failedRequests"`
	TotalTokens        int64         `json:"totalTokens"`
	TotalCost          float64       `json:"totalCost"`
	AverageLatency     time.Duration `json:"averageLatency"`
	Period             time.Duration `json:"period"`
	StartTime          time.Time     `json:"startTime"`
	EndTime            time.Time     `json:"endTime"`
}

// ProviderHealth represents the health status of an AI provider
type ProviderHealth struct {
	Provider          Provider      `json:"provider"`
	Status            string        `json:"status"`
	Latency           time.Duration `json:"latency"`
	ErrorRate         float64       `json:"errorRate"`
	LastError         string        `json:"lastError,omitempty"`
	LastErrorTime     *time.Time    `json:"lastErrorTime,omitempty"`
	LastCheckTime     time.Time     `json:"lastCheckTime"`
	ConsecutiveErrors int           `json:"consecutiveErrors"`
}

// TemplateField represents a field in a template
type TemplateField struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Required     bool                   `json:"required"`
	Description  string                 `json:"description"`
	DefaultValue interface{}            `json:"defaultValue,omitempty"`
	Format       string                 `json:"format,omitempty"`
	Validation   map[string]interface{} `json:"validation,omitempty"`
}

// ValidationRule represents a validation rule for template data
type ValidationRule struct {
	FieldName    string                 `json:"fieldName"`
	RuleType     string                 `json:"ruleType"`
	Parameters   map[string]interface{} `json:"parameters"`
	ErrorMessage string                 `json:"errorMessage"`
	Severity     string                 `json:"severity"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid     bool                   `json:"isValid"`
	Errors      []ValidationError      `json:"errors"`
	Warnings    []ValidationWarning    `json:"warnings"`
	ValidatedAt time.Time              `json:"validatedAt"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Severity string `json:"severity"`
}

// ValidationWarning is defined in aiprovider_default.go

// ConflictingValue represents a conflicting value from different sources
type ConflictingValue struct {
	Value      interface{} `json:"value"`
	Source     string      `json:"source"`
	Confidence float64     `json:"confidence"`
	Timestamp  time.Time   `json:"timestamp"`
}
