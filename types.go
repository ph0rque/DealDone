package main

import (
	"DealDone/internal/infrastructure/ocr"
	"time"
)

// FileSystemItem represents a file or directory in the file system
type FileSystemItem struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	IsDirectory bool             `json:"isDirectory"`
	Size        int64            `json:"size"`
	ModifiedAt  time.Time        `json:"modifiedAt"`
	CreatedAt   time.Time        `json:"createdAt"`
	Permissions FilePermissions  `json:"permissions"`
	Extension   string           `json:"extension,omitempty"`
	MimeType    string           `json:"mimeType,omitempty"`
	Children    []FileSystemItem `json:"children,omitempty"`
	IsExpanded  bool             `json:"isExpanded,omitempty"`
	IsLoading   bool             `json:"isLoading,omitempty"`
}

// FilePermissions represents file access permissions
type FilePermissions struct {
	Readable   bool `json:"readable"`
	Writable   bool `json:"writable"`
	Executable bool `json:"executable"`
}

// FileOperationResult represents the result of a file operation
type FileOperationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Items      []FileSystemItem `json:"items"`
	Query      string           `json:"query"`
	TotalCount int              `json:"totalCount"`
}

// FileOperation represents different types of file operations
type FileOperation string

const (
	OperationCreate FileOperation = "create"
	OperationCopy   FileOperation = "copy"
	OperationMove   FileOperation = "move"
	OperationDelete FileOperation = "delete"
	OperationRename FileOperation = "rename"
	OperationOpen   FileOperation = "open"
)

// CreateFileRequest represents a request to create a file
type CreateFileRequest struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	IsDirectory bool   `json:"isDirectory"`
}

// CopyMoveRequest represents a request to copy or move files
type CopyMoveRequest struct {
	SourcePaths []string `json:"sourcePaths"`
	TargetPath  string   `json:"targetPath"`
	Operation   string   `json:"operation"` // "copy" or "move"
}

// RenameRequest represents a request to rename a file/folder
type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"newName"`
}

// DeleteRequest represents a request to delete files/folders
type DeleteRequest struct {
	Paths []string `json:"paths"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}

// OpenFileRequest represents a request to open a file
type OpenFileRequest struct {
	Path string `json:"path"`
}

// DirectoryListRequest represents a request to list directory contents
type DirectoryListRequest struct {
	Path       string `json:"path"`
	ShowHidden bool   `json:"showHidden"`
}

// Webhook-related types for n8n integration

// WebhookTriggerType represents the type of trigger that initiated the webhook
type WebhookTriggerType string

const (
	TriggerFileChange WebhookTriggerType = "file_change"
	TriggerUserButton WebhookTriggerType = "user_button"
	TriggerAnalyzeAll WebhookTriggerType = "analyze_all"
	TriggerScheduled  WebhookTriggerType = "scheduled"
	TriggerRetry      WebhookTriggerType = "retry"
	TriggerCorrection WebhookTriggerType = "user_correction"
)

// WorkflowType represents the type of n8n workflow being executed
type WorkflowType string

const (
	WorkflowDocumentAnalysis WorkflowType = "document-analysis"
	WorkflowErrorHandling    WorkflowType = "error-handling"
	WorkflowUserCorrections  WorkflowType = "user-corrections"
	WorkflowCleanup          WorkflowType = "cleanup"
	WorkflowBatchProcessing  WorkflowType = "batch-processing"
	WorkflowHealthCheck      WorkflowType = "health-check"
)

// ProcessingPriority represents the priority level for processing
type ProcessingPriority int

const (
	PriorityHigh   ProcessingPriority = 1
	PriorityNormal ProcessingPriority = 2
	PriorityLow    ProcessingPriority = 3
)

// DocumentWebhookPayload represents the payload sent to n8n for document analysis
type DocumentWebhookPayload struct {
	DealName       string                 `json:"dealName" validate:"required"`
	FilePaths      []string               `json:"filePaths" validate:"required,min=1"`
	TriggerType    WebhookTriggerType     `json:"triggerType" validate:"required"`
	WorkflowType   WorkflowType           `json:"workflowType" validate:"required"`
	JobID          string                 `json:"jobId" validate:"required"`
	Priority       ProcessingPriority     `json:"priority" validate:"min=1,max=3"`
	Timestamp      int64                  `json:"timestamp" validate:"required"`
	RetryCount     int                    `json:"retryCount" validate:"min=0"`
	MaxRetries     int                    `json:"maxRetries" validate:"min=0,max=10"`
	TimeoutSeconds int                    `json:"timeoutSeconds" validate:"min=1,max=3600"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`

	// Processing Configuration
	ProcessingConfig *ProcessingConfig `json:"processingConfig,omitempty"`

	// Callback Configuration
	CallbackConfig *CallbackConfig `json:"callbackConfig,omitempty"`
}

// ProcessingConfig contains configuration for document processing behavior
type ProcessingConfig struct {
	EnableOCR               bool     `json:"enableOCR"`
	EnableTemplateDiscovery bool     `json:"enableTemplateDiscovery"`
	EnableFieldExtraction   bool     `json:"enableFieldExtraction"`
	EnableConfidenceScoring bool     `json:"enableConfidenceScoring"`
	RequiredFields          []string `json:"requiredFields,omitempty"`
	ExcludedFileTypes       []string `json:"excludedFileTypes,omitempty"`
	MaxFileSize             int64    `json:"maxFileSize,omitempty"`
	PreferredLanguage       string   `json:"preferredLanguage,omitempty"`
	AnalysisDepth           string   `json:"analysisDepth"` // "basic", "standard", "comprehensive"
}

// CallbackConfig contains configuration for webhook callbacks
type CallbackConfig struct {
	URL                string            `json:"url" validate:"required,url"`
	Method             string            `json:"method" validate:"required,oneof=POST PUT PATCH"`
	Headers            map[string]string `json:"headers,omitempty"`
	AuthType           string            `json:"authType"` // "none", "bearer", "api_key", "hmac"
	Timeout            int               `json:"timeout" validate:"min=1,max=300"`
	RetryOnFailure     bool              `json:"retryOnFailure"`
	IncludeFullResults bool              `json:"includeFullResults"`
}

// WebhookResultPayload represents the result payload received from n8n
type WebhookResultPayload struct {
	JobID              string                 `json:"jobId" validate:"required"`
	DealName           string                 `json:"dealName" validate:"required"`
	WorkflowType       WorkflowType           `json:"workflowType" validate:"required"`
	Status             string                 `json:"status" validate:"required,oneof=completed failed partial_success in_progress"`
	ProcessedDocuments int                    `json:"processedDocuments" validate:"min=0"`
	TotalDocuments     int                    `json:"totalDocuments" validate:"min=0"`
	TemplatesUpdated   []string               `json:"templatesUpdated"`
	AverageConfidence  float64                `json:"averageConfidence" validate:"min=0,max=1"`
	ProcessingTime     int64                  `json:"processingTimeMs" validate:"min=0"`
	StartTime          int64                  `json:"startTime" validate:"required"`
	EndTime            int64                  `json:"endTime,omitempty"`
	Results            *ProcessingResults     `json:"results,omitempty"`
	Errors             []ProcessingError      `json:"errors,omitempty"`
	Warnings           []string               `json:"warnings,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Timestamp          int64                  `json:"timestamp" validate:"required"`

	// Quality Metrics
	QualityMetrics *QualityMetrics `json:"qualityMetrics,omitempty"`

	// Next Steps
	NextSteps []string `json:"nextSteps,omitempty"`
}

// ProcessingResults contains detailed results from document processing
type ProcessingResults struct {
	DocumentResults     []DocumentResult   `json:"documentResults"`
	TemplateResults     []TemplateResult   `json:"templateResults"`
	ExtractionResults   []ExtractionResult `json:"extractionResults"`
	ConflictResults     []ConflictResult   `json:"conflictResults"`
	Summary             *ProcessingSummary `json:"summary"`
	Recommendations     []string           `json:"recommendations,omitempty"`
	ConfidenceBreakdown map[string]float64 `json:"confidenceBreakdown,omitempty"`
}

// DocumentResult represents the processing result for a single document
type DocumentResult struct {
	FilePath         string                 `json:"filePath" validate:"required"`
	FileName         string                 `json:"fileName" validate:"required"`
	Status           string                 `json:"status" validate:"required,oneof=processed failed skipped"`
	Classification   string                 `json:"classification"`
	Confidence       float64                `json:"confidence" validate:"min=0,max=1"`
	ProcessingTime   int64                  `json:"processingTimeMs"`
	ExtractedFields  map[string]interface{} `json:"extractedFields,omitempty"`
	TemplatesMatched []string               `json:"templatesMatched,omitempty"`
	OCRResults       *ocr.OCRResult         `json:"ocrResults,omitempty"`
	Errors           []string               `json:"errors,omitempty"`
	Warnings         []string               `json:"warnings,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateResult represents the result of template processing
type TemplateResult struct {
	TemplatePath       string                 `json:"templatePath" validate:"required"`
	TemplateName       string                 `json:"templateName" validate:"required"`
	Status             string                 `json:"status" validate:"required,oneof=updated unchanged failed"`
	FieldsUpdated      []string               `json:"fieldsUpdated"`
	FieldsConflicted   []string               `json:"fieldsConflicted"`
	AverageConfidence  float64                `json:"averageConfidence" validate:"min=0,max=1"`
	BackupCreated      bool                   `json:"backupCreated"`
	BackupPath         string                 `json:"backupPath,omitempty"`
	UpdatedFields      map[string]interface{} `json:"updatedFields,omitempty"`
	PreviousValues     map[string]interface{} `json:"previousValues,omitempty"`
	ConflictResolution map[string]string      `json:"conflictResolution,omitempty"`
}

// ExtractionResult represents field extraction results
type ExtractionResult struct {
	FieldName        string      `json:"fieldName" validate:"required"`
	ExtractedValue   interface{} `json:"extractedValue"`
	Confidence       float64     `json:"confidence" validate:"min=0,max=1"`
	Source           string      `json:"source"` // Which document/method extracted this
	Method           string      `json:"method"` // OCR, NLP, pattern matching, etc.
	Position         *Position   `json:"position,omitempty"`
	ValidationStatus string      `json:"validationStatus"` // "valid", "invalid", "warning"
	ValidationErrors []string    `json:"validationErrors,omitempty"`
}

// ConflictResult represents data conflict resolution results
type ConflictResult struct {
	FieldName         string                 `json:"fieldName" validate:"required"`
	ConflictType      string                 `json:"conflictType"` // "value", "confidence", "format"
	ConflictingValues []ConflictingValue     `json:"conflictingValues"`
	ResolvedValue     interface{}            `json:"resolvedValue"`
	ResolutionMethod  string                 `json:"resolutionMethod"` // "highest_confidence", "averaging", "manual"
	FinalConfidence   float64                `json:"finalConfidence" validate:"min=0,max=1"`
	RequiresReview    bool                   `json:"requiresReview"`
	ResolutionNotes   string                 `json:"resolutionNotes,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConflictingValue represents a single conflicting value in a conflict
type ConflictingValue struct {
	Value      interface{} `json:"value"`
	Confidence float64     `json:"confidence" validate:"min=0,max=1"`
	Source     string      `json:"source"`
	Method     string      `json:"method"`
	Timestamp  int64       `json:"timestamp"`
}

// Position represents the position of extracted data in a document
type Position struct {
	Page   int     `json:"page,omitempty"`
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// Note: OCRResult is defined in ocrservice.go

// ProcessingSummary provides overall processing statistics
type ProcessingSummary struct {
	TotalDocuments      int     `json:"totalDocuments"`
	ProcessedDocuments  int     `json:"processedDocuments"`
	FailedDocuments     int     `json:"failedDocuments"`
	SkippedDocuments    int     `json:"skippedDocuments"`
	TotalTemplates      int     `json:"totalTemplates"`
	UpdatedTemplates    int     `json:"updatedTemplates"`
	ConflictsResolved   int     `json:"conflictsResolved"`
	ConflictsPending    int     `json:"conflictsPending"`
	AverageConfidence   float64 `json:"averageConfidence" validate:"min=0,max=1"`
	TotalProcessingTime int64   `json:"totalProcessingTimeMs"`
}

// QualityMetrics provides quality assessment of processing results
type QualityMetrics struct {
	OverallScore          float64            `json:"overallScore" validate:"min=0,max=1"`
	DataCompletenessScore float64            `json:"dataCompletenessScore" validate:"min=0,max=1"`
	AccuracyScore         float64            `json:"accuracyScore" validate:"min=0,max=1"`
	ConsistencyScore      float64            `json:"consistencyScore" validate:"min=0,max=1"`
	FieldScores           map[string]float64 `json:"fieldScores,omitempty"`
	QualityIssues         []string           `json:"qualityIssues,omitempty"`
	Recommendations       []string           `json:"recommendations,omitempty"`
}

// ProcessingError represents a processing error with detailed context
type ProcessingError struct {
	Code        string                 `json:"code" validate:"required"`
	Message     string                 `json:"message" validate:"required"`
	Level       string                 `json:"level" validate:"required,oneof=error warning info"`
	Source      string                 `json:"source"` // Which component/step caused the error
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	Recoverable bool                   `json:"recoverable"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// WebhookStatusQuery represents a status query for job tracking
type WebhookStatusQuery struct {
	JobID     string `json:"jobId" validate:"required"`
	DealName  string `json:"dealName,omitempty"`
	Timestamp int64  `json:"timestamp" validate:"required"`
}

// WebhookStatusResponse represents the response to a status query
type WebhookStatusResponse struct {
	JobID          string                 `json:"jobId"`
	DealName       string                 `json:"dealName"`
	WorkflowType   WorkflowType           `json:"workflowType"`
	Status         string                 `json:"status"`
	Progress       float64                `json:"progress" validate:"min=0,max=1"`
	CurrentStep    string                 `json:"currentStep"`
	EstimatedTime  int64                  `json:"estimatedTimeMs"`
	ElapsedTime    int64                  `json:"elapsedTimeMs"`
	StartTime      int64                  `json:"startTime"`
	LastUpdated    int64                  `json:"lastUpdated"`
	ProcessingRate float64                `json:"processingRate"` // documents per minute
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// Error Handling Payloads

// ErrorHandlingPayload represents payload for error handling workflows
type ErrorHandlingPayload struct {
	OriginalJobID  string                 `json:"originalJobId" validate:"required"`
	ErrorJobID     string                 `json:"errorJobId" validate:"required"`
	DealName       string                 `json:"dealName" validate:"required"`
	ErrorType      string                 `json:"errorType" validate:"required"`
	ErrorDetails   ProcessingError        `json:"errorDetails" validate:"required"`
	RetryAttempt   int                    `json:"retryAttempt" validate:"min=1"`
	MaxRetries     int                    `json:"maxRetries" validate:"min=1"`
	RetryStrategy  string                 `json:"retryStrategy"` // "immediate", "exponential", "scheduled"
	FailedStep     string                 `json:"failedStep"`
	RecoveryAction string                 `json:"recoveryAction"` // "retry", "skip", "manual"
	Context        map[string]interface{} `json:"context,omitempty"`
	Timestamp      int64                  `json:"timestamp" validate:"required"`
}

// ErrorRecoveryResult represents the result of error recovery processing
type ErrorRecoveryResult struct {
	ErrorJobID       string                 `json:"errorJobId" validate:"required"`
	OriginalJobID    string                 `json:"originalJobId" validate:"required"`
	RecoveryStatus   string                 `json:"recoveryStatus" validate:"required,oneof=recovered failed manual_intervention_required"`
	RecoveryMethod   string                 `json:"recoveryMethod"`
	ActionsPerformed []string               `json:"actionsPerformed"`
	NewJobID         string                 `json:"newJobId,omitempty"`
	RemainingRetries int                    `json:"remainingRetries"`
	NextRetryTime    int64                  `json:"nextRetryTime,omitempty"`
	Resolution       string                 `json:"resolution,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        int64                  `json:"timestamp" validate:"required"`
}

// User Correction Payloads

// UserCorrectionPayload represents payload for user correction workflows
type UserCorrectionPayload struct {
	CorrectionID   string                 `json:"correctionId" validate:"required"`
	OriginalJobID  string                 `json:"originalJobId" validate:"required"`
	DealName       string                 `json:"dealName" validate:"required"`
	TemplatePath   string                 `json:"templatePath" validate:"required"`
	Corrections    []FieldCorrection      `json:"corrections" validate:"required,min=1"`
	UserID         string                 `json:"userId"`
	CorrectionType string                 `json:"correctionType"` // "manual", "assisted", "bulk"
	ApplyToSimilar bool                   `json:"applyToSimilar"` // Apply learning to similar documents
	Confidence     float64                `json:"confidence" validate:"min=0,max=1"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Timestamp      int64                  `json:"timestamp" validate:"required"`
}

// FieldCorrection represents a single field correction by a user
type FieldCorrection struct {
	FieldName          string      `json:"fieldName" validate:"required"`
	OriginalValue      interface{} `json:"originalValue"`
	CorrectedValue     interface{} `json:"correctedValue" validate:"required"`
	OriginalConfidence float64     `json:"originalConfidence" validate:"min=0,max=1"`
	UserConfidence     float64     `json:"userConfidence" validate:"min=0,max=1"`
	CorrectionReason   string      `json:"correctionReason"` // "wrong_extraction", "formatting", "interpretation"
	SourceDocument     string      `json:"sourceDocument,omitempty"`
	Notes              string      `json:"notes,omitempty"`
}

// CorrectionProcessingResult represents the result of processing user corrections
type CorrectionProcessingResult struct {
	CorrectionID       string                 `json:"correctionId" validate:"required"`
	ProcessingStatus   string                 `json:"processingStatus" validate:"required,oneof=processed failed partial"`
	CorrectionsApplied int                    `json:"correctionsApplied"`
	TotalCorrections   int                    `json:"totalCorrections"`
	LearningUpdated    bool                   `json:"learningUpdated"`
	ModelRetrained     bool                   `json:"modelRetrained"`
	SimilarDocuments   []string               `json:"similarDocuments,omitempty"`
	ImpactAssessment   *CorrectionImpact      `json:"impactAssessment,omitempty"`
	ValidationResults  []ValidationResult     `json:"validationResults,omitempty"`
	Recommendations    []string               `json:"recommendations,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Timestamp          int64                  `json:"timestamp" validate:"required"`
}

// CorrectionImpact assesses the impact of user corrections
type CorrectionImpact struct {
	AffectedTemplates     int     `json:"affectedTemplates"`
	AffectedDocuments     int     `json:"affectedDocuments"`
	ConfidenceImprovement float64 `json:"confidenceImprovement"`
	AccuracyImprovement   float64 `json:"accuracyImprovement"`
	PotentialReprocessing int     `json:"potentialReprocessing"`
	EstimatedBenefit      string  `json:"estimatedBenefit"` // "high", "medium", "low"
}

// ValidationResult represents the result of validating a correction
type ValidationResult struct {
	FieldName        string   `json:"fieldName" validate:"required"`
	ValidationStatus string   `json:"validationStatus" validate:"required,oneof=valid invalid warning"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
	Suggestions      []string `json:"suggestions,omitempty"`
}

// Batch Processing Payloads

// BatchProcessingPayload represents payload for batch processing workflows
type BatchProcessingPayload struct {
	BatchID        string                 `json:"batchId" validate:"required"`
	DealName       string                 `json:"dealName" validate:"required"`
	BatchType      string                 `json:"batchType" validate:"required,oneof=deal_analysis template_update bulk_correction cleanup"`
	Items          []BatchItem            `json:"items" validate:"required,min=1"`
	BatchConfig    *BatchConfig           `json:"batchConfig,omitempty"`
	Priority       ProcessingPriority     `json:"priority"`
	EstimatedTime  int64                  `json:"estimatedTimeMs"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	CallbackConfig *CallbackConfig        `json:"callbackConfig,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      int64                  `json:"timestamp" validate:"required"`
}

// BatchItem represents a single item in a batch processing request
type BatchItem struct {
	ItemID       string                 `json:"itemId" validate:"required"`
	ItemType     string                 `json:"itemType" validate:"required"`
	ItemPath     string                 `json:"itemPath"`
	Priority     ProcessingPriority     `json:"priority"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BatchConfig contains configuration for batch processing
type BatchConfig struct {
	MaxConcurrent        int    `json:"maxConcurrent" validate:"min=1,max=20"`
	FailureHandling      string `json:"failureHandling"` // "stop", "continue", "retry"
	MaxRetries           int    `json:"maxRetries" validate:"min=0,max=5"`
	TimeoutPerItem       int    `json:"timeoutPerItem" validate:"min=1"`
	ProgressNotification bool   `json:"progressNotification"`
	IntermediateResults  bool   `json:"intermediateResults"`
}

// BatchProcessingResult represents the result of batch processing
type BatchProcessingResult struct {
	BatchID        string                 `json:"batchId" validate:"required"`
	BatchStatus    string                 `json:"batchStatus" validate:"required,oneof=completed failed partial cancelled"`
	TotalItems     int                    `json:"totalItems"`
	ProcessedItems int                    `json:"processedItems"`
	FailedItems    int                    `json:"failedItems"`
	SkippedItems   int                    `json:"skippedItems"`
	ItemResults    []BatchItemResult      `json:"itemResults"`
	BatchSummary   *BatchSummary          `json:"batchSummary,omitempty"`
	ProcessingTime int64                  `json:"processingTimeMs"`
	StartTime      int64                  `json:"startTime"`
	EndTime        int64                  `json:"endTime"`
	Errors         []ProcessingError      `json:"errors,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      int64                  `json:"timestamp" validate:"required"`
}

// BatchItemResult represents the result of processing a single batch item
type BatchItemResult struct {
	ItemID         string                 `json:"itemId" validate:"required"`
	Status         string                 `json:"status" validate:"required,oneof=completed failed skipped"`
	ProcessingTime int64                  `json:"processingTimeMs"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Errors         []string               `json:"errors,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// BatchSummary provides summary statistics for batch processing
type BatchSummary struct {
	SuccessRate           float64            `json:"successRate" validate:"min=0,max=1"`
	AverageProcessingTime int64              `json:"averageProcessingTimeMs"`
	ThroughputRate        float64            `json:"throughputRate"` // items per minute
	ResourceUtilization   map[string]float64 `json:"resourceUtilization,omitempty"`
	QualityMetrics        *QualityMetrics    `json:"qualityMetrics,omitempty"`
	Recommendations       []string           `json:"recommendations,omitempty"`
}

// Authentication and Configuration Types

// WebhookAuthConfig represents authentication configuration for webhooks
type WebhookAuthConfig struct {
	APIKey          string `json:"apiKey"`
	SharedSecret    string `json:"sharedSecret"`
	TokenExpiration int64  `json:"tokenExpiration"`
	EnableHMAC      bool   `json:"enableHMAC"`
	AuthType        string `json:"authType"` // "api_key", "hmac", "bearer", "basic"
}

// WebhookValidationResult represents the result of payload validation
type WebhookValidationResult struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	ValidationTime int64    `json:"validationTimeMs"`
	SchemaVersion  string   `json:"schemaVersion"`
}

// Health Check Payloads

// HealthCheckPayload represents payload for health check workflows
type HealthCheckPayload struct {
	CheckID    string                 `json:"checkId" validate:"required"`
	CheckType  string                 `json:"checkType" validate:"required,oneof=system component workflow end_to_end"`
	Components []string               `json:"components,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Timestamp  int64                  `json:"timestamp" validate:"required"`
}

// HealthCheckResult represents the result of health checks
type HealthCheckResult struct {
	CheckID         string                     `json:"checkId" validate:"required"`
	OverallStatus   string                     `json:"overallStatus" validate:"required,oneof=healthy degraded unhealthy"`
	ComponentStatus map[string]ComponentHealth `json:"componentStatus"`
	ResponseTime    int64                      `json:"responseTimeMs"`
	Checks          []IndividualCheck          `json:"checks"`
	Recommendations []string                   `json:"recommendations,omitempty"`
	Timestamp       int64                      `json:"timestamp" validate:"required"`
}

// ComponentHealth represents the health status of a single component
type ComponentHealth struct {
	Status      string                 `json:"status" validate:"required,oneof=healthy degraded unhealthy"`
	Message     string                 `json:"message,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	LastChecked int64                  `json:"lastChecked"`
}

// IndividualCheck represents a single health check result
type IndividualCheck struct {
	Name     string                 `json:"name" validate:"required"`
	Status   string                 `json:"status" validate:"required,oneof=pass fail warn"`
	Message  string                 `json:"message,omitempty"`
	Duration int64                  `json:"durationMs"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// JSON Schema Validation

// JSONSchemaInfo represents information about a JSON schema
type JSONSchemaInfo struct {
	SchemaName    string `json:"schemaName"`
	SchemaVersion string `json:"schemaVersion"`
	Description   string `json:"description"`
	LastUpdated   int64  `json:"lastUpdated"`
}

// Webhook API Version and Compatibility

// APIVersion represents the version of the webhook API
type APIVersion struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
	Label string `json:"label,omitempty"` // alpha, beta, rc
}

// CompatibilityInfo represents API compatibility information
type CompatibilityInfo struct {
	CurrentVersion     APIVersion   `json:"currentVersion"`
	SupportedVersions  []APIVersion `json:"supportedVersions"`
	DeprecatedVersions []APIVersion `json:"deprecatedVersions"`
	MinimumVersion     APIVersion   `json:"minimumVersion"`
}

// Queue Management Types
type QueueItem struct {
	ID                string                 `json:"id"`
	JobID             string                 `json:"jobId"`
	DealName          string                 `json:"dealName"`
	DocumentPath      string                 `json:"documentPath"`
	DocumentName      string                 `json:"documentName"`
	Priority          ProcessingPriority     `json:"priority"`
	QueuedAt          time.Time              `json:"queuedAt"`
	ProcessingStarted *time.Time             `json:"processingStarted,omitempty"`
	ProcessingEnded   *time.Time             `json:"processingEnded,omitempty"`
	Status            QueueItemStatus        `json:"status"`
	Metadata          map[string]interface{} `json:"metadata"`
	Dependencies      []string               `json:"dependencies,omitempty"`
	RetryCount        int                    `json:"retryCount"`
	LastError         *QueueError            `json:"lastError,omitempty"`
	EstimatedDuration time.Duration          `json:"estimatedDuration"`
	ActualDuration    time.Duration          `json:"actualDuration"`
}

type QueueItemStatus string

const (
	QueueStatusPending    QueueItemStatus = "pending"
	QueueStatusProcessing QueueItemStatus = "processing"
	QueueStatusCompleted  QueueItemStatus = "completed"
	QueueStatusFailed     QueueItemStatus = "failed"
	QueueStatusCanceled   QueueItemStatus = "canceled"
	QueueStatusRetrying   QueueItemStatus = "retrying"
	QueueStatusBlocked    QueueItemStatus = "blocked"
)

type QueueError struct {
	ErrorType   string    `json:"errorType"`
	Message     string    `json:"message"`
	OccurredAt  time.Time `json:"occurredAt"`
	IsRetryable bool      `json:"isRetryable"`
	Details     string    `json:"details,omitempty"`
}

type QueueStats struct {
	TotalItems         int                        `json:"totalItems"`
	PendingItems       int                        `json:"pendingItems"`
	ProcessingItems    int                        `json:"processingItems"`
	CompletedItems     int                        `json:"completedItems"`
	FailedItems        int                        `json:"failedItems"`
	StatusBreakdown    map[QueueItemStatus]int    `json:"statusBreakdown"`
	PriorityBreakdown  map[ProcessingPriority]int `json:"priorityBreakdown"`
	AverageWaitTime    time.Duration              `json:"averageWaitTime"`
	AverageProcessTime time.Duration              `json:"averageProcessTime"`
	ThroughputPerHour  float64                    `json:"throughputPerHour"`
	LastUpdated        time.Time                  `json:"lastUpdated"`
}

type QueueQuery struct {
	DealName  string             `json:"dealName,omitempty"`
	Status    QueueItemStatus    `json:"status,omitempty"`
	Priority  ProcessingPriority `json:"priority,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
	SortBy    string             `json:"sortBy,omitempty"`
	SortOrder string             `json:"sortOrder,omitempty"`
	FromTime  *time.Time         `json:"fromTime,omitempty"`
	ToTime    *time.Time         `json:"toTime,omitempty"`
}

type DealFolderMirror struct {
	DealName       string                    `json:"dealName"`
	FolderPath     string                    `json:"folderPath"`
	LastSynced     time.Time                 `json:"lastSynced"`
	SyncStatus     FolderSyncStatus          `json:"syncStatus"`
	FileCount      int                       `json:"fileCount"`
	ProcessedFiles int                       `json:"processedFiles"`
	FileStructure  map[string]FileStructInfo `json:"fileStructure"`
	ConflictFiles  []string                  `json:"conflictFiles,omitempty"`
	SyncErrors     []SyncError               `json:"syncErrors,omitempty"`
}

type FolderSyncStatus string

const (
	SyncStatusSynced    FolderSyncStatus = "synced"
	SyncStatusSyncing   FolderSyncStatus = "syncing"
	SyncStatusOutOfSync FolderSyncStatus = "out-of-sync"
	SyncStatusError     FolderSyncStatus = "error"
	SyncStatusConflict  FolderSyncStatus = "conflict"
)

type FileStructInfo struct {
	Path            string    `json:"path"`
	ModifiedAt      time.Time `json:"modifiedAt"`
	Size            int64     `json:"size"`
	Checksum        string    `json:"checksum"`
	ProcessingState string    `json:"processingState"`
	QueueItemID     string    `json:"queueItemId,omitempty"`
}

type SyncError struct {
	FilePath   string    `json:"filePath"`
	ErrorType  string    `json:"errorType"`
	Message    string    `json:"message"`
	OccurredAt time.Time `json:"occurredAt"`
	Resolved   bool      `json:"resolved"`
}

type ProcessingHistory struct {
	ID              string                 `json:"id"`
	DealName        string                 `json:"dealName"`
	DocumentPath    string                 `json:"documentPath"`
	ProcessingType  string                 `json:"processingType"`
	StartTime       time.Time              `json:"startTime"`
	EndTime         *time.Time             `json:"endTime,omitempty"`
	Status          string                 `json:"status"`
	Results         map[string]interface{} `json:"results,omitempty"`
	TemplatesUsed   []string               `json:"templatesUsed,omitempty"`
	FieldsExtracted int                    `json:"fieldsExtracted"`
	ConfidenceScore float64                `json:"confidenceScore"`
	ProcessingNotes string                 `json:"processingNotes,omitempty"`
	UserCorrections []UserCorrection       `json:"userCorrections,omitempty"`
	Version         int                    `json:"version"`
}

type UserCorrection struct {
	FieldName      string    `json:"fieldName"`
	OriginalValue  string    `json:"originalValue"`
	CorrectedValue string    `json:"correctedValue"`
	CorrectedBy    string    `json:"correctedBy"`
	CorrectedAt    time.Time `json:"correctedAt"`
	Confidence     float64   `json:"confidence"`
	Reason         string    `json:"reason,omitempty"`
}

type QueueConfiguration struct {
	MaxConcurrentJobs      int           `json:"maxConcurrentJobs"`
	MaxRetryAttempts       int           `json:"maxRetryAttempts"`
	RetryBackoffMultiplier float64       `json:"retryBackoffMultiplier"`
	MaxRetryBackoff        time.Duration `json:"maxRetryBackoff"`
	QueueTimeout           time.Duration `json:"queueTimeout"`
	ProcessingTimeout      time.Duration `json:"processingTimeout"`
	HealthCheckInterval    time.Duration `json:"healthCheckInterval"`
	PersistenceInterval    time.Duration `json:"persistenceInterval"`
	CleanupInterval        time.Duration `json:"cleanupInterval"`
	MaxHistoryDays         int           `json:"maxHistoryDays"`
}

type StateSnapshot struct {
	Timestamp         time.Time                   `json:"timestamp"`
	QueueItems        []QueueItem                 `json:"queueItems"`
	DealFolders       map[string]DealFolderMirror `json:"dealFolders"`
	ProcessingHistory []ProcessingHistory         `json:"processingHistory"`
	Configuration     QueueConfiguration          `json:"configuration"`
	Version           string                      `json:"version"`
	Checksum          string                      `json:"checksum"`
}
