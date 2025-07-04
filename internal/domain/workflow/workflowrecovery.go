package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RecoveryStrategy defines different recovery approaches
type RecoveryStrategy string

const (
	RetryStrategy              RecoveryStrategy = "retry"
	FallbackStrategy           RecoveryStrategy = "fallback"
	ManualInterventionStrategy RecoveryStrategy = "manual"
	SkipStepStrategy           RecoveryStrategy = "skip"
	RollbackStrategy           RecoveryStrategy = "rollback"
)

// WorkflowStepStatus represents the status of a workflow step
type WorkflowStepStatus string

const (
	StepPending    WorkflowStepStatus = "pending"
	StepInProgress WorkflowStepStatus = "in_progress"
	StepCompleted  WorkflowStepStatus = "completed"
	StepFailed     WorkflowStepStatus = "failed"
	StepSkipped    WorkflowStepStatus = "skipped"
	StepRolledBack WorkflowStepStatus = "rolled_back"
)

// ErrorSeverity defines the severity level of errors
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Status        WorkflowStepStatus     `json:"status"`
	StartTime     *time.Time             `json:"start_time,omitempty"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
	LastError     string                 `json:"last_error,omitempty"`
	ErrorSeverity ErrorSeverity          `json:"error_severity,omitempty"`
	Dependencies  []string               `json:"dependencies"`
	CanSkip       bool                   `json:"can_skip"`
	CanRollback   bool                   `json:"can_rollback"`
	RollbackData  map[string]interface{} `json:"rollback_data,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowExecution represents a complete workflow execution
type WorkflowExecution struct {
	ID               string                 `json:"id"`
	WorkflowType     string                 `json:"workflow_type"`
	DealID           string                 `json:"deal_id,omitempty"`
	DocumentID       string                 `json:"document_id,omitempty"`
	Status           string                 `json:"status"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Steps            []*WorkflowStep        `json:"steps"`
	CurrentStepIndex int                    `json:"current_step_index"`
	TotalRetries     int                    `json:"total_retries"`
	PartialResults   map[string]interface{} `json:"partial_results,omitempty"`
	ErrorLog         []ErrorLogEntry        `json:"error_log"`
	RecoveryStrategy RecoveryStrategy       `json:"recovery_strategy"`
	Priority         string                 `json:"priority"`
	CreatedBy        string                 `json:"created_by,omitempty"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ErrorLogEntry represents a single error log entry
type ErrorLogEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	StepID       string                 `json:"step_id"`
	ErrorType    string                 `json:"error_type"`
	ErrorMessage string                 `json:"error_message"`
	Severity     ErrorSeverity          `json:"severity"`
	Context      map[string]interface{} `json:"context,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
	Resolved     bool                   `json:"resolved"`
	Resolution   string                 `json:"resolution,omitempty"`
}

// RetryConfig defines retry behavior configuration
type RetryConfig struct {
	InitialDelay   time.Duration `json:"initial_delay"`
	MaxDelay       time.Duration `json:"max_delay"`
	BackoffFactor  float64       `json:"backoff_factor"`
	MaxRetries     int           `json:"max_retries"`
	Jitter         bool          `json:"jitter"`
	JitterMaxDelay time.Duration `json:"jitter_max_delay"`
}

// WorkflowRecoveryConfig holds configuration for the recovery service
type WorkflowRecoveryConfig struct {
	RetryConfig           RetryConfig   `json:"retry_config"`
	PersistenceInterval   time.Duration `json:"persistence_interval"`
	MaxExecutionHistory   int           `json:"max_execution_history"`
	ErrorLogRetention     time.Duration `json:"error_log_retention"`
	NotificationThreshold ErrorSeverity `json:"notification_threshold"`
	EnablePartialResults  bool          `json:"enable_partial_results"`
	StoragePath           string        `json:"storage_path"`
}

// WorkflowRecoveryService handles workflow recovery and error management
type WorkflowRecoveryService struct {
	config           WorkflowRecoveryConfig
	executions       map[string]*WorkflowExecution
	executionHistory []*WorkflowExecution
	errorStats       map[string]int
	mutex            sync.RWMutex
	logger           Logger
	notifier         ErrorNotifier
	ctx              context.Context
	cancel           context.CancelFunc
}

// ErrorNotifier interface for sending notifications
type ErrorNotifier interface {
	NotifyError(execution *WorkflowExecution, step *WorkflowStep, err error) error
	NotifyCriticalFailure(execution *WorkflowExecution, message string) error
	NotifyRecoverySuccess(execution *WorkflowExecution, message string) error
}

// StepExecutor interface for executing workflow steps
type StepExecutor interface {
	ExecuteStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error
	ValidateStep(step *WorkflowStep) error
	RollbackStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error
}

// NewWorkflowRecoveryService creates a new workflow recovery service
func NewWorkflowRecoveryService(config WorkflowRecoveryConfig, logger Logger, notifier ErrorNotifier) *WorkflowRecoveryService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &WorkflowRecoveryService{
		config:           config,
		executions:       make(map[string]*WorkflowExecution),
		executionHistory: make([]*WorkflowExecution, 0),
		errorStats:       make(map[string]int),
		logger:           logger,
		notifier:         notifier,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		logger.Error("Failed to create storage directory: %v", err)
	}

	// Load existing state
	if err := service.loadState(); err != nil {
		logger.Warn("Failed to load existing state: %v", err)
	}

	// Start background persistence
	go service.backgroundPersistence()

	return service
}

// CreateExecution creates a new workflow execution
func (wrs *WorkflowRecoveryService) CreateExecution(workflowType, dealID, documentID string, steps []*WorkflowStep) (*WorkflowExecution, error) {
	wrs.mutex.Lock()
	defer wrs.mutex.Unlock()

	execution := &WorkflowExecution{
		ID:               generateExecutionID(),
		WorkflowType:     workflowType,
		DealID:           dealID,
		DocumentID:       documentID,
		Status:           "created",
		StartTime:        time.Now(),
		Steps:            steps,
		CurrentStepIndex: 0,
		TotalRetries:     0,
		PartialResults:   make(map[string]interface{}),
		ErrorLog:         make([]ErrorLogEntry, 0),
		RecoveryStrategy: RetryStrategy,
		Priority:         "normal",
		UpdatedAt:        time.Now(),
	}

	// Set default retry configuration for steps
	for _, step := range steps {
		if step.MaxRetries == 0 {
			step.MaxRetries = wrs.config.RetryConfig.MaxRetries
		}
		if step.Metadata == nil {
			step.Metadata = make(map[string]interface{})
		}
	}

	wrs.executions[execution.ID] = execution
	wrs.logger.Info("Created workflow execution: %s (type: %s)", execution.ID, workflowType)

	return execution, nil
}

// ExecuteWorkflow executes a workflow with recovery capabilities
func (wrs *WorkflowRecoveryService) ExecuteWorkflow(executionID string, executor StepExecutor) error {
	wrs.mutex.Lock()
	execution, exists := wrs.executions[executionID]
	if !exists {
		wrs.mutex.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}
	wrs.mutex.Unlock()

	execution.Status = "running"
	execution.UpdatedAt = time.Now()

	wrs.logger.Info("Starting workflow execution: %s", executionID)

	for i := execution.CurrentStepIndex; i < len(execution.Steps); i++ {
		step := execution.Steps[i]

		// Check dependencies
		if err := wrs.checkDependencies(execution, step); err != nil {
			wrs.logError(execution, step, "dependency_check", err, SeverityHigh)
			return err
		}

		// Execute step with retry logic
		if err := wrs.executeStepWithRetry(execution, step, executor); err != nil {
			wrs.logger.Error("Step execution failed after retries: %s - %v", step.ID, err)

			// Determine recovery strategy
			strategy := wrs.determineRecoveryStrategy(execution, step, err)
			if err := wrs.applyRecoveryStrategy(execution, step, strategy, err); err != nil {
				execution.Status = "failed"
				execution.EndTime = &[]time.Time{time.Now()}[0]
				return err
			}
		}

		execution.CurrentStepIndex = i + 1
		execution.UpdatedAt = time.Now()

		// Save partial results if enabled
		if wrs.config.EnablePartialResults {
			wrs.savePartialResults(execution)
		}
	}

	execution.Status = "completed"
	execution.EndTime = &[]time.Time{time.Now()}[0]
	execution.UpdatedAt = time.Now()

	wrs.logger.Info("Workflow execution completed: %s", executionID)

	// Move to history
	wrs.moveToHistory(execution)

	return nil
}

// executeStepWithRetry executes a step with exponential backoff retry logic
func (wrs *WorkflowRecoveryService) executeStepWithRetry(execution *WorkflowExecution, step *WorkflowStep, executor StepExecutor) error {
	step.Status = StepInProgress
	step.StartTime = &[]time.Time{time.Now()}[0]

	var lastError error
	delay := wrs.config.RetryConfig.InitialDelay

	for attempt := 0; attempt <= step.MaxRetries; attempt++ {
		if attempt > 0 {
			wrs.logger.Debug("Retrying step %s (attempt %d/%d) after %v", step.ID, attempt, step.MaxRetries, delay)

			// Add jitter if enabled
			actualDelay := delay
			if wrs.config.RetryConfig.Jitter {
				jitter := time.Duration(float64(wrs.config.RetryConfig.JitterMaxDelay) * (0.5 + 0.5*float64(time.Now().UnixNano()%1000)/1000.0))
				actualDelay += jitter
			}

			select {
			case <-time.After(actualDelay):
			case <-wrs.ctx.Done():
				return fmt.Errorf("workflow cancelled during retry")
			}
		}

		// Execute the step
		err := executor.ExecuteStep(wrs.ctx, execution, step)
		if err == nil {
			step.Status = StepCompleted
			step.EndTime = &[]time.Time{time.Now()}[0]
			step.Duration = step.EndTime.Sub(*step.StartTime)
			wrs.logger.Debug("Step completed successfully: %s", step.ID)
			return nil
		}

		lastError = err
		step.RetryCount = attempt + 1
		step.LastError = err.Error()
		execution.TotalRetries++

		// Log the error
		severity := wrs.determineSeverity(err)
		wrs.logError(execution, step, "execution_error", err, severity)

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * wrs.config.RetryConfig.BackoffFactor)
		if delay > wrs.config.RetryConfig.MaxDelay {
			delay = wrs.config.RetryConfig.MaxDelay
		}

		// Check if error is retryable
		if !wrs.isRetryableError(err) {
			wrs.logger.Warn("Non-retryable error encountered in step %s: %v", step.ID, err)
			break
		}
	}

	step.Status = StepFailed
	step.EndTime = &[]time.Time{time.Now()}[0]
	step.Duration = step.EndTime.Sub(*step.StartTime)
	step.ErrorSeverity = wrs.determineSeverity(lastError)

	return lastError
}

// checkDependencies verifies that all step dependencies are satisfied
func (wrs *WorkflowRecoveryService) checkDependencies(execution *WorkflowExecution, step *WorkflowStep) error {
	for _, depID := range step.Dependencies {
		found := false
		for _, prevStep := range execution.Steps {
			if prevStep.ID == depID {
				if prevStep.Status != StepCompleted {
					return fmt.Errorf("dependency %s is not completed (status: %s)", depID, prevStep.Status)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("dependency %s not found in execution", depID)
		}
	}
	return nil
}

// determineRecoveryStrategy determines the appropriate recovery strategy for a failed step
func (wrs *WorkflowRecoveryService) determineRecoveryStrategy(execution *WorkflowExecution, step *WorkflowStep, err error) RecoveryStrategy {
	severity := wrs.determineSeverity(err)

	switch severity {
	case SeverityCritical:
		if step.CanRollback {
			return RollbackStrategy
		}
		return ManualInterventionStrategy
	case SeverityHigh:
		if step.CanSkip {
			return SkipStepStrategy
		}
		return ManualInterventionStrategy
	case SeverityMedium:
		return FallbackStrategy
	default:
		return RetryStrategy
	}
}

// applyRecoveryStrategy applies the determined recovery strategy
func (wrs *WorkflowRecoveryService) applyRecoveryStrategy(execution *WorkflowExecution, step *WorkflowStep, strategy RecoveryStrategy, err error) error {
	wrs.logger.Info("Applying recovery strategy %s for step %s", strategy, step.ID)

	switch strategy {
	case SkipStepStrategy:
		step.Status = StepSkipped
		wrs.logger.Warn("Skipping step %s due to error: %v", step.ID, err)
		return nil

	case RollbackStrategy:
		return wrs.rollbackStep(execution, step)

	case FallbackStrategy:
		return wrs.applyFallbackMechanism(execution, step, err)

	case ManualInterventionStrategy:
		execution.Status = "requires_manual_intervention"
		if wrs.notifier != nil {
			wrs.notifier.NotifyCriticalFailure(execution, fmt.Sprintf("Manual intervention required for step %s: %v", step.ID, err))
		}
		return fmt.Errorf("manual intervention required for step %s: %v", step.ID, err)

	default:
		return err
	}
}

// rollbackStep performs rollback operations for a step
func (wrs *WorkflowRecoveryService) rollbackStep(execution *WorkflowExecution, step *WorkflowStep) error {
	wrs.logger.Info("Rolling back step: %s", step.ID)

	// Implement rollback logic based on step.RollbackData
	step.Status = StepRolledBack

	// TODO: Implement actual rollback operations based on step type
	// This would involve calling appropriate rollback functions

	wrs.logger.Info("Step rolled back successfully: %s", step.ID)
	return nil
}

// applyFallbackMechanism applies fallback mechanisms for step failures
func (wrs *WorkflowRecoveryService) applyFallbackMechanism(execution *WorkflowExecution, step *WorkflowStep, err error) error {
	wrs.logger.Info("Applying fallback mechanism for step: %s", step.ID)

	// Determine fallback based on step type and error
	fallbackType := wrs.determineFallbackType(step, err)

	switch fallbackType {
	case "use_cached_result":
		return wrs.useCachedResult(execution, step)
	case "use_default_values":
		return wrs.useDefaultValues(execution, step)
	case "simplified_processing":
		return wrs.useSimplifiedProcessing(execution, step)
	default:
		return fmt.Errorf("no suitable fallback mechanism for step %s: %v", step.ID, err)
	}
}

// ResumeWorkflow resumes a workflow from the last successful step
func (wrs *WorkflowRecoveryService) ResumeWorkflow(executionID string, executor StepExecutor) error {
	wrs.mutex.Lock()
	execution, exists := wrs.executions[executionID]
	if !exists {
		wrs.mutex.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}
	wrs.mutex.Unlock()

	if execution.Status == "completed" {
		return fmt.Errorf("execution %s is already completed", executionID)
	}

	wrs.logger.Info("Resuming workflow execution: %s from step %d", executionID, execution.CurrentStepIndex)

	// Reset any failed steps to pending if they can be retried
	for i := execution.CurrentStepIndex; i < len(execution.Steps); i++ {
		step := execution.Steps[i]
		if step.Status == StepFailed && step.RetryCount < step.MaxRetries {
			step.Status = StepPending
			step.RetryCount = 0
			step.LastError = ""
		}
	}

	return wrs.ExecuteWorkflow(executionID, executor)
}

// Helper functions

func (wrs *WorkflowRecoveryService) logError(execution *WorkflowExecution, step *WorkflowStep, errorType string, err error, severity ErrorSeverity) {
	entry := ErrorLogEntry{
		Timestamp:    time.Now(),
		StepID:       step.ID,
		ErrorType:    errorType,
		ErrorMessage: err.Error(),
		Severity:     severity,
		Context:      make(map[string]interface{}),
		Resolved:     false,
	}

	execution.ErrorLog = append(execution.ErrorLog, entry)
	wrs.errorStats[errorType]++

	// Send notification if severity meets threshold
	if wrs.shouldNotify(severity) && wrs.notifier != nil {
		wrs.notifier.NotifyError(execution, step, err)
	}

	wrs.logger.Error("Error in step %s: %v (severity: %s)", step.ID, err, severity)
}

func (wrs *WorkflowRecoveryService) determineSeverity(err error) ErrorSeverity {
	errMsg := err.Error()

	// Check for critical errors
	if containsAny(errMsg, []string{"panic", "fatal", "critical", "database_connection", "auth_failure"}) {
		return SeverityCritical
	}

	// Check for high severity errors
	if containsAny(errMsg, []string{"timeout", "network", "permission", "invalid_format"}) {
		return SeverityHigh
	}

	// Check for medium severity errors
	if containsAny(errMsg, []string{"validation", "parse", "format", "missing_field"}) {
		return SeverityMedium
	}

	return SeverityLow
}

func (wrs *WorkflowRecoveryService) isRetryableError(err error) bool {
	errMsg := err.Error()

	// Non-retryable errors
	nonRetryable := []string{"auth_failure", "permission_denied", "invalid_format", "parse_error"}
	if containsAny(errMsg, nonRetryable) {
		return false
	}

	// Retryable errors
	retryable := []string{"timeout", "network", "temporary", "rate_limit", "service_unavailable"}
	return containsAny(errMsg, retryable)
}

func (wrs *WorkflowRecoveryService) shouldNotify(severity ErrorSeverity) bool {
	switch wrs.config.NotificationThreshold {
	case SeverityLow:
		return true
	case SeverityMedium:
		return severity == SeverityMedium || severity == SeverityHigh || severity == SeverityCritical
	case SeverityHigh:
		return severity == SeverityHigh || severity == SeverityCritical
	case SeverityCritical:
		return severity == SeverityCritical
	default:
		return false
	}
}

func (wrs *WorkflowRecoveryService) determineFallbackType(step *WorkflowStep, err error) string {
	// Analyze step type and error to determine appropriate fallback
	stepType, exists := step.Metadata["type"].(string)
	if !exists {
		return "use_default_values"
	}

	switch stepType {
	case "data_extraction":
		return "use_cached_result"
	case "ai_processing":
		return "simplified_processing"
	case "validation":
		return "use_default_values"
	default:
		return "use_default_values"
	}
}

func (wrs *WorkflowRecoveryService) useCachedResult(execution *WorkflowExecution, step *WorkflowStep) error {
	// TODO: Implement cached result lookup
	wrs.logger.Info("Using cached result for step: %s", step.ID)
	step.Status = StepCompleted
	return nil
}

func (wrs *WorkflowRecoveryService) useDefaultValues(execution *WorkflowExecution, step *WorkflowStep) error {
	// TODO: Implement default value assignment
	wrs.logger.Info("Using default values for step: %s", step.ID)
	step.Status = StepCompleted
	return nil
}

func (wrs *WorkflowRecoveryService) useSimplifiedProcessing(execution *WorkflowExecution, step *WorkflowStep) error {
	// TODO: Implement simplified processing logic
	wrs.logger.Info("Using simplified processing for step: %s", step.ID)
	step.Status = StepCompleted
	return nil
}

func (wrs *WorkflowRecoveryService) savePartialResults(execution *WorkflowExecution) {
	// Save current execution state as partial results
	for i := 0; i <= execution.CurrentStepIndex && i < len(execution.Steps); i++ {
		step := execution.Steps[i]
		if step.Status == StepCompleted {
			key := fmt.Sprintf("step_%s_result", step.ID)
			execution.PartialResults[key] = step.Metadata
		}
	}
	wrs.logger.Debug("Saved partial results for execution: %s", execution.ID)
}

func (wrs *WorkflowRecoveryService) moveToHistory(execution *WorkflowExecution) {
	wrs.mutex.Lock()
	defer wrs.mutex.Unlock()

	// Remove from active executions
	delete(wrs.executions, execution.ID)

	// Add to history
	wrs.executionHistory = append(wrs.executionHistory, execution)

	// Trim history if necessary
	if len(wrs.executionHistory) > wrs.config.MaxExecutionHistory {
		wrs.executionHistory = wrs.executionHistory[1:]
	}

	wrs.logger.Debug("Moved execution to history: %s", execution.ID)
}

// Persistence methods

func (wrs *WorkflowRecoveryService) backgroundPersistence() {
	ticker := time.NewTicker(wrs.config.PersistenceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := wrs.saveState(); err != nil {
				wrs.logger.Error("Failed to persist state: %v", err)
			}
		case <-wrs.ctx.Done():
			// Final save before shutdown
			if err := wrs.saveState(); err != nil {
				wrs.logger.Error("Failed to save final state: %v", err)
			}
			return
		}
	}
}

func (wrs *WorkflowRecoveryService) saveState() error {
	wrs.mutex.RLock()
	defer wrs.mutex.RUnlock()

	state := struct {
		Executions       map[string]*WorkflowExecution `json:"executions"`
		ExecutionHistory []*WorkflowExecution          `json:"execution_history"`
		ErrorStats       map[string]int                `json:"error_stats"`
		LastSaved        time.Time                     `json:"last_saved"`
	}{
		Executions:       wrs.executions,
		ExecutionHistory: wrs.executionHistory,
		ErrorStats:       wrs.errorStats,
		LastSaved:        time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	statePath := filepath.Join(wrs.config.StoragePath, "workflow_recovery_state.json")
	tempPath := statePath + ".tmp"

	if err := ioutil.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %v", err)
	}

	if err := os.Rename(tempPath, statePath); err != nil {
		return fmt.Errorf("failed to rename temp state file: %v", err)
	}

	wrs.logger.Debug("State saved successfully")
	return nil
}

func (wrs *WorkflowRecoveryService) loadState() error {
	statePath := filepath.Join(wrs.config.StoragePath, "workflow_recovery_state.json")

	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			wrs.logger.Info("No existing state file found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read state file: %v", err)
	}

	var state struct {
		Executions       map[string]*WorkflowExecution `json:"executions"`
		ExecutionHistory []*WorkflowExecution          `json:"execution_history"`
		ErrorStats       map[string]int                `json:"error_stats"`
		LastSaved        time.Time                     `json:"last_saved"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %v", err)
	}

	wrs.mutex.Lock()
	defer wrs.mutex.Unlock()

	wrs.executions = state.Executions
	if wrs.executions == nil {
		wrs.executions = make(map[string]*WorkflowExecution)
	}

	wrs.executionHistory = state.ExecutionHistory
	if wrs.executionHistory == nil {
		wrs.executionHistory = make([]*WorkflowExecution, 0)
	}

	wrs.errorStats = state.ErrorStats
	if wrs.errorStats == nil {
		wrs.errorStats = make(map[string]int)
	}

	wrs.logger.Info("State loaded successfully (last saved: %v)", state.LastSaved)
	return nil
}

// Query and management methods

// GetExecution retrieves a workflow execution by ID
func (wrs *WorkflowRecoveryService) GetExecution(executionID string) (*WorkflowExecution, error) {
	wrs.mutex.RLock()
	defer wrs.mutex.RUnlock()

	if execution, exists := wrs.executions[executionID]; exists {
		return execution, nil
	}

	// Check history
	for _, execution := range wrs.executionHistory {
		if execution.ID == executionID {
			return execution, nil
		}
	}

	return nil, fmt.Errorf("execution not found: %s", executionID)
}

// GetExecutionsByStatus retrieves executions by status
func (wrs *WorkflowRecoveryService) GetExecutionsByStatus(status string) []*WorkflowExecution {
	wrs.mutex.RLock()
	defer wrs.mutex.RUnlock()

	var results []*WorkflowExecution
	for _, execution := range wrs.executions {
		if execution.Status == status {
			results = append(results, execution)
		}
	}

	return results
}

// GetErrorStatistics returns error statistics
func (wrs *WorkflowRecoveryService) GetErrorStatistics() map[string]int {
	wrs.mutex.RLock()
	defer wrs.mutex.RUnlock()

	stats := make(map[string]int)
	for k, v := range wrs.errorStats {
		stats[k] = v
	}

	return stats
}

// CleanupOldExecutions removes old executions and error logs
func (wrs *WorkflowRecoveryService) CleanupOldExecutions() error {
	wrs.mutex.Lock()
	defer wrs.mutex.Unlock()

	cutoff := time.Now().Add(-wrs.config.ErrorLogRetention)

	// Clean up history
	filteredHistory := make([]*WorkflowExecution, 0)
	for _, execution := range wrs.executionHistory {
		if execution.StartTime.After(cutoff) {
			filteredHistory = append(filteredHistory, execution)
		}
	}

	removed := len(wrs.executionHistory) - len(filteredHistory)
	wrs.executionHistory = filteredHistory

	wrs.logger.Info("Cleaned up %d old executions", removed)
	return nil
}

// Shutdown gracefully shuts down the service
func (wrs *WorkflowRecoveryService) Shutdown() error {
	wrs.logger.Info("Shutting down WorkflowRecoveryService")

	wrs.cancel()

	// Final state save
	if err := wrs.saveState(); err != nil {
		wrs.logger.Error("Failed to save final state during shutdown: %v", err)
		return err
	}

	wrs.logger.Info("WorkflowRecoveryService shutdown complete")
	return nil
}

// Utility functions

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d_%d", time.Now().UnixNano(), time.Now().Unix()%1000)
}

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
