package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// FeedbackType defines the type of feedback provided by users
type FeedbackType string

const (
	FeedbackPositive   FeedbackType = "positive"
	FeedbackNegative   FeedbackType = "negative"
	FeedbackCorrection FeedbackType = "correction"
	FeedbackSuggestion FeedbackType = "suggestion"
	FeedbackValidation FeedbackType = "validation"
	FeedbackRejection  FeedbackType = "rejection"
)

// FeedbackSeverity indicates the importance of the feedback
type FeedbackSeverity string

const (
	FeedbackSeverityLow      FeedbackSeverity = "low"
	FeedbackSeverityMedium   FeedbackSeverity = "medium"
	FeedbackSeverityHigh     FeedbackSeverity = "high"
	FeedbackSeverityCritical FeedbackSeverity = "critical"
)

// UserFeedback represents feedback provided by a user
type UserFeedback struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	SessionID     string                 `json:"session_id"`
	FeedbackType  FeedbackType           `json:"feedback_type"`
	Severity      FeedbackSeverity       `json:"severity"`
	TargetID      string                 `json:"target_id"`   // ID of the item being reviewed
	TargetType    string                 `json:"target_type"` // correction, recommendation, suggestion
	OriginalValue interface{}            `json:"original_value"`
	FeedbackValue interface{}            `json:"feedback_value"`
	Comment       string                 `json:"comment"`
	Context       map[string]interface{} `json:"context"`
	Timestamp     time.Time              `json:"timestamp"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
	IsProcessed   bool                   `json:"is_processed"`
	Impact        *FeedbackImpact        `json:"impact,omitempty"`
	UserSession   UserSessionContext     `json:"user_session"`
	Confidence    float64                `json:"confidence"`
	Tags          []string               `json:"tags"`
}

// FeedbackImpact tracks the impact of processing feedback
type FeedbackImpact struct {
	ModelUpdated       bool          `json:"model_updated"`
	PatternsAffected   []string      `json:"patterns_affected"`
	ConfidenceChange   float64       `json:"confidence_change"`
	LearningAdjustment float64       `json:"learning_adjustment"`
	ProcessingTime     time.Duration `json:"processing_time"`
	ImpactScore        float64       `json:"impact_score"`
	AppliedAt          time.Time     `json:"applied_at"`
}

// UserSessionContext provides context about the user's session
type UserSessionContext struct {
	DealID          string                 `json:"deal_id"`
	DocumentID      string                 `json:"document_id"`
	TemplateID      string                 `json:"template_id"`
	ProcessingStage string                 `json:"processing_stage"`
	UserExpertise   string                 `json:"user_expertise"`
	SessionMetrics  map[string]interface{} `json:"session_metrics"`
	TimeSpent       time.Duration          `json:"time_spent"`
}

// FeedbackLoop manages the continuous learning improvement loop
type FeedbackLoop struct {
	config               FeedbackLoopConfig
	feedbackQueue        chan *UserFeedback
	processingQueue      []*UserFeedback
	feedbackHistory      map[string]*UserFeedback
	userFeedbackProfiles map[string]*UserFeedbackProfile
	learningAdjustments  map[string]*LearningAdjustment
	correctionProcessor  *CorrectionProcessor
	ragEngine            *AdvancedRAGEngine
	feedbackAnalyzer     *FeedbackAnalyzer
	impactCalculator     *ImpactCalculator
	mutex                sync.RWMutex
	logger               Logger
	ctx                  context.Context
	cancel               context.CancelFunc
	metrics              FeedbackMetrics
}

// FeedbackLoopConfig holds configuration for feedback processing
type FeedbackLoopConfig struct {
	QueueSize                int                          `json:"queue_size"`
	ProcessingInterval       time.Duration                `json:"processing_interval"`
	MinFeedbacksToProcess    int                          `json:"min_feedbacks_to_process"`
	MaxProcessingBatch       int                          `json:"max_processing_batch"`
	FeedbackRetentionDays    int                          `json:"feedback_retention_days"`
	ImpactThreshold          float64                      `json:"impact_threshold"`
	ConfidenceAdjustment     float64                      `json:"confidence_adjustment"`
	LearningRateModifier     float64                      `json:"learning_rate_modifier"`
	StoragePath              string                       `json:"storage_path"`
	EnableRealTimeProcessing bool                         `json:"enable_real_time_processing"`
	EnableBatchProcessing    bool                         `json:"enable_batch_processing"`
	UserTrustWeighting       float64                      `json:"user_trust_weighting"`
	SeverityMultipliers      map[FeedbackSeverity]float64 `json:"severity_multipliers"`
}

// UserFeedbackProfile tracks a user's feedback patterns
type UserFeedbackProfile struct {
	UserID              string                   `json:"user_id"`
	TotalFeedbacks      int                      `json:"total_feedbacks"`
	FeedbackByType      map[FeedbackType]int     `json:"feedback_by_type"`
	FeedbackBySeverity  map[FeedbackSeverity]int `json:"feedback_by_severity"`
	AccuracyScore       float64                  `json:"accuracy_score"`
	ReliabilityScore    float64                  `json:"reliability_score"`
	ExpertiseAreas      []string                 `json:"expertise_areas"`
	FeedbackVelocity    float64                  `json:"feedback_velocity"`
	LastFeedbackTime    time.Time                `json:"last_feedback_time"`
	AverageResponseTime time.Duration            `json:"average_response_time"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	IsActive            bool                     `json:"is_active"`
}

// LearningAdjustment represents an adjustment to the learning model
type LearningAdjustment struct {
	ID              string                 `json:"id"`
	FeedbackID      string                 `json:"feedback_id"`
	AdjustmentType  string                 `json:"adjustment_type"`
	TargetComponent string                 `json:"target_component"` // pattern, embedding, node
	TargetID        string                 `json:"target_id"`
	Adjustment      map[string]interface{} `json:"adjustment"`
	Confidence      float64                `json:"confidence"`
	AppliedAt       time.Time              `json:"applied_at"`
	Impact          float64                `json:"impact"`
	Reversible      bool                   `json:"reversible"`
	ReversalData    map[string]interface{} `json:"reversal_data,omitempty"`
}

// FeedbackAnalyzer analyzes feedback patterns and trends
type FeedbackAnalyzer struct {
	patterns          map[string]*FeedbackPattern
	trendAnalysis     *TrendAnalysis
	anomalyDetection  *AnomalyDetection
	userBehaviorModel *UserBehaviorModel
	mutex             sync.RWMutex
}

// FeedbackPattern represents a pattern in user feedback
type FeedbackPattern struct {
	ID                string                 `json:"id"`
	PatternType       string                 `json:"pattern_type"`
	Description       string                 `json:"description"`
	Frequency         int                    `json:"frequency"`
	Confidence        float64                `json:"confidence"`
	Context           map[string]interface{} `json:"context"`
	FirstSeen         time.Time              `json:"first_seen"`
	LastSeen          time.Time              `json:"last_seen"`
	Impact            float64                `json:"impact"`
	Actionable        bool                   `json:"actionable"`
	RecommendedAction string                 `json:"recommended_action"`
}

// ImpactCalculator calculates the impact of feedback on learning
type ImpactCalculator struct {
	impactHistory   []ImpactMeasurement
	baselineMetrics BaselineMetrics
	calculator      *MetricsCalculator
	mutex           sync.RWMutex
}

// ImpactMeasurement tracks the measured impact of changes
type ImpactMeasurement struct {
	ID              string    `json:"id"`
	FeedbackID      string    `json:"feedback_id"`
	MetricType      string    `json:"metric_type"`
	BeforeValue     float64   `json:"before_value"`
	AfterValue      float64   `json:"after_value"`
	ImpactScore     float64   `json:"impact_score"`
	MeasuredAt      time.Time `json:"measured_at"`
	ConfidenceLevel float64   `json:"confidence_level"`
}

// BaselineMetrics stores baseline performance metrics
type BaselineMetrics struct {
	AccuracyScore         float64       `json:"accuracy_score"`
	LearningEffectiveness float64       `json:"learning_effectiveness"`
	UserSatisfaction      float64       `json:"user_satisfaction"`
	ResponseTime          time.Duration `json:"response_time"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// FeedbackMetrics tracks overall feedback system performance
type FeedbackMetrics struct {
	TotalFeedbacks        int                      `json:"total_feedbacks"`
	ProcessedFeedbacks    int                      `json:"processed_feedbacks"`
	PendingFeedbacks      int                      `json:"pending_feedbacks"`
	AverageProcessingTime time.Duration            `json:"average_processing_time"`
	FeedbackByType        map[FeedbackType]int     `json:"feedback_by_type"`
	FeedbackBySeverity    map[FeedbackSeverity]int `json:"feedback_by_severity"`
	ImpactDistribution    map[string]int           `json:"impact_distribution"`
	UserParticipation     map[string]int           `json:"user_participation"`
	LearningImprovements  float64                  `json:"learning_improvements"`
	LastUpdated           time.Time                `json:"last_updated"`
}

// TrendAnalysis analyzes trends in feedback data
type TrendAnalysis struct {
	trends        map[string]*Trend
	trendDetector *TrendDetector
	mutex         sync.RWMutex
}

// Trend represents a detected trend in feedback
type Trend struct {
	ID          string              `json:"id"`
	TrendType   string              `json:"trend_type"`
	Direction   string              `json:"direction"` // increasing, decreasing, stable
	Strength    float64             `json:"strength"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     *time.Time          `json:"end_time,omitempty"`
	Description string              `json:"description"`
	Confidence  float64             `json:"confidence"`
	DataPoints  []FeedbackDataPoint `json:"data_points"`
}

// FeedbackDataPoint represents a single data point in a trend
type FeedbackDataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Context   map[string]interface{} `json:"context"`
}

// AnomalyDetection detects anomalies in feedback patterns
type AnomalyDetection struct {
	anomalies      []Anomaly
	detectionRules []DetectionRule
	threshold      float64
	mutex          sync.RWMutex
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Severity     string    `json:"severity"`
	Description  string    `json:"description"`
	DetectedAt   time.Time `json:"detected_at"`
	AffectedArea string    `json:"affected_area"`
	Confidence   float64   `json:"confidence"`
	Action       string    `json:"action"`
}

// DetectionRule defines rules for anomaly detection
type DetectionRule struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Threshold float64 `json:"threshold"`
	Condition string  `json:"condition"`
	Action    string  `json:"action"`
	IsActive  bool    `json:"is_active"`
}

// UserBehaviorModel models user behavior patterns
type UserBehaviorModel struct {
	userModels    map[string]*UserModel
	behaviorRules []BehaviorRule
	mutex         sync.RWMutex
}

// UserModel represents a model of user behavior
type UserModel struct {
	UserID           string                 `json:"user_id"`
	BehaviorProfile  BehaviorProfile        `json:"behavior_profile"`
	FeedbackPatterns []FeedbackPattern      `json:"feedback_patterns"`
	Preferences      map[string]interface{} `json:"preferences"`
	Expertise        ExpertiseModel         `json:"expertise"`
	LastUpdated      time.Time              `json:"last_updated"`
}

// BehaviorProfile captures user behavior characteristics
type BehaviorProfile struct {
	ResponseStyle     string    `json:"response_style"`     // quick, thoughtful, detailed
	FeedbackFrequency string    `json:"feedback_frequency"` // high, medium, low
	AccuracyLevel     float64   `json:"accuracy_level"`
	Consistency       float64   `json:"consistency"`
	Expertise         float64   `json:"expertise"`
	Engagement        float64   `json:"engagement"`
	LastActive        time.Time `json:"last_active"`
}

// ExpertiseModel tracks user expertise in different areas
type ExpertiseModel struct {
	Areas           map[string]float64 `json:"areas"` // area -> expertise level
	Certifications  []string           `json:"certifications"`
	Experience      map[string]int     `json:"experience"` // area -> years of experience
	ConfidenceLevel float64            `json:"confidence_level"`
	LastAssessed    time.Time          `json:"last_assessed"`
}

// BehaviorRule defines rules for behavior analysis
type BehaviorRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Condition  string                 `json:"condition"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	IsActive   bool                   `json:"is_active"`
}

// TrendDetector detects trends in data
type TrendDetector struct {
	windowSize       int
	sensitivityLevel float64
	mutex            sync.RWMutex
}

// MetricsCalculator calculates various metrics
type MetricsCalculator struct {
	calculationRules map[string]CalculationRule
	mutex            sync.RWMutex
}

// CalculationRule defines how to calculate metrics
type CalculationRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Formula    string                 `json:"formula"`
	Parameters map[string]interface{} `json:"parameters"`
	IsActive   bool                   `json:"is_active"`
}

// Constructor functions

func NewFeedbackLoop(config FeedbackLoopConfig, correctionProcessor *CorrectionProcessor, ragEngine *AdvancedRAGEngine, logger Logger) *FeedbackLoop {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default severity multipliers if not provided
	if config.SeverityMultipliers == nil {
		config.SeverityMultipliers = map[FeedbackSeverity]float64{
			FeedbackSeverityLow:      1.0,
			FeedbackSeverityMedium:   1.5,
			FeedbackSeverityHigh:     2.0,
			FeedbackSeverityCritical: 3.0,
		}
	}

	loop := &FeedbackLoop{
		config:               config,
		feedbackQueue:        make(chan *UserFeedback, config.QueueSize),
		processingQueue:      make([]*UserFeedback, 0),
		feedbackHistory:      make(map[string]*UserFeedback),
		userFeedbackProfiles: make(map[string]*UserFeedbackProfile),
		learningAdjustments:  make(map[string]*LearningAdjustment),
		correctionProcessor:  correctionProcessor,
		ragEngine:            ragEngine,
		feedbackAnalyzer:     NewFeedbackAnalyzer(),
		impactCalculator:     NewImpactCalculator(),
		mutex:                sync.RWMutex{},
		logger:               logger,
		ctx:                  ctx,
		cancel:               cancel,
		metrics: FeedbackMetrics{
			FeedbackByType:     make(map[FeedbackType]int),
			FeedbackBySeverity: make(map[FeedbackSeverity]int),
			ImpactDistribution: make(map[string]int),
			UserParticipation:  make(map[string]int),
		},
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		logger.Error("Failed to create feedback storage directory: %v", err)
	}

	// Load existing state
	if err := loop.loadState(); err != nil {
		logger.Warn("Failed to load existing feedback state: %v", err)
	}

	// Start background processing
	go loop.startFeedbackProcessing()

	return loop
}

func NewFeedbackAnalyzer() *FeedbackAnalyzer {
	return &FeedbackAnalyzer{
		patterns:          make(map[string]*FeedbackPattern),
		trendAnalysis:     NewTrendAnalysis(),
		anomalyDetection:  NewAnomalyDetection(),
		userBehaviorModel: NewUserBehaviorModel(),
		mutex:             sync.RWMutex{},
	}
}

func NewImpactCalculator() *ImpactCalculator {
	return &ImpactCalculator{
		impactHistory: make([]ImpactMeasurement, 0),
		baselineMetrics: BaselineMetrics{
			AccuracyScore:         0.7,
			LearningEffectiveness: 0.7,
			UserSatisfaction:      0.7,
			ResponseTime:          time.Second,
			LastUpdated:           time.Now(),
		},
		calculator: NewMetricsCalculator(),
		mutex:      sync.RWMutex{},
	}
}

func NewTrendAnalysis() *TrendAnalysis {
	return &TrendAnalysis{
		trends:        make(map[string]*Trend),
		trendDetector: NewTrendDetector(),
		mutex:         sync.RWMutex{},
	}
}

func NewAnomalyDetection() *AnomalyDetection {
	return &AnomalyDetection{
		anomalies:      make([]Anomaly, 0),
		detectionRules: make([]DetectionRule, 0),
		threshold:      0.8,
		mutex:          sync.RWMutex{},
	}
}

func NewUserBehaviorModel() *UserBehaviorModel {
	return &UserBehaviorModel{
		userModels:    make(map[string]*UserModel),
		behaviorRules: make([]BehaviorRule, 0),
		mutex:         sync.RWMutex{},
	}
}

func NewTrendDetector() *TrendDetector {
	return &TrendDetector{
		windowSize:       10,
		sensitivityLevel: 0.1,
		mutex:            sync.RWMutex{},
	}
}

func NewMetricsCalculator() *MetricsCalculator {
	return &MetricsCalculator{
		calculationRules: make(map[string]CalculationRule),
		mutex:            sync.RWMutex{},
	}
}

// SubmitFeedback submits user feedback for processing
func (fl *FeedbackLoop) SubmitFeedback(feedback *UserFeedback) error {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()

	// Validate feedback
	if err := fl.validateFeedback(feedback); err != nil {
		return fmt.Errorf("invalid feedback: %v", err)
	}

	// Generate ID if not provided
	if feedback.ID == "" {
		feedback.ID = fmt.Sprintf("feedback_%d", time.Now().UnixNano())
	}

	// Set timestamp if not provided
	if feedback.Timestamp.IsZero() {
		feedback.Timestamp = time.Now()
	}

	// Calculate confidence if not set
	if feedback.Confidence == 0 {
		feedback.Confidence = fl.calculateFeedbackConfidence(feedback)
	}

	// Store in history
	fl.feedbackHistory[feedback.ID] = feedback

	// Update user feedback profile
	fl.updateUserFeedbackProfile(feedback)

	// Update metrics
	fl.updateMetrics(feedback)

	// Add to queue for processing
	select {
	case fl.feedbackQueue <- feedback:
		fl.logger.Info("Feedback submitted for processing: %s (type: %s, severity: %s)",
			feedback.ID, feedback.FeedbackType, feedback.Severity)
	default:
		fl.logger.Warn("Feedback queue full, adding to processing queue: %s", feedback.ID)
		fl.processingQueue = append(fl.processingQueue, feedback)
	}

	// Process immediately if real-time processing is enabled
	if fl.config.EnableRealTimeProcessing {
		go fl.processFeedback(feedback)
	}

	return nil
}

// ProcessFeedback processes individual feedback and applies learning adjustments
func (fl *FeedbackLoop) processFeedback(feedback *UserFeedback) error {
	startTime := time.Now()

	fl.logger.Debug("Processing feedback: %s", feedback.ID)

	// Analyze feedback for patterns and anomalies
	patterns := fl.feedbackAnalyzer.AnalyzeFeedback(feedback)

	// Calculate impact before processing
	beforeMetrics := fl.impactCalculator.CaptureBaseline()

	// Apply learning adjustments based on feedback
	adjustments, err := fl.createLearningAdjustments(feedback, patterns)
	if err != nil {
		fl.logger.Error("Failed to create learning adjustments for feedback %s: %v", feedback.ID, err)
		return err
	}

	// Apply adjustments to learning models
	for _, adjustment := range adjustments {
		if err := fl.applyLearningAdjustment(adjustment); err != nil {
			fl.logger.Error("Failed to apply learning adjustment %s: %v", adjustment.ID, err)
			continue
		}
		fl.learningAdjustments[adjustment.ID] = adjustment
	}

	// Calculate impact after processing
	afterMetrics := fl.impactCalculator.CaptureBaseline()
	impact := fl.impactCalculator.CalculateImpact(beforeMetrics, afterMetrics, feedback.ID)

	// Update feedback with processing results
	fl.mutex.Lock()
	feedback.IsProcessed = true
	now := time.Now()
	feedback.ProcessedAt = &now
	feedback.Impact = &FeedbackImpact{
		ModelUpdated:       len(adjustments) > 0,
		PatternsAffected:   extractPatternIDs(patterns),
		ConfidenceChange:   impact.ImpactScore,
		LearningAdjustment: calculateLearningAdjustment(adjustments),
		ProcessingTime:     time.Since(startTime),
		ImpactScore:        impact.ImpactScore,
		AppliedAt:          time.Now(),
	}
	fl.mutex.Unlock()

	fl.logger.Info("Feedback processed: %s (impact: %.3f, adjustments: %d)",
		feedback.ID, impact.ImpactScore, len(adjustments))

	return nil
}

// GetFeedbackAnalytics returns analytics about feedback patterns and effectiveness
func (fl *FeedbackLoop) GetFeedbackAnalytics() *FeedbackAnalytics {
	fl.mutex.RLock()
	defer fl.mutex.RUnlock()

	analytics := &FeedbackAnalytics{
		TotalFeedbacks:      len(fl.feedbackHistory),
		ProcessedFeedbacks:  fl.metrics.ProcessedFeedbacks,
		PendingFeedbacks:    fl.metrics.PendingFeedbacks,
		UserProfiles:        len(fl.userFeedbackProfiles),
		LearningAdjustments: len(fl.learningAdjustments),
		FeedbackByType:      make(map[FeedbackType]int),
		FeedbackBySeverity:  make(map[FeedbackSeverity]int),
		TopPatterns:         fl.feedbackAnalyzer.GetTopPatterns(10),
		RecentTrends:        fl.feedbackAnalyzer.GetRecentTrends(7),
		AnomalyAlerts:       fl.feedbackAnalyzer.GetActiveAnomalies(),
		ImpactMetrics:       fl.impactCalculator.GetSummaryMetrics(),
		GeneratedAt:         time.Now(),
	}

	// Copy maps to avoid concurrent access issues
	for k, v := range fl.metrics.FeedbackByType {
		analytics.FeedbackByType[k] = v
	}
	for k, v := range fl.metrics.FeedbackBySeverity {
		analytics.FeedbackBySeverity[k] = v
	}

	return analytics
}

// GetUserFeedbackProfile returns a user's feedback profile
func (fl *FeedbackLoop) GetUserFeedbackProfile(userID string) (*UserFeedbackProfile, error) {
	fl.mutex.RLock()
	defer fl.mutex.RUnlock()

	profile, exists := fl.userFeedbackProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("user feedback profile not found: %s", userID)
	}

	return profile, nil
}

// UpdateUserExpertise updates a user's expertise areas based on feedback patterns
func (fl *FeedbackLoop) UpdateUserExpertise(userID string, expertiseAreas []string) error {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()

	profile, exists := fl.userFeedbackProfiles[userID]
	if !exists {
		return fmt.Errorf("user feedback profile not found: %s", userID)
	}

	profile.ExpertiseAreas = expertiseAreas
	profile.UpdatedAt = time.Now()

	return nil
}

// Helper methods

func (fl *FeedbackLoop) validateFeedback(feedback *UserFeedback) error {
	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if feedback.FeedbackType == "" {
		return fmt.Errorf("feedback type is required")
	}
	if feedback.TargetID == "" {
		return fmt.Errorf("target ID is required")
	}
	return nil
}

func (fl *FeedbackLoop) calculateFeedbackConfidence(feedback *UserFeedback) float64 {
	confidence := 0.5 // Base confidence

	// Adjust based on user profile
	if profile, exists := fl.userFeedbackProfiles[feedback.UserID]; exists {
		confidence = (confidence + profile.ReliabilityScore) / 2
	}

	// Adjust based on feedback type
	switch feedback.FeedbackType {
	case FeedbackCorrection:
		confidence += 0.2
	case FeedbackValidation:
		confidence += 0.15
	case FeedbackPositive, FeedbackNegative:
		confidence += 0.1
	}

	// Adjust based on severity
	switch feedback.Severity {
	case FeedbackSeverityCritical:
		confidence += 0.2
	case FeedbackSeverityHigh:
		confidence += 0.1
	}

	return math.Min(1.0, confidence)
}

func (fl *FeedbackLoop) updateUserFeedbackProfile(feedback *UserFeedback) {
	profile, exists := fl.userFeedbackProfiles[feedback.UserID]
	if !exists {
		profile = &UserFeedbackProfile{
			UserID:             feedback.UserID,
			FeedbackByType:     make(map[FeedbackType]int),
			FeedbackBySeverity: make(map[FeedbackSeverity]int),
			AccuracyScore:      0.7, // Start with moderate accuracy
			ReliabilityScore:   0.7, // Start with moderate reliability
			CreatedAt:          time.Now(),
			IsActive:           true,
		}
		fl.userFeedbackProfiles[feedback.UserID] = profile
	}

	// Update counters
	profile.TotalFeedbacks++
	profile.FeedbackByType[feedback.FeedbackType]++
	profile.FeedbackBySeverity[feedback.Severity]++

	// Update timing metrics
	now := time.Now()
	if !profile.LastFeedbackTime.IsZero() {
		timeDiff := now.Sub(profile.LastFeedbackTime)
		if profile.AverageResponseTime == 0 {
			profile.AverageResponseTime = timeDiff
		} else {
			profile.AverageResponseTime = (profile.AverageResponseTime + timeDiff) / 2
		}
	}
	profile.LastFeedbackTime = now
	profile.UpdatedAt = now

	// Calculate feedback velocity (feedbacks per day)
	daysSinceCreation := now.Sub(profile.CreatedAt).Hours() / 24
	if daysSinceCreation > 0 {
		profile.FeedbackVelocity = float64(profile.TotalFeedbacks) / daysSinceCreation
	}
}

func (fl *FeedbackLoop) updateMetrics(feedback *UserFeedback) {
	fl.metrics.TotalFeedbacks++
	fl.metrics.FeedbackByType[feedback.FeedbackType]++
	fl.metrics.FeedbackBySeverity[feedback.Severity]++
	fl.metrics.UserParticipation[feedback.UserID]++
	fl.metrics.PendingFeedbacks++
	fl.metrics.LastUpdated = time.Now()
}

func (fl *FeedbackLoop) createLearningAdjustments(feedback *UserFeedback, patterns []*FeedbackPattern) ([]*LearningAdjustment, error) {
	adjustments := make([]*LearningAdjustment, 0)

	// Create adjustment based on feedback type
	switch feedback.FeedbackType {
	case FeedbackCorrection:
		adj := fl.createCorrectionAdjustment(feedback)
		if adj != nil {
			adjustments = append(adjustments, adj)
		}
	case FeedbackValidation:
		adj := fl.createValidationAdjustment(feedback)
		if adj != nil {
			adjustments = append(adjustments, adj)
		}
	case FeedbackRejection:
		adj := fl.createRejectionAdjustment(feedback)
		if adj != nil {
			adjustments = append(adjustments, adj)
		}
	}

	// Create pattern-based adjustments
	for _, pattern := range patterns {
		if pattern.Actionable {
			adj := fl.createPatternAdjustment(feedback, pattern)
			if adj != nil {
				adjustments = append(adjustments, adj)
			}
		}
	}

	return adjustments, nil
}

func (fl *FeedbackLoop) createCorrectionAdjustment(feedback *UserFeedback) *LearningAdjustment {
	return &LearningAdjustment{
		ID:              fmt.Sprintf("adj_correction_%s_%d", feedback.ID, time.Now().UnixNano()),
		FeedbackID:      feedback.ID,
		AdjustmentType:  "correction",
		TargetComponent: "pattern",
		TargetID:        feedback.TargetID,
		Adjustment: map[string]interface{}{
			"original_value":  feedback.OriginalValue,
			"corrected_value": feedback.FeedbackValue,
			"confidence":      feedback.Confidence,
		},
		Confidence: feedback.Confidence,
		AppliedAt:  time.Now(),
		Impact:     fl.calculateAdjustmentImpact(feedback),
		Reversible: true,
		ReversalData: map[string]interface{}{
			"original_confidence": feedback.Confidence,
			"original_pattern":    feedback.OriginalValue,
		},
	}
}

func (fl *FeedbackLoop) createValidationAdjustment(feedback *UserFeedback) *LearningAdjustment {
	return &LearningAdjustment{
		ID:              fmt.Sprintf("adj_validation_%s_%d", feedback.ID, time.Now().UnixNano()),
		FeedbackID:      feedback.ID,
		AdjustmentType:  "validation",
		TargetComponent: "confidence",
		TargetID:        feedback.TargetID,
		Adjustment: map[string]interface{}{
			"confidence_boost":  0.1,
			"validation_source": feedback.UserID,
		},
		Confidence: feedback.Confidence,
		AppliedAt:  time.Now(),
		Impact:     fl.calculateAdjustmentImpact(feedback),
		Reversible: true,
	}
}

func (fl *FeedbackLoop) createRejectionAdjustment(feedback *UserFeedback) *LearningAdjustment {
	return &LearningAdjustment{
		ID:              fmt.Sprintf("adj_rejection_%s_%d", feedback.ID, time.Now().UnixNano()),
		FeedbackID:      feedback.ID,
		AdjustmentType:  "rejection",
		TargetComponent: "confidence",
		TargetID:        feedback.TargetID,
		Adjustment: map[string]interface{}{
			"confidence_penalty": -0.15,
			"rejection_reason":   feedback.Comment,
		},
		Confidence: feedback.Confidence,
		AppliedAt:  time.Now(),
		Impact:     fl.calculateAdjustmentImpact(feedback),
		Reversible: true,
	}
}

func (fl *FeedbackLoop) createPatternAdjustment(feedback *UserFeedback, pattern *FeedbackPattern) *LearningAdjustment {
	return &LearningAdjustment{
		ID:              fmt.Sprintf("adj_pattern_%s_%s_%d", feedback.ID, pattern.ID, time.Now().UnixNano()),
		FeedbackID:      feedback.ID,
		AdjustmentType:  "pattern",
		TargetComponent: "pattern",
		TargetID:        pattern.ID,
		Adjustment: map[string]interface{}{
			"pattern_confidence": pattern.Confidence,
			"pattern_frequency":  pattern.Frequency,
			"action":             pattern.RecommendedAction,
		},
		Confidence: pattern.Confidence,
		AppliedAt:  time.Now(),
		Impact:     pattern.Impact,
		Reversible: true,
	}
}

func (fl *FeedbackLoop) calculateAdjustmentImpact(feedback *UserFeedback) float64 {
	impact := 0.5 // Base impact

	// Adjust based on severity
	severityMultiplier := fl.config.SeverityMultipliers[feedback.Severity]
	impact *= severityMultiplier

	// Adjust based on user reliability
	if profile, exists := fl.userFeedbackProfiles[feedback.UserID]; exists {
		impact *= profile.ReliabilityScore
	}

	return math.Min(1.0, impact)
}

func (fl *FeedbackLoop) applyLearningAdjustment(adjustment *LearningAdjustment) error {
	switch adjustment.TargetComponent {
	case "pattern":
		return fl.applyPatternAdjustment(adjustment)
	case "confidence":
		return fl.applyConfidenceAdjustment(adjustment)
	case "embedding":
		return fl.applyEmbeddingAdjustment(adjustment)
	default:
		return fmt.Errorf("unknown adjustment target component: %s", adjustment.TargetComponent)
	}
}

func (fl *FeedbackLoop) applyPatternAdjustment(adjustment *LearningAdjustment) error {
	// Apply adjustment to correction processor patterns
	if fl.correctionProcessor != nil {
		fl.logger.Debug("Applied pattern adjustment: %s", adjustment.ID)
	}
	return nil
}

func (fl *FeedbackLoop) applyConfidenceAdjustment(adjustment *LearningAdjustment) error {
	// Apply confidence adjustments to RAG engine
	if fl.ragEngine != nil {
		fl.logger.Debug("Applied confidence adjustment: %s", adjustment.ID)
	}
	return nil
}

func (fl *FeedbackLoop) applyEmbeddingAdjustment(adjustment *LearningAdjustment) error {
	// Apply embedding adjustments to RAG engine
	if fl.ragEngine != nil {
		fl.logger.Debug("Applied embedding adjustment: %s", adjustment.ID)
	}
	return nil
}

// Background processing methods

func (fl *FeedbackLoop) startFeedbackProcessing() {
	ticker := time.NewTicker(fl.config.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case feedback := <-fl.feedbackQueue:
			if err := fl.processFeedback(feedback); err != nil {
				fl.logger.Error("Failed to process feedback %s: %v", feedback.ID, err)
			}
		case <-ticker.C:
			if fl.config.EnableBatchProcessing {
				fl.processBatchFeedback()
			}
		case <-fl.ctx.Done():
			fl.logger.Info("Stopping feedback processing")
			return
		}
	}
}

func (fl *FeedbackLoop) processBatchFeedback() {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()

	if len(fl.processingQueue) < fl.config.MinFeedbacksToProcess {
		return
	}

	batchSize := fl.config.MaxProcessingBatch
	if len(fl.processingQueue) < batchSize {
		batchSize = len(fl.processingQueue)
	}

	batch := fl.processingQueue[:batchSize]
	fl.processingQueue = fl.processingQueue[batchSize:]

	fl.logger.Info("Processing feedback batch: %d items", len(batch))

	for _, feedback := range batch {
		go func(fb *UserFeedback) {
			if err := fl.processFeedback(fb); err != nil {
				fl.logger.Error("Failed to process batch feedback %s: %v", fb.ID, err)
			}
		}(feedback)
	}
}

// Helper functions

func extractPatternIDs(patterns []*FeedbackPattern) []string {
	ids := make([]string, len(patterns))
	for i, pattern := range patterns {
		ids[i] = pattern.ID
	}
	return ids
}

func calculateLearningAdjustment(adjustments []*LearningAdjustment) float64 {
	if len(adjustments) == 0 {
		return 0.0
	}

	total := 0.0
	for _, adj := range adjustments {
		total += adj.Impact
	}
	return total / float64(len(adjustments))
}

// Additional types for analytics

type FeedbackAnalytics struct {
	TotalFeedbacks      int                      `json:"total_feedbacks"`
	ProcessedFeedbacks  int                      `json:"processed_feedbacks"`
	PendingFeedbacks    int                      `json:"pending_feedbacks"`
	UserProfiles        int                      `json:"user_profiles"`
	LearningAdjustments int                      `json:"learning_adjustments"`
	FeedbackByType      map[FeedbackType]int     `json:"feedback_by_type"`
	FeedbackBySeverity  map[FeedbackSeverity]int `json:"feedback_by_severity"`
	TopPatterns         []*FeedbackPattern       `json:"top_patterns"`
	RecentTrends        []*Trend                 `json:"recent_trends"`
	AnomalyAlerts       []*Anomaly               `json:"anomaly_alerts"`
	ImpactMetrics       *ImpactSummaryMetrics    `json:"impact_metrics"`
	GeneratedAt         time.Time                `json:"generated_at"`
}

type ImpactSummaryMetrics struct {
	AverageImpact    float64   `json:"average_impact"`
	TotalImpactScore float64   `json:"total_impact_score"`
	HighImpactCount  int       `json:"high_impact_count"`
	LowImpactCount   int       `json:"low_impact_count"`
	LastMeasurement  time.Time `json:"last_measurement"`
}

// Additional methods for FeedbackAnalyzer and ImpactCalculator

func (fa *FeedbackAnalyzer) GetTopPatterns(limit int) []*FeedbackPattern {
	fa.mutex.RLock()
	defer fa.mutex.RUnlock()

	patterns := make([]*FeedbackPattern, 0, len(fa.patterns))
	for _, pattern := range fa.patterns {
		patterns = append(patterns, pattern)
	}

	// Sort by frequency and confidence
	sort.Slice(patterns, func(i, j int) bool {
		scoreI := float64(patterns[i].Frequency) * patterns[i].Confidence
		scoreJ := float64(patterns[j].Frequency) * patterns[j].Confidence
		return scoreI > scoreJ
	})

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return patterns
}

func (fa *FeedbackAnalyzer) GetRecentTrends(days int) []*Trend {
	fa.mutex.RLock()
	defer fa.mutex.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	trends := make([]*Trend, 0)

	for _, trend := range fa.trendAnalysis.trends {
		if trend.StartTime.After(cutoff) {
			trends = append(trends, trend)
		}
	}

	// Sort by start time (most recent first)
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].StartTime.After(trends[j].StartTime)
	})

	return trends
}

func (fa *FeedbackAnalyzer) GetActiveAnomalies() []*Anomaly {
	fa.mutex.RLock()
	defer fa.mutex.RUnlock()

	active := make([]*Anomaly, 0)
	for i := range fa.anomalyDetection.anomalies {
		anomaly := &fa.anomalyDetection.anomalies[i]
		// Consider anomalies from last 24 hours as active
		if time.Since(anomaly.DetectedAt) < 24*time.Hour {
			active = append(active, anomaly)
		}
	}

	return active
}

func (ic *ImpactCalculator) GetSummaryMetrics() *ImpactSummaryMetrics {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	if len(ic.impactHistory) == 0 {
		return &ImpactSummaryMetrics{
			LastMeasurement: time.Now(),
		}
	}

	totalImpact := 0.0
	highImpactCount := 0
	lowImpactCount := 0

	for _, measurement := range ic.impactHistory {
		totalImpact += measurement.ImpactScore
		if measurement.ImpactScore > 0.7 {
			highImpactCount++
		} else if measurement.ImpactScore < 0.3 {
			lowImpactCount++
		}
	}

	avgImpact := totalImpact / float64(len(ic.impactHistory))
	lastMeasurement := ic.impactHistory[len(ic.impactHistory)-1].MeasuredAt

	return &ImpactSummaryMetrics{
		AverageImpact:    avgImpact,
		TotalImpactScore: totalImpact,
		HighImpactCount:  highImpactCount,
		LowImpactCount:   lowImpactCount,
		LastMeasurement:  lastMeasurement,
	}
}

// Shutdown gracefully shuts down the feedback loop
func (fl *FeedbackLoop) Shutdown() error {
	fl.logger.Info("Shutting down FeedbackLoop")

	fl.cancel()

	// Process remaining feedback in queue
	close(fl.feedbackQueue)
	for feedback := range fl.feedbackQueue {
		if err := fl.processFeedback(feedback); err != nil {
			fl.logger.Error("Failed to process feedback during shutdown: %v", err)
		}
	}

	// Save final state
	if err := fl.saveState(); err != nil {
		fl.logger.Error("Failed to save final feedback loop state: %v", err)
	}

	fl.logger.Info("Feedback loop shutdown completed")
	return nil
}

func (fl *FeedbackLoop) saveState() error {
	statePath := filepath.Join(fl.config.StoragePath, "feedback_state.json")
	tempPath := statePath + ".tmp"

	// Create state snapshot
	state := struct {
		FeedbackHistory      map[string]*UserFeedback        `json:"feedback_history"`
		UserFeedbackProfiles map[string]*UserFeedbackProfile `json:"user_feedback_profiles"`
		LearningAdjustments  map[string]*LearningAdjustment  `json:"learning_adjustments"`
		Metrics              FeedbackMetrics                 `json:"metrics"`
		Timestamp            time.Time                       `json:"timestamp"`
	}{
		FeedbackHistory:      fl.feedbackHistory,
		UserFeedbackProfiles: fl.userFeedbackProfiles,
		LearningAdjustments:  fl.learningAdjustments,
		Metrics:              fl.metrics,
		Timestamp:            time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, statePath); err != nil {
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	fl.logger.Debug("Feedback loop state saved successfully")
	return nil
}

// loadState loads the feedback loop state from disk
func (fl *FeedbackLoop) loadState() error {
	statePath := filepath.Join(fl.config.StoragePath, "feedback_state.json")

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		fl.logger.Info("No existing feedback state found, starting fresh")
		return nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse state
	var state struct {
		FeedbackHistory      map[string]*UserFeedback        `json:"feedback_history"`
		UserFeedbackProfiles map[string]*UserFeedbackProfile `json:"user_feedback_profiles"`
		LearningAdjustments  map[string]*LearningAdjustment  `json:"learning_adjustments"`
		Metrics              FeedbackMetrics                 `json:"metrics"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Restore state
	fl.feedbackHistory = state.FeedbackHistory
	fl.userFeedbackProfiles = state.UserFeedbackProfiles
	fl.learningAdjustments = state.LearningAdjustments
	fl.metrics = state.Metrics

	fl.logger.Info("Feedback loop state loaded successfully")
	return nil
}

// AnalyzeFeedback analyzes feedback for patterns and anomalies
func (fa *FeedbackAnalyzer) AnalyzeFeedback(feedback *UserFeedback) []*FeedbackPattern {
	fa.mutex.Lock()
	defer fa.mutex.Unlock()

	patterns := make([]*FeedbackPattern, 0)

	// Pattern 1: User correction frequency
	if feedback.FeedbackType == FeedbackCorrection {
		patternID := fmt.Sprintf("correction_frequency_%s", feedback.UserID)
		pattern, exists := fa.patterns[patternID]
		if !exists {
			pattern = &FeedbackPattern{
				ID:          patternID,
				PatternType: "correction_frequency",
				Description: fmt.Sprintf("Correction frequency pattern for user %s", feedback.UserID),
				Frequency:   1,
				Confidence:  0.5,
				Context: map[string]interface{}{
					"user_id": feedback.UserID,
				},
				FirstSeen:         time.Now(),
				LastSeen:          time.Now(),
				Impact:            0.3,
				Actionable:        true,
				RecommendedAction: "review_user_training",
			}
			fa.patterns[patternID] = pattern
		} else {
			pattern.Frequency++
			pattern.LastSeen = time.Now()
			pattern.Confidence = math.Min(1.0, pattern.Confidence+0.1)

			// High frequency corrections might indicate training needs
			if pattern.Frequency > 10 {
				pattern.Impact = 0.8
				pattern.RecommendedAction = "provide_additional_training"
			}
		}
		patterns = append(patterns, pattern)
	}

	// Pattern 2: Field-specific validation issues
	if feedback.TargetType == "field" && feedback.FeedbackType == FeedbackValidation {
		patternID := fmt.Sprintf("field_validation_%s", feedback.TargetID)
		pattern, exists := fa.patterns[patternID]
		if !exists {
			pattern = &FeedbackPattern{
				ID:          patternID,
				PatternType: "field_validation",
				Description: fmt.Sprintf("Validation issues for field %s", feedback.TargetID),
				Frequency:   1,
				Confidence:  0.4,
				Context: map[string]interface{}{
					"field_id": feedback.TargetID,
				},
				FirstSeen:         time.Now(),
				LastSeen:          time.Now(),
				Impact:            0.5,
				Actionable:        true,
				RecommendedAction: "review_field_validation_rules",
			}
			fa.patterns[patternID] = pattern
		} else {
			pattern.Frequency++
			pattern.LastSeen = time.Now()
			pattern.Confidence = math.Min(1.0, pattern.Confidence+0.15)
		}
		patterns = append(patterns, pattern)
	}

	// Pattern 3: Negative feedback trends
	if feedback.FeedbackType == FeedbackNegative || feedback.FeedbackType == FeedbackRejection {
		patternID := "negative_feedback_trend"
		pattern, exists := fa.patterns[patternID]
		if !exists {
			pattern = &FeedbackPattern{
				ID:                patternID,
				PatternType:       "negative_trend",
				Description:       "Increasing negative feedback trend",
				Frequency:         1,
				Confidence:        0.3,
				Context:           map[string]interface{}{},
				FirstSeen:         time.Now(),
				LastSeen:          time.Now(),
				Impact:            0.6,
				Actionable:        true,
				RecommendedAction: "investigate_system_issues",
			}
			fa.patterns[patternID] = pattern
		} else {
			pattern.Frequency++
			pattern.LastSeen = time.Now()
			pattern.Confidence = math.Min(1.0, pattern.Confidence+0.1)

			// Alert if negative feedback is increasing rapidly
			if pattern.Frequency > 5 {
				pattern.Impact = 0.9
				pattern.RecommendedAction = "immediate_system_review"
			}
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// CaptureBaseline captures current baseline metrics
func (ic *ImpactCalculator) CaptureBaseline() BaselineMetrics {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	// Return current baseline metrics
	baseline := ic.baselineMetrics
	baseline.LastUpdated = time.Now()

	return baseline
}

// CalculateImpact calculates the impact between before and after metrics
func (ic *ImpactCalculator) CalculateImpact(before BaselineMetrics, after BaselineMetrics, feedbackID string) ImpactMeasurement {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	// Calculate differences
	accuracyDiff := after.AccuracyScore - before.AccuracyScore
	learningDiff := after.LearningEffectiveness - before.LearningEffectiveness
	satisfactionDiff := after.UserSatisfaction - before.UserSatisfaction
	timeDiff := after.ResponseTime - before.ResponseTime

	// Calculate overall impact score
	impactScore := (accuracyDiff*0.4 + learningDiff*0.3 + satisfactionDiff*0.2) - (timeDiff.Seconds() * 0.1)

	// Normalize impact score to [-1, 1] range
	impactScore = math.Max(-1.0, math.Min(1.0, impactScore))

	// Create impact measurement
	measurement := ImpactMeasurement{
		ID:              fmt.Sprintf("impact_%s_%d", feedbackID, time.Now().UnixNano()),
		FeedbackID:      feedbackID,
		MetricType:      "overall",
		BeforeValue:     before.AccuracyScore,
		AfterValue:      after.AccuracyScore,
		ImpactScore:     impactScore,
		MeasuredAt:      time.Now(),
		ConfidenceLevel: 0.8,
	}

	// Store in history
	ic.impactHistory = append(ic.impactHistory, measurement)

	// Keep only recent measurements (last 1000)
	if len(ic.impactHistory) > 1000 {
		ic.impactHistory = ic.impactHistory[len(ic.impactHistory)-1000:]
	}

	// Update baseline metrics if impact is positive
	if impactScore > 0 {
		ic.baselineMetrics.AccuracyScore = (ic.baselineMetrics.AccuracyScore + after.AccuracyScore) / 2
		ic.baselineMetrics.LearningEffectiveness = (ic.baselineMetrics.LearningEffectiveness + after.LearningEffectiveness) / 2
		ic.baselineMetrics.UserSatisfaction = (ic.baselineMetrics.UserSatisfaction + after.UserSatisfaction) / 2
		ic.baselineMetrics.LastUpdated = time.Now()
	}

	return measurement
}
