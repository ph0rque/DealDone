package deployment

import (
	"fmt"
	"sync"
	"time"
)

// ABTestingFramework manages A/B testing for gradual feature rollout
type ABTestingFramework struct {
	mu                sync.RWMutex
	experiments       map[string]*Experiment
	userAssignments   map[string]map[string]string // userID -> experimentID -> variant
	metrics           map[string]*ExperimentMetrics
	feedbackCollector *FeedbackCollector
	adoptionTracker   *AdoptionTracker
	eventLogger       *EventLogger
	config            *ABTestConfig
	lastUpdate        time.Time
}

// Experiment represents an A/B test experiment
type Experiment struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Status         ExperimentStatus       `json:"status"`
	Type           ExperimentType         `json:"type"`
	Variants       []Variant              `json:"variants"`
	TrafficSplit   map[string]float64     `json:"traffic_split"`
	TargetAudience *TargetAudience        `json:"target_audience"`
	StartDate      time.Time              `json:"start_date"`
	EndDate        time.Time              `json:"end_date"`
	Goals          []ExperimentGoal       `json:"goals"`
	Metrics        []MetricDefinition     `json:"metrics"`
	Configuration  map[string]interface{} `json:"configuration"`
	CreatedBy      string                 `json:"created_by"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Tags           []string               `json:"tags"`
}

// Variant represents a variant in an A/B test
type Variant struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	IsControl     bool                   `json:"is_control"`
	Configuration map[string]interface{} `json:"configuration"`
	Weight        float64                `json:"weight"`
	Enabled       bool                   `json:"enabled"`
}

// ExperimentMetrics tracks metrics for an experiment
type ExperimentMetrics struct {
	ExperimentID       string                     `json:"experiment_id"`
	VariantMetrics     map[string]*VariantMetrics `json:"variant_metrics"`
	OverallMetrics     *OverallMetrics            `json:"overall_metrics"`
	StatisticalResults *StatisticalResults        `json:"statistical_results"`
	LastUpdated        time.Time                  `json:"last_updated"`
}

// VariantMetrics tracks metrics for a specific variant
type VariantMetrics struct {
	VariantID          string             `json:"variant_id"`
	UserCount          int64              `json:"user_count"`
	ConversionRate     float64            `json:"conversion_rate"`
	EngagementRate     float64            `json:"engagement_rate"`
	RetentionRate      float64            `json:"retention_rate"`
	SatisfactionScore  float64            `json:"satisfaction_score"`
	ErrorRate          float64            `json:"error_rate"`
	PerformanceMetrics map[string]float64 `json:"performance_metrics"`
	CustomMetrics      map[string]float64 `json:"custom_metrics"`
	SampleSize         int64              `json:"sample_size"`
	Timestamp          time.Time          `json:"timestamp"`
}

// OverallMetrics provides overall experiment insights
type OverallMetrics struct {
	TotalUsers        int64         `json:"total_users"`
	TotalConversions  int64         `json:"total_conversions"`
	AverageEngagement float64       `json:"average_engagement"`
	WinningVariant    string        `json:"winning_variant"`
	Confidence        float64       `json:"confidence"`
	Duration          time.Duration `json:"duration"`
	Status            string        `json:"status"`
}

// StatisticalResults provides statistical analysis
type StatisticalResults struct {
	PValue             float64            `json:"p_value"`
	ConfidenceLevel    float64            `json:"confidence_level"`
	StatisticalPower   float64            `json:"statistical_power"`
	EffectSize         float64            `json:"effect_size"`
	SampleSizeNeeded   int64              `json:"sample_size_needed"`
	IsSignificant      bool               `json:"is_significant"`
	WinningVariant     string             `json:"winning_variant"`
	LiftPercentage     float64            `json:"lift_percentage"`
	VariantComparisons map[string]float64 `json:"variant_comparisons"`
}

// FeedbackCollector collects user feedback during experiments
type FeedbackCollector struct {
	mu           sync.RWMutex
	feedback     map[string][]UserFeedback
	surveys      map[string]*Survey
	ratings      map[string][]Rating
	suggestions  []Suggestion
	lastAnalysis time.Time
}

// UserFeedback represents user feedback
type UserFeedback struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	Type         FeedbackType           `json:"type"`
	Content      string                 `json:"content"`
	Rating       int                    `json:"rating"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"`
	Category     string                 `json:"category"`
	Sentiment    SentimentScore         `json:"sentiment"`
}

// Survey represents a user survey
type Survey struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	Questions     []SurveyQuestion `json:"questions"`
	ExperimentID  string           `json:"experiment_id"`
	VariantID     string           `json:"variant_id"`
	Active        bool             `json:"active"`
	ResponseCount int              `json:"response_count"`
	CreatedAt     time.Time        `json:"created_at"`
}

// SurveyQuestion represents a survey question
type SurveyQuestion struct {
	ID       string       `json:"id"`
	Type     QuestionType `json:"type"`
	Text     string       `json:"text"`
	Options  []string     `json:"options"`
	Required bool         `json:"required"`
	Order    int          `json:"order"`
}

// Rating represents a user rating
type Rating struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ExperimentID string    `json:"experiment_id"`
	VariantID    string    `json:"variant_id"`
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	Category     string    `json:"category"`
	Comment      string    `json:"comment"`
	Timestamp    time.Time `json:"timestamp"`
}

// Suggestion represents a user suggestion
type Suggestion struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Priority     PriorityLevel          `json:"priority"`
	Status       SuggestionStatus       `json:"status"`
	Category     string                 `json:"category"`
	Metadata     map[string]interface{} `json:"metadata"`
	Votes        int                    `json:"votes"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AdoptionTracker tracks feature adoption metrics
type AdoptionTracker struct {
	mu              sync.RWMutex
	adoptionMetrics map[string]*AdoptionMetrics
	userJourney     map[string][]UserJourneyEvent
	cohortAnalysis  map[string]*CohortAnalysis
	retentionData   map[string]*RetentionData
	lastUpdate      time.Time
}

// AdoptionMetrics tracks adoption statistics
type AdoptionMetrics struct {
	FeatureID           string           `json:"feature_id"`
	ExperimentID        string           `json:"experiment_id"`
	VariantID           string           `json:"variant_id"`
	TotalUsers          int64            `json:"total_users"`
	ActiveUsers         int64            `json:"active_users"`
	NewUsers            int64            `json:"new_users"`
	ReturningUsers      int64            `json:"returning_users"`
	AdoptionRate        float64          `json:"adoption_rate"`
	RetentionRate       float64          `json:"retention_rate"`
	EngagementScore     float64          `json:"engagement_score"`
	TimeToAdoption      time.Duration    `json:"time_to_adoption"`
	UsageFrequency      map[string]int64 `json:"usage_frequency"`
	UserSegments        map[string]int64 `json:"user_segments"`
	GeographicData      map[string]int64 `json:"geographic_data"`
	DeviceData          map[string]int64 `json:"device_data"`
	FeatureInteractions map[string]int64 `json:"feature_interactions"`
	Timestamp           time.Time        `json:"timestamp"`
}

// UserJourneyEvent represents an event in the user journey
type UserJourneyEvent struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	EventType    string                 `json:"event_type"`
	EventName    string                 `json:"event_name"`
	Properties   map[string]interface{} `json:"properties"`
	Timestamp    time.Time              `json:"timestamp"`
	SessionID    string                 `json:"session_id"`
	Source       string                 `json:"source"`
}

// CohortAnalysis provides cohort-based insights
type CohortAnalysis struct {
	CohortID        string             `json:"cohort_id"`
	ExperimentID    string             `json:"experiment_id"`
	VariantID       string             `json:"variant_id"`
	CohortSize      int64              `json:"cohort_size"`
	RetentionRates  map[string]float64 `json:"retention_rates"`
	EngagementRates map[string]float64 `json:"engagement_rates"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
	TimeIntervals   []string           `json:"time_intervals"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// RetentionData tracks user retention
type RetentionData struct {
	ExperimentID     string             `json:"experiment_id"`
	VariantID        string             `json:"variant_id"`
	Day1Retention    float64            `json:"day1_retention"`
	Day7Retention    float64            `json:"day7_retention"`
	Day30Retention   float64            `json:"day30_retention"`
	WeeklyRetention  map[string]float64 `json:"weekly_retention"`
	MonthlyRetention map[string]float64 `json:"monthly_retention"`
	ChurnRate        float64            `json:"churn_rate"`
	LastUpdated      time.Time          `json:"last_updated"`
}

// EventLogger logs experiment events
type EventLogger struct {
	mu     sync.RWMutex
	events []ExperimentEvent
	config *LoggerConfig
}

// ExperimentEvent represents an experiment event
type ExperimentEvent struct {
	ID           string                 `json:"id"`
	Type         EventType              `json:"type"`
	ExperimentID string                 `json:"experiment_id"`
	VariantID    string                 `json:"variant_id"`
	UserID       string                 `json:"user_id"`
	EventData    map[string]interface{} `json:"event_data"`
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"`
	SessionID    string                 `json:"session_id"`
}

// Configuration and supporting types
type ABTestConfig struct {
	DefaultTrafficSplit    map[string]float64 `json:"default_traffic_split"`
	MinSampleSize          int64              `json:"min_sample_size"`
	MaxExperimentDuration  time.Duration      `json:"max_experiment_duration"`
	ConfidenceLevel        float64            `json:"confidence_level"`
	StatisticalPowerTarget float64            `json:"statistical_power_target"`
	AutoStopEnabled        bool               `json:"auto_stop_enabled"`
	MetricsUpdateInterval  time.Duration      `json:"metrics_update_interval"`
}

type TargetAudience struct {
	UserSegments    []string         `json:"user_segments"`
	GeographicRules []GeographicRule `json:"geographic_rules"`
	DeviceRules     []DeviceRule     `json:"device_rules"`
	BehaviorRules   []BehaviorRule   `json:"behavior_rules"`
	Percentage      float64          `json:"percentage"`
	ExcludeRules    []ExclusionRule  `json:"exclude_rules"`
}

type GeographicRule struct {
	Type      string   `json:"type"`
	Countries []string `json:"countries"`
	Regions   []string `json:"regions"`
	Cities    []string `json:"cities"`
}

type DeviceRule struct {
	Types     []string `json:"types"`
	Platforms []string `json:"platforms"`
	Browsers  []string `json:"browsers"`
}

type BehaviorRule struct {
	EventType string      `json:"event_type"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	TimeFrame string      `json:"time_frame"`
}

type ExclusionRule struct {
	Type   string      `json:"type"`
	Field  string      `json:"field"`
	Value  interface{} `json:"value"`
	Reason string      `json:"reason"`
}

type ExperimentGoal struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        GoalType `json:"type"`
	MetricName  string   `json:"metric_name"`
	TargetValue float64  `json:"target_value"`
	Operator    string   `json:"operator"`
	Priority    int      `json:"priority"`
	Description string   `json:"description"`
}

type MetricDefinition struct {
	Name        string     `json:"name"`
	Type        MetricType `json:"type"`
	Description string     `json:"description"`
	Aggregation string     `json:"aggregation"`
	Filters     []Filter   `json:"filters"`
	Formula     string     `json:"formula"`
	Unit        string     `json:"unit"`
	IsPrimary   bool       `json:"is_primary"`
}

type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type LoggerConfig struct {
	BufferSize         int           `json:"buffer_size"`
	FlushInterval      time.Duration `json:"flush_interval"`
	RetentionDays      int           `json:"retention_days"`
	CompressionEnabled bool          `json:"compression_enabled"`
}

// Enums
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "cancelled"
)

type ExperimentType string

const (
	ExperimentTypeAB           ExperimentType = "ab"
	ExperimentTypeMultivariate ExperimentType = "multivariate"
	ExperimentTypeFeatureFlag  ExperimentType = "feature_flag"
	ExperimentTypeBandit       ExperimentType = "bandit"
)

type FeedbackType string

const (
	FeedbackTypeRating     FeedbackType = "rating"
	FeedbackTypeComment    FeedbackType = "comment"
	FeedbackTypeSuggestion FeedbackType = "suggestion"
	FeedbackTypeBugReport  FeedbackType = "bug_report"
)

type SentimentScore string

const (
	SentimentPositive SentimentScore = "positive"
	SentimentNeutral  SentimentScore = "neutral"
	SentimentNegative SentimentScore = "negative"
)

type QuestionType string

const (
	QuestionTypeText     QuestionType = "text"
	QuestionTypeRating   QuestionType = "rating"
	QuestionTypeChoice   QuestionType = "choice"
	QuestionTypeMultiple QuestionType = "multiple"
)

type PriorityLevel string

const (
	PriorityLow      PriorityLevel = "low"
	PriorityMedium   PriorityLevel = "medium"
	PriorityHigh     PriorityLevel = "high"
	PriorityCritical PriorityLevel = "critical"
)

type SuggestionStatus string

const (
	SuggestionStatusNew         SuggestionStatus = "new"
	SuggestionStatusReviewing   SuggestionStatus = "reviewing"
	SuggestionStatusApproved    SuggestionStatus = "approved"
	SuggestionStatusRejected    SuggestionStatus = "rejected"
	SuggestionStatusImplemented SuggestionStatus = "implemented"
)

type EventType string

const (
	EventTypeExperimentStart EventType = "experiment_start"
	EventTypeExperimentEnd   EventType = "experiment_end"
	EventTypeUserAssignment  EventType = "user_assignment"
	EventTypeConversion      EventType = "conversion"
	EventTypeEngagement      EventType = "engagement"
	EventTypeFeedback        EventType = "feedback"
)

type GoalType string

const (
	GoalTypeConversion GoalType = "conversion"
	GoalTypeEngagement GoalType = "engagement"
	GoalTypeRetention  GoalType = "retention"
	GoalTypeRevenue    GoalType = "revenue"
	GoalTypeCustom     GoalType = "custom"
)

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeRate      MetricType = "rate"
)

// NewABTestingFramework creates a new A/B testing framework
func NewABTestingFramework() *ABTestingFramework {
	return &ABTestingFramework{
		experiments:       make(map[string]*Experiment),
		userAssignments:   make(map[string]map[string]string),
		metrics:           make(map[string]*ExperimentMetrics),
		feedbackCollector: NewFeedbackCollector(),
		adoptionTracker:   NewAdoptionTracker(),
		eventLogger:       NewEventLogger(),
		config: &ABTestConfig{
			DefaultTrafficSplit: map[string]float64{
				"control":   50.0,
				"treatment": 50.0,
			},
			MinSampleSize:          1000,
			MaxExperimentDuration:  30 * 24 * time.Hour,
			ConfidenceLevel:        0.95,
			StatisticalPowerTarget: 0.8,
			AutoStopEnabled:        true,
			MetricsUpdateInterval:  1 * time.Hour,
		},
		lastUpdate: time.Now(),
	}
}

// CreateExperiment creates a new A/B test experiment
func (ab *ABTestingFramework) CreateExperiment(experiment *Experiment) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment.CreatedAt = time.Now()
	experiment.UpdatedAt = time.Now()
	experiment.Status = ExperimentStatusDraft

	// Validate experiment configuration
	if err := ab.validateExperiment(experiment); err != nil {
		return fmt.Errorf("experiment validation failed: %w", err)
	}

	ab.experiments[experiment.ID] = experiment

	// Initialize metrics
	ab.metrics[experiment.ID] = &ExperimentMetrics{
		ExperimentID:   experiment.ID,
		VariantMetrics: make(map[string]*VariantMetrics),
		OverallMetrics: &OverallMetrics{},
		StatisticalResults: &StatisticalResults{
			ConfidenceLevel: ab.config.ConfidenceLevel,
		},
		LastUpdated: time.Now(),
	}

	// Initialize variant metrics
	for _, variant := range experiment.Variants {
		ab.metrics[experiment.ID].VariantMetrics[variant.ID] = &VariantMetrics{
			VariantID:          variant.ID,
			PerformanceMetrics: make(map[string]float64),
			CustomMetrics:      make(map[string]float64),
			Timestamp:          time.Now(),
		}
	}

	ab.lastUpdate = time.Now()
	return nil
}

// StartExperiment starts an A/B test experiment
func (ab *ABTestingFramework) StartExperiment(experimentID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment, exists := ab.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != ExperimentStatusDraft {
		return fmt.Errorf("experiment %s is not in draft status", experimentID)
	}

	experiment.Status = ExperimentStatusRunning
	experiment.StartDate = time.Now()
	experiment.UpdatedAt = time.Now()

	// Log experiment start event
	ab.eventLogger.LogEvent(&ExperimentEvent{
		ID:           fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:         EventTypeExperimentStart,
		ExperimentID: experimentID,
		EventData: map[string]interface{}{
			"start_time": time.Now(),
			"variants":   len(experiment.Variants),
		},
		Timestamp: time.Now(),
		Source:    "ab_framework",
	})

	return nil
}

// AssignUserToVariant assigns a user to a variant
func (ab *ABTestingFramework) AssignUserToVariant(userID, experimentID string) (string, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment, exists := ab.experiments[experimentID]
	if !exists {
		return "", fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != ExperimentStatusRunning {
		return "", fmt.Errorf("experiment %s is not running", experimentID)
	}

	// Check if user is already assigned
	if userExperiments, exists := ab.userAssignments[userID]; exists {
		if variantID, exists := userExperiments[experimentID]; exists {
			return variantID, nil
		}
	}

	// Assign user to variant based on traffic split
	variantID := ab.selectVariantForUser(userID, experiment)

	// Store assignment
	if ab.userAssignments[userID] == nil {
		ab.userAssignments[userID] = make(map[string]string)
	}
	ab.userAssignments[userID][experimentID] = variantID

	// Update metrics
	if metrics, exists := ab.metrics[experimentID].VariantMetrics[variantID]; exists {
		metrics.UserCount++
		metrics.SampleSize++
		metrics.Timestamp = time.Now()
	}

	// Log assignment event
	ab.eventLogger.LogEvent(&ExperimentEvent{
		ID:           fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:         EventTypeUserAssignment,
		ExperimentID: experimentID,
		VariantID:    variantID,
		UserID:       userID,
		EventData: map[string]interface{}{
			"assignment_time": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "ab_framework",
	})

	return variantID, nil
}

// selectVariantForUser selects a variant for a user based on traffic split
func (ab *ABTestingFramework) selectVariantForUser(userID string, experiment *Experiment) string {
	// Use consistent hashing for user assignment
	hash := ab.hashUserForExperiment(userID, experiment.ID)
	position := float64(hash%10000) / 100.0 // 0-100

	cumulativeWeight := 0.0
	for _, variant := range experiment.Variants {
		if !variant.Enabled {
			continue
		}

		cumulativeWeight += variant.Weight
		if position <= cumulativeWeight {
			return variant.ID
		}
	}

	// Fallback to control variant
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			return variant.ID
		}
	}

	// Last resort: return first enabled variant
	for _, variant := range experiment.Variants {
		if variant.Enabled {
			return variant.ID
		}
	}

	return experiment.Variants[0].ID
}

// hashUserForExperiment creates a hash for consistent user assignment
func (ab *ABTestingFramework) hashUserForExperiment(userID, experimentID string) int {
	combined := userID + experimentID
	hash := 0
	for i := 0; i < len(combined); i++ {
		hash = hash*31 + int(combined[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// TrackConversion tracks a conversion event
func (ab *ABTestingFramework) TrackConversion(userID, experimentID string, value float64) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// Get user's variant assignment
	variantID, exists := ab.getUserVariant(userID, experimentID)
	if !exists {
		return fmt.Errorf("user %s not assigned to experiment %s", userID, experimentID)
	}

	// Update conversion metrics
	if metrics, exists := ab.metrics[experimentID].VariantMetrics[variantID]; exists {
		// Simple conversion rate calculation (could be more sophisticated)
		totalConversions := metrics.ConversionRate * float64(metrics.UserCount)
		totalConversions += 1
		metrics.ConversionRate = totalConversions / float64(metrics.UserCount)
		metrics.Timestamp = time.Now()
	}

	// Log conversion event
	ab.eventLogger.LogEvent(&ExperimentEvent{
		ID:           fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:         EventTypeConversion,
		ExperimentID: experimentID,
		VariantID:    variantID,
		UserID:       userID,
		EventData: map[string]interface{}{
			"conversion_value": value,
			"timestamp":        time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "ab_framework",
	})

	return nil
}

// GetUserVariant gets the variant assigned to a user
func (ab *ABTestingFramework) GetUserVariant(userID, experimentID string) (string, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	variantID, exists := ab.getUserVariant(userID, experimentID)
	if !exists {
		return "", fmt.Errorf("user %s not assigned to experiment %s", userID, experimentID)
	}

	return variantID, nil
}

// getUserVariant internal method to get user variant
func (ab *ABTestingFramework) getUserVariant(userID, experimentID string) (string, bool) {
	if userExperiments, exists := ab.userAssignments[userID]; exists {
		if variantID, exists := userExperiments[experimentID]; exists {
			return variantID, true
		}
	}
	return "", false
}

// GetExperimentMetrics returns metrics for an experiment
func (ab *ABTestingFramework) GetExperimentMetrics(experimentID string) (*ExperimentMetrics, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	metrics, exists := ab.metrics[experimentID]
	if !exists {
		return nil, fmt.Errorf("metrics for experiment %s not found", experimentID)
	}

	return metrics, nil
}

// StopExperiment stops an experiment
func (ab *ABTestingFramework) StopExperiment(experimentID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment, exists := ab.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	experiment.Status = ExperimentStatusCompleted
	experiment.EndDate = time.Now()
	experiment.UpdatedAt = time.Now()

	// Calculate final metrics
	ab.calculateFinalMetrics(experimentID)

	// Log experiment end event
	ab.eventLogger.LogEvent(&ExperimentEvent{
		ID:           fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:         EventTypeExperimentEnd,
		ExperimentID: experimentID,
		EventData: map[string]interface{}{
			"end_time": time.Now(),
			"duration": time.Since(experiment.StartDate),
		},
		Timestamp: time.Now(),
		Source:    "ab_framework",
	})

	return nil
}

// calculateFinalMetrics calculates final statistical metrics
func (ab *ABTestingFramework) calculateFinalMetrics(experimentID string) {
	metrics := ab.metrics[experimentID]

	// Simple statistical calculations (in a real implementation, use proper statistical libraries)
	var totalUsers int64
	var bestVariant string
	var bestConversionRate float64

	for variantID, variantMetrics := range metrics.VariantMetrics {
		totalUsers += variantMetrics.UserCount
		if variantMetrics.ConversionRate > bestConversionRate {
			bestConversionRate = variantMetrics.ConversionRate
			bestVariant = variantID
		}
	}

	metrics.OverallMetrics = &OverallMetrics{
		TotalUsers:     totalUsers,
		WinningVariant: bestVariant,
		Confidence:     95.0, // Simplified
		Duration:       time.Since(ab.experiments[experimentID].StartDate),
		Status:         "completed",
	}

	metrics.StatisticalResults = &StatisticalResults{
		WinningVariant:     bestVariant,
		IsSignificant:      bestConversionRate > 0.1, // Simplified
		ConfidenceLevel:    0.95,
		StatisticalPower:   0.8,
		VariantComparisons: make(map[string]float64),
	}

	metrics.LastUpdated = time.Now()
}

// validateExperiment validates experiment configuration
func (ab *ABTestingFramework) validateExperiment(experiment *Experiment) error {
	if experiment.ID == "" {
		return fmt.Errorf("experiment ID is required")
	}

	if len(experiment.Variants) < 2 {
		return fmt.Errorf("experiment must have at least 2 variants")
	}

	// Check for control variant
	hasControl := false
	totalWeight := 0.0
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			hasControl = true
		}
		totalWeight += variant.Weight
	}

	if !hasControl {
		return fmt.Errorf("experiment must have a control variant")
	}

	if totalWeight != 100.0 {
		return fmt.Errorf("variant weights must sum to 100")
	}

	return nil
}

// Supporting component constructors
func NewFeedbackCollector() *FeedbackCollector {
	return &FeedbackCollector{
		feedback:     make(map[string][]UserFeedback),
		surveys:      make(map[string]*Survey),
		ratings:      make(map[string][]Rating),
		suggestions:  make([]Suggestion, 0),
		lastAnalysis: time.Now(),
	}
}

func NewAdoptionTracker() *AdoptionTracker {
	return &AdoptionTracker{
		adoptionMetrics: make(map[string]*AdoptionMetrics),
		userJourney:     make(map[string][]UserJourneyEvent),
		cohortAnalysis:  make(map[string]*CohortAnalysis),
		retentionData:   make(map[string]*RetentionData),
		lastUpdate:      time.Now(),
	}
}

func NewEventLogger() *EventLogger {
	return &EventLogger{
		events: make([]ExperimentEvent, 0),
		config: &LoggerConfig{
			BufferSize:         10000,
			FlushInterval:      5 * time.Minute,
			RetentionDays:      90,
			CompressionEnabled: true,
		},
	}
}

// LogEvent logs an experiment event
func (el *EventLogger) LogEvent(event *ExperimentEvent) {
	el.mu.Lock()
	defer el.mu.Unlock()

	el.events = append(el.events, *event)

	// Keep buffer size manageable
	if len(el.events) > el.config.BufferSize {
		el.events = el.events[len(el.events)-el.config.BufferSize:]
	}
}
