package n8n

import (
	"time"
)

// WorkflowStatus represents the status of an n8n workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusRetrying  WorkflowStatus = "retrying"
)

// WorkflowExecution represents an n8n workflow execution
type WorkflowExecution struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflowId"`
	WorkflowName string                 `json:"workflowName"`
	Status       WorkflowStatus         `json:"status"`
	StartedAt    time.Time              `json:"startedAt"`
	FinishedAt   *time.Time             `json:"finishedAt,omitempty"`
	Data         map[string]interface{} `json:"data"`
	Error        string                 `json:"error,omitempty"`
	RetryOf      string                 `json:"retryOf,omitempty"`
	RetryCount   int                    `json:"retryCount"`
}

// WorkflowConfig represents configuration for n8n workflows
type WorkflowConfig struct {
	WorkflowID string                 `json:"workflowId"`
	Name       string                 `json:"name"`
	WebhookURL string                 `json:"webhookUrl"`
	Timeout    time.Duration          `json:"timeout"`
	MaxRetries int                    `json:"maxRetries"`
	RetryDelay time.Duration          `json:"retryDelay"`
	Active     bool                   `json:"active"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// JobStatus represents the status of a job in n8n
type JobStatus struct {
	JobID          string                 `json:"jobId"`
	WorkflowID     string                 `json:"workflowId"`
	ExecutionID    string                 `json:"executionId"`
	Status         WorkflowStatus         `json:"status"`
	Progress       float64                `json:"progress"`
	CurrentNode    string                 `json:"currentNode"`
	ProcessedNodes int                    `json:"processedNodes"`
	TotalNodes     int                    `json:"totalNodes"`
	StartTime      time.Time              `json:"startTime"`
	UpdateTime     time.Time              `json:"updateTime"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowTrigger represents a trigger for an n8n workflow
type WorkflowTrigger struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Event       string                 `json:"event"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Schedule    string                 `json:"schedule,omitempty"`
	WebhookPath string                 `json:"webhookPath,omitempty"`
}

// WorkflowNode represents a node in an n8n workflow
type WorkflowNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Position   NodePosition           `json:"position"`
	Parameters map[string]interface{} `json:"parameters"`
	Disabled   bool                   `json:"disabled"`
}

// NodePosition represents the position of a node in the workflow canvas
type NodePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WorkflowConnection represents a connection between nodes
type WorkflowConnection struct {
	Source ConnectionPoint `json:"source"`
	Target ConnectionPoint `json:"target"`
}

// ConnectionPoint represents a connection point on a node
type ConnectionPoint struct {
	NodeID string `json:"nodeId"`
	Type   string `json:"type"`
	Index  int    `json:"index"`
}

// WebhookResponse represents the response from an n8n webhook
type WebhookResponse struct {
	Success      bool                   `json:"success"`
	ExecutionID  string                 `json:"executionId,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Error        string                 `json:"error,omitempty"`
	ErrorDetails map[string]interface{} `json:"errorDetails,omitempty"`
}

// WorkflowMetrics represents metrics for workflow execution
type WorkflowMetrics struct {
	WorkflowID           string        `json:"workflowId"`
	TotalExecutions      int64         `json:"totalExecutions"`
	SuccessfulExecutions int64         `json:"successfulExecutions"`
	FailedExecutions     int64         `json:"failedExecutions"`
	AverageExecutionTime time.Duration `json:"averageExecutionTime"`
	LastExecutionTime    time.Time     `json:"lastExecutionTime"`
	ErrorRate            float64       `json:"errorRate"`
}

// Integration represents the n8n integration configuration
type Integration struct {
	BaseURL        string                    `json:"baseUrl"`
	APIKey         string                    `json:"apiKey"`
	Workflows      map[string]WorkflowConfig `json:"workflows"`
	DefaultTimeout time.Duration             `json:"defaultTimeout"`
	MaxConcurrent  int                       `json:"maxConcurrent"`
	RetryConfig    RetryConfig               `json:"retryConfig"`
}

// RetryConfig represents retry configuration for n8n operations
type RetryConfig struct {
	MaxAttempts  int           `json:"maxAttempts"`
	InitialDelay time.Duration `json:"initialDelay"`
	MaxDelay     time.Duration `json:"maxDelay"`
	Multiplier   float64       `json:"multiplier"`
}
