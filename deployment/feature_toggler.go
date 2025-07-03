package deployment

import (
	"sync"
	"time"
)

// FeatureToggler controls feature rollout with percentage-based deployment
type FeatureToggler struct {
	mu           sync.RWMutex
	features     map[string]*Feature
	userFeatures map[string]map[string]bool
	rolloutRules []RolloutRule
	lastUpdate   time.Time
}

// Feature represents a feature flag configuration
type Feature struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Enabled      bool               `json:"enabled"`
	Percentage   float64            `json:"percentage"`
	Environment  string             `json:"environment"`
	Version      string             `json:"version"`
	RolloutRules []RolloutRule      `json:"rollout_rules"`
	Conditions   []FeatureCondition `json:"conditions"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Owner        string             `json:"owner"`
	Tags         []string           `json:"tags"`
}

// RolloutRule defines rules for feature rollout
type RolloutRule struct {
	Type        RolloutType `json:"type"`
	Percentage  float64     `json:"percentage"`
	UserGroups  []string    `json:"user_groups"`
	Environment string      `json:"environment"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Enabled     bool        `json:"enabled"`
}

// FeatureCondition defines conditions for feature activation
type FeatureCondition struct {
	Type     ConditionType `json:"type"`
	Operator string        `json:"operator"`
	Value    interface{}   `json:"value"`
	Field    string        `json:"field"`
}

// RolloutType represents the type of rollout
type RolloutType string

const (
	RolloutTypePercentage  RolloutType = "percentage"
	RolloutTypeUserGroup   RolloutType = "user_group"
	RolloutTypeEnvironment RolloutType = "environment"
	RolloutTypeTime        RolloutType = "time"
)

// NewFeatureToggler creates a new feature toggler instance
func NewFeatureToggler() *FeatureToggler {
	return &FeatureToggler{
		features:     make(map[string]*Feature),
		userFeatures: make(map[string]map[string]bool),
		rolloutRules: make([]RolloutRule, 0),
		lastUpdate:   time.Now(),
	}
}

// IsFeatureEnabled checks if a feature is enabled for a user
func (ft *FeatureToggler) IsFeatureEnabled(featureName, userID string) bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	feature, exists := ft.features[featureName]
	if !exists {
		return false
	}

	if !feature.Enabled {
		return false
	}

	// Check user-specific override
	if userFeatures, exists := ft.userFeatures[userID]; exists {
		if enabled, exists := userFeatures[featureName]; exists {
			return enabled
		}
	}

	// Check percentage rollout
	if feature.Percentage < 100 {
		hash := hashString(userID + featureName)
		return float64(hash%100) < feature.Percentage
	}

	return true
}

// SetFeature sets a feature flag
func (ft *FeatureToggler) SetFeature(feature *Feature) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	feature.UpdatedAt = time.Now()
	ft.features[feature.Name] = feature
	ft.lastUpdate = time.Now()
}

// GetFeature gets a feature flag
func (ft *FeatureToggler) GetFeature(featureName string) (*Feature, bool) {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	feature, exists := ft.features[featureName]
	return feature, exists
}

// hashString creates a hash for consistent percentage rollout
func hashString(s string) int {
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = hash*31 + int(s[i])
	}
	return hash
}
