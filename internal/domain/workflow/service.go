package workflow

import (
	"context"
)

// Service defines the interface for workflow management operations
type Service interface {
	// Workflow execution
	CreateWorkflowExecution(workflowType, dealID, documentID string, steps []WorkflowStep) (*WorkflowExecution, error)
	GetWorkflowExecution(executionID string) (*WorkflowExecution, error)
	GetWorkflowExecutionsByStatus(status WorkflowStatus) ([]*WorkflowExecution, error)
	ExecuteWorkflowExecution(executionID string) error
	ResumeWorkflowExecution(executionID string) error

	// Workflow recovery
	RecoverFailedWorkflow(executionID string) error
	GetWorkflowErrorStatistics() map[string]int
	CleanupOldWorkflowExecutions() error
	GetWorkflowRecoveryStatus() map[string]interface{}

	// Conflict resolution
	ResolveConflict(conflict *ConflictRecord) (*ConflictResolution, error)
	GetConflictHistory(dealID string) ([]*ConflictRecord, error)

	// Correction processing
	DetectUserCorrection(correction *CorrectionEntry) error
	MonitorTemplateDataChanges(dealID, templateID string, beforeData, afterData map[string]interface{}, userID string) error
	GetLearningInsights() (*LearningInsights, error)
	ApplyLearningToDocument(documentData map[string]interface{}, context ProcessingContext) (*ProcessingResult, error)
	GetCorrectionHistory(filters CorrectionHistoryFilters) ([]*CorrectionEntry, error)
	GetLearningPatterns() (map[string]*LearningPattern, error)
	UpdateLearningPattern(patternID string, updates PatternUpdate) error
	GetCorrectionStatistics() (*CorrectionStatistics, error)
	ForceLearningUpdate() error
}

// StepExecutor defines the interface for executing workflow steps
type StepExecutor interface {
	ExecuteStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error
	ValidateStep(step *WorkflowStep) error
	RollbackStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error
}

// ErrorNotifier defines the interface for workflow error notifications
type ErrorNotifier interface {
	NotifyError(execution *WorkflowExecution, step *WorkflowStep, err error) error
	NotifyCriticalFailure(execution *WorkflowExecution, message string) error
	NotifyRecoverySuccess(execution *WorkflowExecution, message string) error
}
