package deployment

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ConfigurationManager manages deployment configurations and environment variables
type ConfigurationManager struct {
	mu                   sync.RWMutex
	environments         map[string]*EnvironmentConfig
	globalConfiguration  *GlobalConfig
	secretsManager       *SecretsManager
	configurationSources []ConfigurationSource
	changeListeners      []ConfigurationChangeListener
	configurationCache   map[string]interface{}
	cacheExpiry          time.Duration
	lastUpdate           time.Time
}

// EnvironmentConfig represents configuration for a specific environment
type EnvironmentConfig struct {
	Name             string                 `json:"name"`
	Type             EnvironmentType        `json:"type"`
	Variables        map[string]string      `json:"variables"`
	Secrets          map[string]string      `json:"secrets"`
	DatabaseConfig   *DatabaseConfig        `json:"database_config"`
	AIProviderConfig *AIProviderConfig      `json:"ai_provider_config"`
	N8NConfig        *N8NConfig            `json:"n8n_config"`
	MonitoringConfig *MonitoringConfig      `json:"monitoring_config"`
	FeatureFlags     map[string]bool        `json:"feature_flags"`
	ResourceLimits   *ResourceLimits        `json:"resource_limits"`
	SecurityConfig   *SecurityConfig        `json:"security_config"`
	LoggingConfig    *LoggingConfig         `json:"logging_config"`
	Version          string                 `json:"version"`
	LastUpdated      time.Time              `json:"last_updated"`
}

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
	ApplicationName    string                 `json:"application_name"`
	Version           string                 `json:"version"`
	DefaultTimeout    time.Duration          `json:"default_timeout"`
	MaxRetries        int                    `json:"max_retries"`
	RateLimits        map[string]RateLimit   `json:"rate_limits"`
	DefaultFeatureFlags map[string]bool       `json:"default_feature_flags"`
	SecurityDefaults  *SecurityDefaults      `json:"security_defaults"`
	LoggingDefaults   *LoggingDefaults       `json:"logging_defaults"`
	MonitoringDefaults *MonitoringDefaults   `json:"monitoring_defaults"`
	BackupConfig      *BackupConfig          `json:"backup_config"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Database        string        `json:"database"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	MaxConnections  int           `json:"max_connections"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	SSLMode         string        `json:"ssl_mode"`
	BackupEnabled   bool          `json:"backup_enabled"`
}

// AIProviderConfig represents AI provider configuration
type AIProviderConfig struct {
	OpenAIConfig  *OpenAIConfig  `json:"openai_config"`
	ClaudeConfig  *ClaudeConfig  `json:"claude_config"`
	DefaultProvider string        `json:"default_provider"`
	MaxTokens     int            `json:"max_tokens"`
	Temperature   float64        `json:"temperature"`
	Timeout       time.Duration  `json:"timeout"`
	RateLimits    map[string]int `json:"rate_limits"`
}

// OpenAIConfig represents OpenAI configuration
type OpenAIConfig struct {
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	OrgID     string `json:"org_id"`
	Enabled   bool   `json:"enabled"`
}

// ClaudeConfig represents Claude configuration
type ClaudeConfig struct {
	APIKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	Enabled   bool   `json:"enabled"`
}

// N8NConfig represents n8n workflow configuration
type N8NConfig struct {
	BaseURL       string            `json:"base_url"`
	APIKey        string            `json:"api_key"`
	Workflows     map[string]string `json:"workflows"`
	Timeout       time.Duration     `json:"timeout"`
	MaxRetries    int               `json:"max_retries"`
	WebhookSecret string            `json:"webhook_secret"`
	Enabled       bool              `json:"enabled"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled         bool              `json:"enabled"`
	MetricsEndpoint string            `json:"metrics_endpoint"`
	LogLevel        string            `json:"log_level"`
	Alerting        *AlertingConfig   `json:"alerting"`
	Dashboards      map[string]string `json:"dashboards"`
	RetentionDays   int               `json:"retention_days"`
}

// AlertingConfig represents alerting configuration
type AlertingConfig struct {
	Enabled          bool              `json:"enabled"`
	SlackWebhook     string            `json:"slack_webhook"`
	EmailRecipients  []string          `json:"email_recipients"`
	SMSRecipients    []string          `json:"sms_recipients"`
	AlertThresholds  map[string]float64 `json:"alert_thresholds"`
	EscalationRules  []EscalationRule   `json:"escalation_rules"`
}

// EscalationRule represents an escalation rule
type EscalationRule struct {
	Severity      string        `json:"severity"`
	DelayMinutes  int           `json:"delay_minutes"`
	Recipients    []string      `json:"recipients"`
	MaxAttempts   int           `json:"max_attempts"`
	Enabled       bool          `json:"enabled"`
}

// ResourceLimits represents resource limits
type ResourceLimits struct {
	CPULimit      string `json:"cpu_limit"`
	MemoryLimit   string `json:"memory_limit"`
	StorageLimit  string `json:"storage_limit"`
	NetworkLimit  string `json:"network_limit"`
	MaxConcurrency int   `json:"max_concurrency"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	TLSConfig     *TLSConfig     `json:"tls_config"`
	AuthConfig    *AuthConfig    `json:"auth_config"`
	RateLimiting  *RateLimiting  `json:"rate_limiting"`
	IPWhitelist   []string       `json:"ip_whitelist"`
	IPBlacklist   []string       `json:"ip_blacklist"`
	CORSConfig    *CORSConfig    `json:"cors_config"`
	CSPConfig     *CSPConfig     `json:"csp_config"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled     bool   `json:"enabled"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	CAFile      string `json:"ca_file"`
	MinVersion  string `json:"min_version"`
	CipherSuites []string `json:"cipher_suites"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret     string        `json:"jwt_secret"`
	TokenExpiry   time.Duration `json:"token_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Issuer        string        `json:"issuer"`
	Audience      string        `json:"audience"`
}

// RateLimiting represents rate limiting configuration
type RateLimiting struct {
	Enabled       bool                  `json:"enabled"`
	GlobalLimit   RateLimit             `json:"global_limit"`
	EndpointLimits map[string]RateLimit `json:"endpoint_limits"`
	UserLimits    map[string]RateLimit  `json:"user_limits"`
}

// RateLimit represents a rate limit
type RateLimit struct {
	Requests   int           `json:"requests"`
	Duration   time.Duration `json:"duration"`
	BurstLimit int           `json:"burst_limit"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
	ExposedHeaders []string `json:"exposed_headers"`
	MaxAge         int      `json:"max_age"`
}

// CSPConfig represents Content Security Policy configuration
type CSPConfig struct {
	DefaultSrc []string `json:"default_src"`
	ScriptSrc  []string `json:"script_src"`
	StyleSrc   []string `json:"style_src"`
	ImgSrc     []string `json:"img_src"`
	FontSrc    []string `json:"font_src"`
	ConnectSrc []string `json:"connect_src"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level        string            `json:"level"`
	Format       string            `json:"format"`
	Output       string            `json:"output"`
	FilePath     string            `json:"file_path"`
	MaxSize      int               `json:"max_size"`
	MaxBackups   int               `json:"max_backups"`
	MaxAge       int               `json:"max_age"`
	Structured   bool              `json:"structured"`
	Fields       map[string]string `json:"fields"`
}

// ConfigurationSource represents a configuration source
type ConfigurationSource interface {
	GetConfiguration(environment string) (*EnvironmentConfig, error)
	GetGlobalConfiguration() (*GlobalConfig, error)
	Watch(callback func(config *EnvironmentConfig)) error
}

// ConfigurationChangeListener listens for configuration changes
type ConfigurationChangeListener interface {
	OnConfigurationChange(environment string, oldConfig, newConfig *EnvironmentConfig)
}

// SecretsManager manages secrets and sensitive configuration
type SecretsManager struct {
	mu      sync.RWMutex
	secrets map[string]string
	vault   VaultProvider
}

// VaultProvider interface for secret storage
type VaultProvider interface {
	GetSecret(key string) (string, error)
	SetSecret(key, value string) error
	DeleteSecret(key string) error
	ListSecrets() ([]string, error)
}

// Security defaults and other configuration types
type SecurityDefaults struct {
	RequireHTTPS     bool `json:"require_https"`
	RequireAuth      bool `json:"require_auth"`
	SessionTimeout   time.Duration `json:"session_timeout"`
	MaxLoginAttempts int  `json:"max_login_attempts"`
}

type LoggingDefaults struct {
	Level         string `json:"level"`
	Format        string `json:"format"`
	IncludeTrace  bool   `json:"include_trace"`
	SanitizeData  bool   `json:"sanitize_data"`
}

type MonitoringDefaults struct {
	MetricsEnabled   bool          `json:"metrics_enabled"`
	TracingEnabled   bool          `json:"tracing_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	AlertsEnabled    bool          `json:"alerts_enabled"`
}

type BackupConfig struct {
	Enabled       bool          `json:"enabled"`
	Interval      time.Duration `json:"interval"`
	RetentionDays int           `json:"retention_days"`
	StoragePath   string        `json:"storage_path"`
	Compression   bool          `json:"compression"`
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{
		environments:         make(map[string]*EnvironmentConfig),
		globalConfiguration:  &GlobalConfig{},
		secretsManager:       NewSecretsManager(),
		configurationSources: make([]ConfigurationSource, 0),
		changeListeners:      make([]ConfigurationChangeListener, 0),
		configurationCache:   make(map[string]interface{}),
		cacheExpiry:          5 * time.Minute,
		lastUpdate:           time.Now(),
	}
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager() *SecretsManager {
	return &SecretsManager{
		secrets: make(map[string]string),
	}
}

// LoadConfiguration loads configuration from environment variables and files
func (cm *ConfigurationManager) LoadConfiguration() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Load global configuration
	if err := cm.loadGlobalConfiguration(); err != nil {
		return fmt.Errorf("failed to load global configuration: %w", err)
	}

	// Load environment-specific configurations
	environments := []string{"development", "staging", "production"}
	for _, env := range environments {
		if err := cm.loadEnvironmentConfiguration(env); err != nil {
			return fmt.Errorf("failed to load configuration for environment %s: %w", env, err)
		}
	}

	// Load secrets
	if err := cm.loadSecrets(); err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	cm.lastUpdate = time.Now()
	return nil
}

// loadGlobalConfiguration loads global configuration settings
func (cm *ConfigurationManager) loadGlobalConfiguration() error {
	cm.globalConfiguration = &GlobalConfig{
		ApplicationName: getEnvOrDefault("APP_NAME", "DealDone"),
		Version:         getEnvOrDefault("APP_VERSION", "1.2.0"),
		DefaultTimeout:  getDurationEnvOrDefault("DEFAULT_TIMEOUT", 30*time.Second),
		MaxRetries:      getIntEnvOrDefault("MAX_RETRIES", 3),
		RateLimits:      make(map[string]RateLimit),
		DefaultFeatureFlags: map[string]bool{
			"ai_enhanced_analysis": true,
			"n8n_workflows":       true,
			"advanced_monitoring": true,
			"automatic_rollback":  true,
			"canary_deployment":   true,
		},
		SecurityDefaults: &SecurityDefaults{
			RequireHTTPS:     getBoolEnvOrDefault("REQUIRE_HTTPS", true),
			RequireAuth:      getBoolEnvOrDefault("REQUIRE_AUTH", true),
			SessionTimeout:   getDurationEnvOrDefault("SESSION_TIMEOUT", 24*time.Hour),
			MaxLoginAttempts: getIntEnvOrDefault("MAX_LOGIN_ATTEMPTS", 3),
		},
		LoggingDefaults: &LoggingDefaults{
			Level:        getEnvOrDefault("LOG_LEVEL", "info"),
			Format:       getEnvOrDefault("LOG_FORMAT", "json"),
			IncludeTrace: getBoolEnvOrDefault("LOG_INCLUDE_TRACE", false),
			SanitizeData: getBoolEnvOrDefault("LOG_SANITIZE_DATA", true),
		},
		MonitoringDefaults: &MonitoringDefaults{
			MetricsEnabled:      getBoolEnvOrDefault("METRICS_ENABLED", true),
			TracingEnabled:      getBoolEnvOrDefault("TRACING_ENABLED", true),
			HealthCheckInterval: getDurationEnvOrDefault("HEALTH_CHECK_INTERVAL", 30*time.Second),
			AlertsEnabled:       getBoolEnvOrDefault("ALERTS_ENABLED", true),
		},
		BackupConfig: &BackupConfig{
			Enabled:       getBoolEnvOrDefault("BACKUP_ENABLED", true),
			Interval:      getDurationEnvOrDefault("BACKUP_INTERVAL", 24*time.Hour),
			RetentionDays: getIntEnvOrDefault("BACKUP_RETENTION_DAYS", 30),
			StoragePath:   getEnvOrDefault("BACKUP_STORAGE_PATH", "/var/backups"),
			Compression:   getBoolEnvOrDefault("BACKUP_COMPRESSION", true),
		},
	}

	return nil
}

// loadEnvironmentConfiguration loads configuration for a specific environment
func (cm *ConfigurationManager) loadEnvironmentConfiguration(env string) error {
	envPrefix := strings.ToUpper(env) + "_"
	
	config := &EnvironmentConfig{
		Name:    env,
		Type:    EnvironmentType(env),
		Variables: make(map[string]string),
		Secrets:   make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:              getEnvOrDefault(envPrefix+"DB_HOST", "localhost"),
			Port:              getIntEnvOrDefault(envPrefix+"DB_PORT", 5432),
			Database:          getEnvOrDefault(envPrefix+"DB_NAME", "dealdone"),
			Username:          getEnvOrDefault(envPrefix+"DB_USER", "dealdone"),
			Password:          getEnvOrDefault(envPrefix+"DB_PASSWORD", ""),
			MaxConnections:    getIntEnvOrDefault(envPrefix+"DB_MAX_CONNECTIONS", 10),
			ConnectionTimeout: getDurationEnvOrDefault(envPrefix+"DB_TIMEOUT", 30*time.Second),
			SSLMode:          getEnvOrDefault(envPrefix+"DB_SSL_MODE", "require"),
			BackupEnabled:    getBoolEnvOrDefault(envPrefix+"DB_BACKUP_ENABLED", true),
		},
		AIProviderConfig: &AIProviderConfig{
			OpenAIConfig: &OpenAIConfig{
				APIKey:  getEnvOrDefault(envPrefix+"OPENAI_API_KEY", ""),
				Model:   getEnvOrDefault(envPrefix+"OPENAI_MODEL", "gpt-4"),
				BaseURL: getEnvOrDefault(envPrefix+"OPENAI_BASE_URL", "https://api.openai.com/v1"),
				OrgID:   getEnvOrDefault(envPrefix+"OPENAI_ORG_ID", ""),
				Enabled: getBoolEnvOrDefault(envPrefix+"OPENAI_ENABLED", true),
			},
			ClaudeConfig: &ClaudeConfig{
				APIKey:  getEnvOrDefault(envPrefix+"CLAUDE_API_KEY", ""),
				Model:   getEnvOrDefault(envPrefix+"CLAUDE_MODEL", "claude-3-opus-20240229"),
				BaseURL: getEnvOrDefault(envPrefix+"CLAUDE_BASE_URL", "https://api.anthropic.com"),
				Enabled: getBoolEnvOrDefault(envPrefix+"CLAUDE_ENABLED", true),
			},
			DefaultProvider: getEnvOrDefault(envPrefix+"AI_DEFAULT_PROVIDER", "openai"),
			MaxTokens:       getIntEnvOrDefault(envPrefix+"AI_MAX_TOKENS", 4096),
			Temperature:     getFloat64EnvOrDefault(envPrefix+"AI_TEMPERATURE", 0.7),
			Timeout:         getDurationEnvOrDefault(envPrefix+"AI_TIMEOUT", 60*time.Second),
			RateLimits:      make(map[string]int),
		},
		N8NConfig: &N8NConfig{
			BaseURL:       getEnvOrDefault(envPrefix+"N8N_BASE_URL", "http://localhost:5678"),
			APIKey:        getEnvOrDefault(envPrefix+"N8N_API_KEY", ""),
			Workflows:     make(map[string]string),
			Timeout:       getDurationEnvOrDefault(envPrefix+"N8N_TIMEOUT", 120*time.Second),
			MaxRetries:    getIntEnvOrDefault(envPrefix+"N8N_MAX_RETRIES", 3),
			WebhookSecret: getEnvOrDefault(envPrefix+"N8N_WEBHOOK_SECRET", ""),
			Enabled:       getBoolEnvOrDefault(envPrefix+"N8N_ENABLED", true),
		},
		MonitoringConfig: &MonitoringConfig{
			Enabled:         getBoolEnvOrDefault(envPrefix+"MONITORING_ENABLED", true),
			MetricsEndpoint: getEnvOrDefault(envPrefix+"METRICS_ENDPOINT", "/metrics"),
			LogLevel:        getEnvOrDefault(envPrefix+"LOG_LEVEL", "info"),
			Alerting: &AlertingConfig{
				Enabled:         getBoolEnvOrDefault(envPrefix+"ALERTING_ENABLED", true),
				SlackWebhook:    getEnvOrDefault(envPrefix+"SLACK_WEBHOOK", ""),
				EmailRecipients: getStringSliceEnv(envPrefix+"EMAIL_RECIPIENTS"),
				SMSRecipients:   getStringSliceEnv(envPrefix+"SMS_RECIPIENTS"),
				AlertThresholds: make(map[string]float64),
				EscalationRules: make([]EscalationRule, 0),
			},
			Dashboards:    make(map[string]string),
			RetentionDays: getIntEnvOrDefault(envPrefix+"METRICS_RETENTION_DAYS", 30),
		},
		FeatureFlags: map[string]bool{
			"ai_enhanced_analysis": getBoolEnvOrDefault(envPrefix+"FF_AI_ENHANCED_ANALYSIS", true),
			"n8n_workflows":       getBoolEnvOrDefault(envPrefix+"FF_N8N_WORKFLOWS", true),
			"advanced_monitoring": getBoolEnvOrDefault(envPrefix+"FF_ADVANCED_MONITORING", true),
			"automatic_rollback":  getBoolEnvOrDefault(envPrefix+"FF_AUTOMATIC_ROLLBACK", true),
			"canary_deployment":   getBoolEnvOrDefault(envPrefix+"FF_CANARY_DEPLOYMENT", false),
		},
		ResourceLimits: &ResourceLimits{
			CPULimit:       getEnvOrDefault(envPrefix+"CPU_LIMIT", "1000m"),
			MemoryLimit:    getEnvOrDefault(envPrefix+"MEMORY_LIMIT", "2Gi"),
			StorageLimit:   getEnvOrDefault(envPrefix+"STORAGE_LIMIT", "10Gi"),
			NetworkLimit:   getEnvOrDefault(envPrefix+"NETWORK_LIMIT", "1Gbps"),
			MaxConcurrency: getIntEnvOrDefault(envPrefix+"MAX_CONCURRENCY", 100),
		},
		SecurityConfig: &SecurityConfig{
			TLSConfig: &TLSConfig{
				Enabled:      getBoolEnvOrDefault(envPrefix+"TLS_ENABLED", true),
				CertFile:     getEnvOrDefault(envPrefix+"TLS_CERT_FILE", ""),
				KeyFile:      getEnvOrDefault(envPrefix+"TLS_KEY_FILE", ""),
				CAFile:       getEnvOrDefault(envPrefix+"TLS_CA_FILE", ""),
				MinVersion:   getEnvOrDefault(envPrefix+"TLS_MIN_VERSION", "1.2"),
				CipherSuites: getStringSliceEnv(envPrefix+"TLS_CIPHER_SUITES"),
			},
			AuthConfig: &AuthConfig{
				JWTSecret:     getEnvOrDefault(envPrefix+"JWT_SECRET", ""),
				TokenExpiry:   getDurationEnvOrDefault(envPrefix+"TOKEN_EXPIRY", 1*time.Hour),
				RefreshExpiry: getDurationEnvOrDefault(envPrefix+"REFRESH_EXPIRY", 24*time.Hour),
				Issuer:        getEnvOrDefault(envPrefix+"JWT_ISSUER", "dealdone"),
				Audience:      getEnvOrDefault(envPrefix+"JWT_AUDIENCE", "dealdone-api"),
			},
			RateLimiting: &RateLimiting{
				Enabled: getBoolEnvOrDefault(envPrefix+"RATE_LIMITING_ENABLED", true),
				GlobalLimit: RateLimit{
					Requests:   getIntEnvOrDefault(envPrefix+"RATE_LIMIT_REQUESTS", 100),
					Duration:   getDurationEnvOrDefault(envPrefix+"RATE_LIMIT_DURATION", 1*time.Minute),
					BurstLimit: getIntEnvOrDefault(envPrefix+"RATE_LIMIT_BURST", 150),
				},
				EndpointLimits: make(map[string]RateLimit),
				UserLimits:     make(map[string]RateLimit),
			},
			IPWhitelist: getStringSliceEnv(envPrefix+"IP_WHITELIST"),
			IPBlacklist: getStringSliceEnv(envPrefix+"IP_BLACKLIST"),
			CORSConfig: &CORSConfig{
				AllowedOrigins: getStringSliceEnv(envPrefix+"CORS_ALLOWED_ORIGINS"),
				AllowedMethods: getStringSliceEnv(envPrefix+"CORS_ALLOWED_METHODS"),
				AllowedHeaders: getStringSliceEnv(envPrefix+"CORS_ALLOWED_HEADERS"),
				ExposedHeaders: getStringSliceEnv(envPrefix+"CORS_EXPOSED_HEADERS"),
				MaxAge:         getIntEnvOrDefault(envPrefix+"CORS_MAX_AGE", 86400),
			},
		},
		LoggingConfig: &LoggingConfig{
			Level:      getEnvOrDefault(envPrefix+"LOG_LEVEL", "info"),
			Format:     getEnvOrDefault(envPrefix+"LOG_FORMAT", "json"),
			Output:     getEnvOrDefault(envPrefix+"LOG_OUTPUT", "stdout"),
			FilePath:   getEnvOrDefault(envPrefix+"LOG_FILE_PATH", ""),
			MaxSize:    getIntEnvOrDefault(envPrefix+"LOG_MAX_SIZE", 100),
			MaxBackups: getIntEnvOrDefault(envPrefix+"LOG_MAX_BACKUPS", 5),
			MaxAge:     getIntEnvOrDefault(envPrefix+"LOG_MAX_AGE", 30),
			Structured: getBoolEnvOrDefault(envPrefix+"LOG_STRUCTURED", true),
			Fields:     make(map[string]string),
		},
		Version:     getEnvOrDefault(envPrefix+"VERSION", "1.2.0"),
		LastUpdated: time.Now(),
	}

	// Load N8N workflow configurations
	config.N8NConfig.Workflows = map[string]string{
		"enhanced_analyze_all":     getEnvOrDefault(envPrefix+"N8N_WORKFLOW_ENHANCED_ANALYZE_ALL", "enhanced-analyze-all"),
		"entity_extraction":        getEnvOrDefault(envPrefix+"N8N_WORKFLOW_ENTITY_EXTRACTION", "entity-extraction-specialist"),
		"financial_validation":     getEnvOrDefault(envPrefix+"N8N_WORKFLOW_FINANCIAL_VALIDATION", "financial-data-validator"),
		"template_quality":         getEnvOrDefault(envPrefix+"N8N_WORKFLOW_TEMPLATE_QUALITY", "template-quality-assessor"),
		"error_recovery":           getEnvOrDefault(envPrefix+"N8N_WORKFLOW_ERROR_RECOVERY", "error-recovery-handler"),
		"document_classification":  getEnvOrDefault(envPrefix+"N8N_WORKFLOW_DOCUMENT_CLASSIFICATION", "document-classification-routing"),
		"template_discovery":       getEnvOrDefault(envPrefix+"N8N_WORKFLOW_TEMPLATE_DISCOVERY", "template-discovery-mapping"),
		"result_aggregation":       getEnvOrDefault(envPrefix+"N8N_WORKFLOW_RESULT_AGGREGATION", "result-aggregation-notifications"),
	}

	// Load AI provider rate limits
	config.AIProviderConfig.RateLimits = map[string]int{
		"openai_rpm":    getIntEnvOrDefault(envPrefix+"OPENAI_RPM", 3000),
		"openai_tpm":    getIntEnvOrDefault(envPrefix+"OPENAI_TPM", 40000),
		"claude_rpm":    getIntEnvOrDefault(envPrefix+"CLAUDE_RPM", 1000),
		"claude_tpm":    getIntEnvOrDefault(envPrefix+"CLAUDE_TPM", 20000),
	}

	// Load alert thresholds
	config.MonitoringConfig.Alerting.AlertThresholds = map[string]float64{
		"error_rate":           getFloat64EnvOrDefault(envPrefix+"ALERT_ERROR_RATE", 5.0),
		"response_time":        getFloat64EnvOrDefault(envPrefix+"ALERT_RESPONSE_TIME", 5000.0),
		"cpu_usage":            getFloat64EnvOrDefault(envPrefix+"ALERT_CPU_USAGE", 80.0),
		"memory_usage":         getFloat64EnvOrDefault(envPrefix+"ALERT_MEMORY_USAGE", 85.0),
		"disk_usage":           getFloat64EnvOrDefault(envPrefix+"ALERT_DISK_USAGE", 90.0),
		"queue_depth":          getFloat64EnvOrDefault(envPrefix+"ALERT_QUEUE_DEPTH", 100.0),
		"failed_deployments":   getFloat64EnvOrDefault(envPrefix+"ALERT_FAILED_DEPLOYMENTS", 2.0),
	}

	// Load dashboard configurations
	config.MonitoringConfig.Dashboards = map[string]string{
		"executive":     getEnvOrDefault(envPrefix+"DASHBOARD_EXECUTIVE_URL", ""),
		"operational":   getEnvOrDefault(envPrefix+"DASHBOARD_OPERATIONAL_URL", ""),
		"technical":     getEnvOrDefault(envPrefix+"DASHBOARD_TECHNICAL_URL", ""),
		"business":      getEnvOrDefault(envPrefix+"DASHBOARD_BUSINESS_URL", ""),
	}

	cm.environments[env] = config
	return nil
}

// loadSecrets loads secrets from environment variables or vault
func (cm *ConfigurationManager) loadSecrets() error {
	// Load secrets from environment variables
	secrets := map[string]string{
		"db_password":        os.Getenv("DB_PASSWORD"),
		"openai_api_key":     os.Getenv("OPENAI_API_KEY"),
		"claude_api_key":     os.Getenv("CLAUDE_API_KEY"),
		"n8n_api_key":        os.Getenv("N8N_API_KEY"),
		"jwt_secret":         os.Getenv("JWT_SECRET"),
		"webhook_secret":     os.Getenv("WEBHOOK_SECRET"),
		"slack_webhook":      os.Getenv("SLACK_WEBHOOK_URL"),
		"email_password":     os.Getenv("EMAIL_PASSWORD"),
		"tls_cert":           os.Getenv("TLS_CERT_FILE"),
		"tls_key":            os.Getenv("TLS_KEY_FILE"),
	}

	for key, value := range secrets {
		if value != "" {
			cm.secretsManager.SetSecret(key, value)
		}
	}

	return nil
}

// GetEnvironmentConfig returns configuration for a specific environment
func (cm *ConfigurationManager) GetEnvironmentConfig(environment string) map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config, exists := cm.environments[environment]
	if !exists {
		return make(map[string]interface{})
	}

	// Convert to map for easier use
	configMap := make(map[string]interface{})
	configBytes, _ := json.Marshal(config)
	json.Unmarshal(configBytes, &configMap)

	return configMap
}

// GetConfiguration returns configuration for a specific environment
func (cm *ConfigurationManager) GetConfiguration(environment string) (*EnvironmentConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config, exists := cm.environments[environment]
	if !exists {
		return nil, fmt.Errorf("configuration for environment %s not found", environment)
	}

	return config, nil
}

// GetGlobalConfiguration returns global configuration
func (cm *ConfigurationManager) GetGlobalConfiguration() *GlobalConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.globalConfiguration
}

// UpdateConfiguration updates configuration for a specific environment
func (cm *ConfigurationManager) UpdateConfiguration(environment string, config *EnvironmentConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	oldConfig := cm.environments[environment]
	config.LastUpdated = time.Now()
	cm.environments[environment] = config

	// Notify listeners
	for _, listener := range cm.changeListeners {
		listener.OnConfigurationChange(environment, oldConfig, config)
	}

	return nil
}

// GetSecret retrieves a secret value
func (cm *ConfigurationManager) GetSecret(key string) (string, error) {
	return cm.secretsManager.GetSecret(key)
}

// SetSecret sets a secret value
func (cm *ConfigurationManager) SetSecret(key, value string) error {
	return cm.secretsManager.SetSecret(key, value)
}

// AddConfigurationSource adds a configuration source
func (cm *ConfigurationManager) AddConfigurationSource(source ConfigurationSource) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.configurationSources = append(cm.configurationSources, source)
}

// AddChangeListener adds a configuration change listener
func (cm *ConfigurationManager) AddChangeListener(listener ConfigurationChangeListener) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.changeListeners = append(cm.changeListeners, listener)
}

// ValidateConfiguration validates configuration settings
func (cm *ConfigurationManager) ValidateConfiguration(environment string) error {
	config, err := cm.GetConfiguration(environment)
	if err != nil {
		return err
	}

	// Validate database configuration
	if config.DatabaseConfig.Host == "" {
		return fmt.Errorf("database host is required")
	}

	// Validate AI provider configuration
	if config.AIProviderConfig.OpenAIConfig.APIKey == "" && config.AIProviderConfig.ClaudeConfig.APIKey == "" {
		return fmt.Errorf("at least one AI provider API key is required")
	}

	// Validate N8N configuration
	if config.N8NConfig.Enabled && config.N8NConfig.BaseURL == "" {
		return fmt.Errorf("N8N base URL is required when N8N is enabled")
	}

	// Validate monitoring configuration
	if config.MonitoringConfig.Enabled && config.MonitoringConfig.MetricsEndpoint == "" {
		return fmt.Errorf("metrics endpoint is required when monitoring is enabled")
	}

	return nil
}

// GetFeatureFlag returns the value of a feature flag
func (cm *ConfigurationManager) GetFeatureFlag(environment, flag string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config, exists := cm.environments[environment]
	if !exists {
		return false
	}

	value, exists := config.FeatureFlags[flag]
	if !exists {
		// Check global defaults
		if cm.globalConfiguration != nil {
			if defaultValue, exists := cm.globalConfiguration.DefaultFeatureFlags[flag]; exists {
				return defaultValue
			}
		}
		return false
	}

	return value
}

// SetFeatureFlag sets the value of a feature flag
func (cm *ConfigurationManager) SetFeatureFlag(environment, flag string, value bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.environments[environment]
	if !exists {
		return fmt.Errorf("environment %s not found", environment)
	}

	config.FeatureFlags[flag] = value
	config.LastUpdated = time.Now()

	return nil
}

// GetAllFeatureFlags returns all feature flags for an environment
func (cm *ConfigurationManager) GetAllFeatureFlags(environment string) map[string]bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config, exists := cm.environments[environment]
	if !exists {
		return make(map[string]bool)
	}

	return config.FeatureFlags
}

// SecretsManager methods
func (sm *SecretsManager) GetSecret(key string) (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	value, exists := sm.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret %s not found", key)
	}

	return value, nil
}

func (sm *SecretsManager) SetSecret(key, value string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.secrets[key] = value
	return nil
}

func (sm *SecretsManager) DeleteSecret(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.secrets, key)
	return nil
}

func (sm *SecretsManager) ListSecrets() ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	keys := make([]string, 0, len(sm.secrets))
	for key := range sm.secrets {
		keys = append(keys, key)
	}

	return keys, nil
}

// Utility functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseIntSafely(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloat64EnvOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := parseFloat64Safely(value); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return make([]string, 0)
}

// Safe parsing functions
func parseIntSafely(value string) (int, error) {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	return result, err
}

func parseFloat64Safely(value string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	return result, err
}
``` 