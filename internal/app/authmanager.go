package app

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuthManager handles authentication and API key management
type AuthManager struct {
	keys           map[string]*APIKeyInfo
	keyStorage     string
	mu             sync.RWMutex
	config         *AuthConfig
	rateLimiter    *RateLimiter
	auditLogger    *AuditLogger
	encryptionKey  []byte
	lastRotation   time.Time
	rotationTicker *time.Ticker
	stopRotation   chan bool
}

// APIKeyInfo contains information about an API key
type APIKeyInfo struct {
	KeyID           string                 `json:"keyId"`
	KeyHash         string                 `json:"keyHash"` // SHA256 hash of the actual key
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Permissions     []string               `json:"permissions"`
	CreatedAt       time.Time              `json:"createdAt"`
	ExpiresAt       *time.Time             `json:"expiresAt,omitempty"`
	LastUsed        *time.Time             `json:"lastUsed,omitempty"`
	UsageCount      int64                  `json:"usageCount"`
	IsActive        bool                   `json:"isActive"`
	RateLimitTier   string                 `json:"rateLimitTier"` // "basic", "premium", "unlimited"
	Tags            []string               `json:"tags,omitempty"`
	IPWhitelist     []string               `json:"ipWhitelist,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	RotationHistory []KeyRotationEvent     `json:"rotationHistory,omitempty"`
}

// KeyRotationEvent tracks key rotation events
type KeyRotationEvent struct {
	EventID     string    `json:"eventId"`
	OldKeyHash  string    `json:"oldKeyHash"`
	NewKeyHash  string    `json:"newKeyHash"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	InitiatedBy string    `json:"initiatedBy"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	DefaultExpiration   time.Duration     `json:"defaultExpiration"`
	MaxKeyLifetime      time.Duration     `json:"maxKeyLifetime"`
	RequireHMAC         bool              `json:"requireHMAC"`
	MinKeyLength        int               `json:"minKeyLength"`
	EnableAutoRotation  bool              `json:"enableAutoRotation"`
	RotationInterval    time.Duration     `json:"rotationInterval"`
	MaxKeysPerUser      int               `json:"maxKeysPerUser"`
	EnableAuditLogging  bool              `json:"enableAuditLogging"`
	HashAlgorithm       string            `json:"hashAlgorithm"`
	EncryptionAlgorithm string            `json:"encryptionAlgorithm"`
	RateLimitEnabled    bool              `json:"rateLimitEnabled"`
	SecurityPolicies    *SecurityPolicies `json:"securityPolicies"`
}

// SecurityPolicies defines security enforcement policies
type SecurityPolicies struct {
	RequireIPWhitelist     bool          `json:"requireIpWhitelist"`
	MaxFailedAttempts      int           `json:"maxFailedAttempts"`
	LockoutDuration        time.Duration `json:"lockoutDuration"`
	RequireStrongKeys      bool          `json:"requireStrongKeys"`
	DisallowWeakAlgorithms bool          `json:"disallowWeakAlgorithms"`
	EnableKeyBlacklisting  bool          `json:"enableKeyBlacklisting"`
	RequireCertificates    bool          `json:"requireCertificates"`
	EnableMutualTLS        bool          `json:"enableMutualTls"`
}

// AuthenticationRequest represents an authentication request
type AuthenticationRequest struct {
	APIKey    string            `json:"apiKey,omitempty"`
	Signature string            `json:"signature,omitempty"`
	Timestamp int64             `json:"timestamp"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Body      string            `json:"body,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	ClientIP  string            `json:"clientIp,omitempty"`
	UserAgent string            `json:"userAgent,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

// AuthenticationResult represents the result of an authentication attempt
type AuthenticationResult struct {
	Success       bool                   `json:"success"`
	KeyID         string                 `json:"keyId,omitempty"`
	Permissions   []string               `json:"permissions,omitempty"`
	RateLimitTier string                 `json:"rateLimitTier,omitempty"`
	ExpiresAt     *time.Time             `json:"expiresAt,omitempty"`
	ErrorCode     string                 `json:"errorCode,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	AuditID       string                 `json:"auditId,omitempty"`
}

// KeyGenerationRequest represents a request to generate a new API key
type KeyGenerationRequest struct {
	Name           string                 `json:"name" validate:"required"`
	Description    string                 `json:"description"`
	Permissions    []string               `json:"permissions"`
	ExpirationDays int                    `json:"expirationDays,omitempty"`
	RateLimitTier  string                 `json:"rateLimitTier"`
	Tags           []string               `json:"tags,omitempty"`
	IPWhitelist    []string               `json:"ipWhitelist,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	AutoRotate     bool                   `json:"autoRotate"`
}

// KeyGenerationResult represents the result of key generation
type KeyGenerationResult struct {
	KeyID       string      `json:"keyId"`
	APIKey      string      `json:"apiKey"` // Only returned once during generation
	KeyInfo     *APIKeyInfo `json:"keyInfo"`
	ExpiresAt   *time.Time  `json:"expiresAt,omitempty"`
	GeneratedAt time.Time   `json:"generatedAt"`
}

// AuditEvent represents an authentication audit event
type AuditEvent struct {
	EventID   string                 `json:"eventId"`
	EventType string                 `json:"eventType"` // "auth_success", "auth_failure", "key_created", etc.
	KeyID     string                 `json:"keyId,omitempty"`
	ClientIP  string                 `json:"clientIp,omitempty"`
	UserAgent string                 `json:"userAgent,omitempty"`
	RequestID string                 `json:"requestId,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  string                 `json:"severity"` // "info", "warning", "error", "critical"
}

// AuditLogger handles audit logging for authentication events
type AuditLogger struct {
	logFile     *os.File
	mu          sync.Mutex
	enabled     bool
	maxFileSize int64
	retention   time.Duration
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(keyStoragePath string, config *AuthConfig) (*AuthManager, error) {
	if config == nil {
		config = &AuthConfig{
			DefaultExpiration:   24 * time.Hour * 30,  // 30 days
			MaxKeyLifetime:      24 * time.Hour * 365, // 1 year
			RequireHMAC:         true,
			MinKeyLength:        32,
			EnableAutoRotation:  false,
			RotationInterval:    24 * time.Hour * 30, // 30 days
			MaxKeysPerUser:      10,
			EnableAuditLogging:  true,
			RateLimitEnabled:    true,
			HashAlgorithm:       "SHA256",
			EncryptionAlgorithm: "AES-256-GCM",
			SecurityPolicies: &SecurityPolicies{
				RequireIPWhitelist:     false,
				MaxFailedAttempts:      5,
				LockoutDuration:        15 * time.Minute,
				RequireStrongKeys:      true,
				DisallowWeakAlgorithms: true,
				EnableKeyBlacklisting:  true,
			},
		}
	}

	// Generate encryption key for sensitive data
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Initialize audit logger
	var auditLogger *AuditLogger
	if config.EnableAuditLogging {
		auditLogPath := filepath.Join(filepath.Dir(keyStoragePath), "auth_audit.log")
		logger, err := NewAuditLogger(auditLogPath)
		if err != nil {
			log.Printf("Warning: Failed to initialize audit logger: %v", err)
		} else {
			auditLogger = logger
		}
	}

	// Initialize rate limiter
	var rateLimiter *RateLimiter
	if config.RateLimitEnabled {
		rateLimiter = NewRateLimiter(60) // 60 requests per minute default
	}

	am := &AuthManager{
		keys:          make(map[string]*APIKeyInfo),
		keyStorage:    keyStoragePath,
		config:        config,
		rateLimiter:   rateLimiter,
		auditLogger:   auditLogger,
		encryptionKey: encryptionKey,
		stopRotation:  make(chan bool),
	}

	// Load existing keys
	if err := am.loadKeys(); err != nil {
		log.Printf("Warning: Failed to load existing keys: %v", err)
	}

	// Start auto-rotation if enabled
	if config.EnableAutoRotation {
		am.startAutoRotation()
	}

	return am, nil
}

// GenerateAPIKey generates a new API key
func (am *AuthManager) GenerateAPIKey(req *KeyGenerationRequest) (*KeyGenerationResult, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Validate request
	if err := am.validateKeyGenerationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid key generation request: %w", err)
	}

	// Generate secure random key
	keyBytes := make([]byte, max(am.config.MinKeyLength, 32))
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	apiKey := base64.URLEncoding.EncodeToString(keyBytes)
	keyID := am.generateKeyID()
	keyHash := am.hashKey(apiKey)

	// Set expiration
	var expiresAt *time.Time
	if req.ExpirationDays > 0 {
		expiration := time.Now().Add(time.Duration(req.ExpirationDays) * 24 * time.Hour)
		expiresAt = &expiration
	} else if am.config.DefaultExpiration > 0 {
		expiration := time.Now().Add(am.config.DefaultExpiration)
		expiresAt = &expiration
	}

	// Create key info
	keyInfo := &APIKeyInfo{
		KeyID:         keyID,
		KeyHash:       keyHash,
		Name:          req.Name,
		Description:   req.Description,
		Permissions:   req.Permissions,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		IsActive:      true,
		RateLimitTier: req.RateLimitTier,
		Tags:          req.Tags,
		IPWhitelist:   req.IPWhitelist,
		Metadata:      req.Metadata,
		UsageCount:    0,
	}

	// Store key info
	am.keys[keyID] = keyInfo

	// Save to storage
	if err := am.saveKeys(); err != nil {
		delete(am.keys, keyID)
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	// Log key creation
	if am.auditLogger != nil {
		am.auditLogger.LogEvent(&AuditEvent{
			EventID:   am.generateEventID(),
			EventType: "key_created",
			KeyID:     keyID,
			Details: map[string]interface{}{
				"name":        req.Name,
				"permissions": req.Permissions,
				"tier":        req.RateLimitTier,
			},
			Timestamp: time.Now(),
			Severity:  "info",
		})
	}

	return &KeyGenerationResult{
		KeyID:       keyID,
		APIKey:      apiKey, // Only returned here, never stored
		KeyInfo:     keyInfo,
		ExpiresAt:   expiresAt,
		GeneratedAt: time.Now(),
	}, nil
}

// AuthenticateRequest authenticates an incoming request
func (am *AuthManager) AuthenticateRequest(req *AuthenticationRequest) (*AuthenticationResult, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := &AuthenticationResult{
		Success: false,
	}

	// Validate request
	if req.APIKey == "" {
		result.ErrorCode = "MISSING_API_KEY"
		result.ErrorMessage = "API key is required"
		am.logAuthEvent("auth_failure", "", req, result)
		return result, nil
	}

	// Find key by hash
	keyHash := am.hashKey(req.APIKey)
	var keyInfo *APIKeyInfo
	var keyID string

	for id, info := range am.keys {
		if info.KeyHash == keyHash {
			keyInfo = info
			keyID = id
			break
		}
	}

	if keyInfo == nil {
		result.ErrorCode = "INVALID_API_KEY"
		result.ErrorMessage = "Invalid API key"
		am.logAuthEvent("auth_failure", "", req, result)
		return result, nil
	}

	// Check if key is active
	if !keyInfo.IsActive {
		result.ErrorCode = "KEY_INACTIVE"
		result.ErrorMessage = "API key is inactive"
		result.KeyID = keyID
		am.logAuthEvent("auth_failure", keyID, req, result)
		return result, nil
	}

	// Check expiration
	if keyInfo.ExpiresAt != nil && time.Now().After(*keyInfo.ExpiresAt) {
		result.ErrorCode = "KEY_EXPIRED"
		result.ErrorMessage = "API key has expired"
		result.KeyID = keyID
		am.logAuthEvent("auth_failure", keyID, req, result)
		return result, nil
	}

	// Check IP whitelist
	if len(keyInfo.IPWhitelist) > 0 && req.ClientIP != "" {
		allowed := false
		for _, ip := range keyInfo.IPWhitelist {
			if ip == req.ClientIP {
				allowed = true
				break
			}
		}
		if !allowed {
			result.ErrorCode = "IP_NOT_WHITELISTED"
			result.ErrorMessage = "Client IP not in whitelist"
			result.KeyID = keyID
			am.logAuthEvent("auth_failure", keyID, req, result)
			return result, nil
		}
	}

	// Check rate limiting
	if am.rateLimiter != nil && am.config.RateLimitEnabled {
		if !am.rateLimiter.TryAcquire() {
			result.ErrorCode = "RATE_LIMIT_EXCEEDED"
			result.ErrorMessage = "Rate limit exceeded"
			result.KeyID = keyID
			am.logAuthEvent("rate_limit_exceeded", keyID, req, result)
			return result, nil
		}
	}

	// Verify HMAC signature if required
	if am.config.RequireHMAC && req.Signature != "" {
		if !am.verifyHMACSignature(req, req.APIKey) {
			result.ErrorCode = "INVALID_SIGNATURE"
			result.ErrorMessage = "Invalid HMAC signature"
			result.KeyID = keyID
			am.logAuthEvent("auth_failure", keyID, req, result)
			return result, nil
		}
	}

	// Authentication successful
	result.Success = true
	result.KeyID = keyID
	result.Permissions = keyInfo.Permissions
	result.RateLimitTier = keyInfo.RateLimitTier
	result.ExpiresAt = keyInfo.ExpiresAt

	// Update usage statistics
	am.updateKeyUsage(keyID)

	// Log successful authentication
	am.logAuthEvent("auth_success", keyID, req, result)

	return result, nil
}

// RevokeAPIKey revokes an API key
func (am *AuthManager) RevokeAPIKey(keyID string, reason string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	keyInfo, exists := am.keys[keyID]
	if !exists {
		return fmt.Errorf("key not found: %s", keyID)
	}

	keyInfo.IsActive = false

	if err := am.saveKeys(); err != nil {
		return fmt.Errorf("failed to save key revocation: %w", err)
	}

	// Log key revocation
	if am.auditLogger != nil {
		am.auditLogger.LogEvent(&AuditEvent{
			EventID:   am.generateEventID(),
			EventType: "key_revoked",
			KeyID:     keyID,
			Details: map[string]interface{}{
				"reason": reason,
			},
			Timestamp: time.Now(),
			Severity:  "warning",
		})
	}

	return nil
}

// ListAPIKeys returns a list of all API keys (without the actual key values)
func (am *AuthManager) ListAPIKeys() ([]*APIKeyInfo, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keys := make([]*APIKeyInfo, 0, len(am.keys))
	for _, keyInfo := range am.keys {
		// Create a copy to avoid exposing internal data
		keyCopy := *keyInfo
		keys = append(keys, &keyCopy)
	}

	return keys, nil
}

// GetAPIKeyInfo returns information about a specific API key
func (am *AuthManager) GetAPIKeyInfo(keyID string) (*APIKeyInfo, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keyInfo, exists := am.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// Return a copy
	keyCopy := *keyInfo
	return &keyCopy, nil
}

// Helper methods

func (am *AuthManager) validateKeyGenerationRequest(req *KeyGenerationRequest) error {
	if req.Name == "" {
		return fmt.Errorf("key name is required")
	}

	if req.RateLimitTier == "" {
		req.RateLimitTier = "basic"
	}

	validTiers := []string{"basic", "premium", "unlimited"}
	tierValid := false
	for _, tier := range validTiers {
		if req.RateLimitTier == tier {
			tierValid = true
			break
		}
	}
	if !tierValid {
		return fmt.Errorf("invalid rate limit tier: %s", req.RateLimitTier)
	}

	return nil
}

func (am *AuthManager) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func (am *AuthManager) generateKeyID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("key_%x", bytes)
}

func (am *AuthManager) generateEventID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("evt_%x", bytes)
}

func (am *AuthManager) verifyHMACSignature(req *AuthenticationRequest, key string) bool {
	// Create message to sign
	message := fmt.Sprintf("%s|%s|%d|%s", req.Method, req.Path, req.Timestamp, req.Body)

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(req.Signature), []byte(expectedSignature))
}

func (am *AuthManager) updateKeyUsage(keyID string) {
	keyInfo := am.keys[keyID]
	if keyInfo != nil {
		keyInfo.UsageCount++
		now := time.Now()
		keyInfo.LastUsed = &now
	}
}

func (am *AuthManager) logAuthEvent(eventType, keyID string, req *AuthenticationRequest, result *AuthenticationResult) {
	if am.auditLogger == nil {
		return
	}

	severity := "info"
	if !result.Success {
		severity = "warning"
	}

	am.auditLogger.LogEvent(&AuditEvent{
		EventID:   am.generateEventID(),
		EventType: eventType,
		KeyID:     keyID,
		ClientIP:  req.ClientIP,
		UserAgent: req.UserAgent,
		RequestID: req.RequestID,
		Details: map[string]interface{}{
			"method":    req.Method,
			"path":      req.Path,
			"errorCode": result.ErrorCode,
			"success":   result.Success,
		},
		Timestamp: time.Now(),
		Severity:  severity,
	})
}

func (am *AuthManager) loadKeys() error {
	if _, err := os.Stat(am.keyStorage); os.IsNotExist(err) {
		// File doesn't exist, start with empty key store
		return nil
	}

	data, err := os.ReadFile(am.keyStorage)
	if err != nil {
		return fmt.Errorf("failed to read key storage file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var keys map[string]*APIKeyInfo
	if err := json.Unmarshal(data, &keys); err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	am.keys = keys
	return nil
}

func (am *AuthManager) saveKeys() error {
	// Ensure directory exists
	dir := filepath.Dir(am.keyStorage)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key storage directory: %w", err)
	}

	data, err := json.MarshalIndent(am.keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %w", err)
	}

	if err := os.WriteFile(am.keyStorage, data, 0600); err != nil {
		return fmt.Errorf("failed to write key storage file: %w", err)
	}

	return nil
}

func (am *AuthManager) startAutoRotation() {
	am.rotationTicker = time.NewTicker(am.config.RotationInterval)
	go func() {
		for {
			select {
			case <-am.rotationTicker.C:
				am.performAutoRotation()
			case <-am.stopRotation:
				am.rotationTicker.Stop()
				return
			}
		}
	}()
}

func (am *AuthManager) performAutoRotation() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	rotationThreshold := now.Add(-am.config.RotationInterval)

	for keyID, keyInfo := range am.keys {
		if keyInfo.IsActive && keyInfo.CreatedAt.Before(rotationThreshold) {
			// Note: Actual rotation would be implemented separately
			log.Printf("Key %s eligible for rotation", keyID)
		}
	}

	am.lastRotation = now
}

// Cleanup resources
func (am *AuthManager) Close() error {
	if am.rotationTicker != nil {
		am.stopRotation <- true
		am.rotationTicker.Stop()
	}

	if am.auditLogger != nil {
		return am.auditLogger.Close()
	}

	return nil
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		logFile:     logFile,
		enabled:     true,
		maxFileSize: 100 * 1024 * 1024,   // 100MB
		retention:   90 * 24 * time.Hour, // 90 days
	}, nil
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event *AuditEvent) error {
	if !al.enabled || al.logFile == nil {
		return nil
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	line := fmt.Sprintf("%s\n", string(eventJSON))
	if _, err := al.logFile.WriteString(line); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return al.logFile.Sync()
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.logFile != nil {
		return al.logFile.Close()
	}
	return nil
}

// Note: max function is available from utils.go
