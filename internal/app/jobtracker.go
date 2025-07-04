package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// JobStatus represents the status of a processing job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCanceled   JobStatus = "canceled"
	JobStatusPartial    JobStatus = "partial"
)

// JobTracker manages job tracking, status queries, and processing history
type JobTracker struct {
	jobs       map[string]*JobInfo
	history    []string // Job IDs in chronological order
	mu         sync.RWMutex
	configPath string
	maxHistory int
}

// JobInfo represents detailed information about a processing job
type JobInfo struct {
	JobID              string                 `json:"jobId"`
	DealName           string                 `json:"dealName"`
	Status             JobStatus              `json:"status"`
	TriggerType        WebhookTriggerType     `json:"triggerType"`
	FilePaths          []string               `json:"filePaths"`
	CreatedAt          int64                  `json:"createdAt"`
	UpdatedAt          int64                  `json:"updatedAt"`
	StartedAt          int64                  `json:"startedAt,omitempty"`
	CompletedAt        int64                  `json:"completedAt,omitempty"`
	Progress           float64                `json:"progress"`
	CurrentStep        string                 `json:"currentStep"`
	EstimatedTime      int64                  `json:"estimatedTimeMs"`
	ProcessedDocuments int                    `json:"processedDocuments"`
	TotalDocuments     int                    `json:"totalDocuments"`
	ProcessingResults  *WebhookResultPayload  `json:"processingResults,omitempty"`
	Errors             []string               `json:"errors,omitempty"`
	RetryCount         int                    `json:"retryCount"`
	MaxRetries         int                    `json:"maxRetries"`
	QueuePosition      int                    `json:"queuePosition"`
	ProcessingHistory  []JobHistoryEntry      `json:"processingHistory"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// JobHistoryEntry represents a single entry in job processing history
type JobHistoryEntry struct {
	Timestamp int64   `json:"timestamp"`
	Status    string  `json:"status"`
	Step      string  `json:"step"`
	Message   string  `json:"message,omitempty"`
	Progress  float64 `json:"progress"`
	Error     string  `json:"error,omitempty"`
}

// JobQuery represents parameters for querying jobs
type JobQuery struct {
	DealName    string             `json:"dealName,omitempty"`
	Status      JobStatus          `json:"status,omitempty"`
	TriggerType WebhookTriggerType `json:"triggerType,omitempty"`
	Limit       int                `json:"limit,omitempty"`
	Offset      int                `json:"offset,omitempty"`
	SortBy      string             `json:"sortBy,omitempty"`    // "createdAt", "updatedAt", "progress"
	SortOrder   string             `json:"sortOrder,omitempty"` // "asc", "desc"
}

// JobSummary represents aggregated job statistics
type JobSummary struct {
	TotalJobs       int               `json:"totalJobs"`
	ActiveJobs      int               `json:"activeJobs"`
	CompletedJobs   int               `json:"completedJobs"`
	FailedJobs      int               `json:"failedJobs"`
	AverageProgress float64           `json:"averageProgress"`
	StatusCounts    map[JobStatus]int `json:"statusCounts"`
	DealCounts      map[string]int    `json:"dealCounts"`
	RecentActivity  []JobHistoryEntry `json:"recentActivity"`
}

// NewJobTracker creates a new job tracker instance
func NewJobTracker(configService *ConfigService) *JobTracker {
	configPath := ""
	if configService != nil {
		configPath = filepath.Join(configService.GetDealDoneRoot(), ".dealdone", "job_tracker.json")
	}

	tracker := &JobTracker{
		jobs:       make(map[string]*JobInfo),
		history:    make([]string, 0),
		configPath: configPath,
		maxHistory: 1000, // Keep last 1000 jobs in history
	}

	// Load existing data
	if configPath != "" {
		tracker.loadFromDisk()
	}

	return tracker
}

// CreateJob creates a new job with initial status
func (jt *JobTracker) CreateJob(jobID, dealName string, triggerType WebhookTriggerType, filePaths []string) *JobInfo {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	now := time.Now().UnixMilli()

	job := &JobInfo{
		JobID:              jobID,
		DealName:           dealName,
		Status:             JobStatusPending,
		TriggerType:        triggerType,
		FilePaths:          filePaths,
		CreatedAt:          now,
		UpdatedAt:          now,
		Progress:           0.0,
		CurrentStep:        "Created",
		TotalDocuments:     len(filePaths),
		ProcessedDocuments: 0,
		RetryCount:         0,
		MaxRetries:         3,
		ProcessingHistory: []JobHistoryEntry{
			{
				Timestamp: now,
				Status:    string(JobStatusPending),
				Step:      "Created",
				Message:   fmt.Sprintf("Job created for deal %s with %d documents", dealName, len(filePaths)),
				Progress:  0.0,
			},
		},
		Metadata: make(map[string]interface{}),
	}

	jt.jobs[jobID] = job
	jt.addToHistory(jobID)
	jt.saveToDisk()

	return job
}

// UpdateJob updates job status and progress
func (jt *JobTracker) UpdateJob(jobID string, updates map[string]interface{}) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	now := time.Now().UnixMilli()
	job.UpdatedAt = now

	// Track what changed for history
	var historyMessage string
	var statusChanged bool

	// Update fields based on provided updates
	if status, ok := updates["status"].(string); ok {
		oldStatus := job.Status
		job.Status = JobStatus(status)
		if job.Status != oldStatus {
			statusChanged = true
			historyMessage = fmt.Sprintf("Status changed from %s to %s", oldStatus, job.Status)

			if job.Status == JobStatusProcessing && job.StartedAt == 0 {
				job.StartedAt = now
			}
			if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
				job.CompletedAt = now
			}
		}
	}

	if progress, ok := updates["progress"].(float64); ok {
		oldProgress := job.Progress
		job.Progress = progress
		if !statusChanged && progress != oldProgress {
			historyMessage = fmt.Sprintf("Progress updated to %.1f%%", progress*100)
		}
	}

	if step, ok := updates["currentStep"].(string); ok {
		job.CurrentStep = step
		if historyMessage == "" {
			historyMessage = fmt.Sprintf("Current step: %s", step)
		}
	}

	if estimatedTime, ok := updates["estimatedTime"].(int64); ok {
		job.EstimatedTime = estimatedTime
	}

	if processedDocs, ok := updates["processedDocuments"].(int); ok {
		job.ProcessedDocuments = processedDocs
	}

	if queuePos, ok := updates["queuePosition"].(int); ok {
		job.QueuePosition = queuePos
	}

	if errors, ok := updates["errors"].([]string); ok {
		job.Errors = errors
	}

	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			job.Metadata[k] = v
		}
	}

	// Add history entry
	if historyMessage != "" {
		historyEntry := JobHistoryEntry{
			Timestamp: now,
			Status:    string(job.Status),
			Step:      job.CurrentStep,
			Message:   historyMessage,
			Progress:  job.Progress,
		}

		if len(job.Errors) > 0 {
			historyEntry.Error = job.Errors[len(job.Errors)-1] // Latest error
		}

		job.ProcessingHistory = append(job.ProcessingHistory, historyEntry)
	}

	jt.saveToDisk()
	return nil
}

// CompleteJob marks a job as completed with results
func (jt *JobTracker) CompleteJob(jobID string, results *WebhookResultPayload) error {
	updates := map[string]interface{}{
		"status":             string(JobStatusCompleted),
		"progress":           1.0,
		"currentStep":        "Completed",
		"processedDocuments": results.ProcessedDocuments,
	}

	if err := jt.UpdateJob(jobID, updates); err != nil {
		return err
	}

	// Store the complete results
	jt.mu.Lock()
	defer jt.mu.Unlock()

	if job, exists := jt.jobs[jobID]; exists {
		job.ProcessingResults = results
		jt.saveToDisk()
	}

	return nil
}

// FailJob marks a job as failed with error information
func (jt *JobTracker) FailJob(jobID string, errorMsg string) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = JobStatusFailed
	job.UpdatedAt = time.Now().UnixMilli()
	job.CompletedAt = job.UpdatedAt
	job.Errors = append(job.Errors, errorMsg)

	// Add failure to history
	job.ProcessingHistory = append(job.ProcessingHistory, JobHistoryEntry{
		Timestamp: job.UpdatedAt,
		Status:    string(JobStatusFailed),
		Step:      job.CurrentStep,
		Message:   "Job failed",
		Progress:  job.Progress,
		Error:     errorMsg,
	})

	jt.saveToDisk()
	return nil
}

// GetJob retrieves job information by ID
func (jt *JobTracker) GetJob(jobID string) (*JobInfo, error) {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy to prevent external modifications
	jobCopy := *job
	return &jobCopy, nil
}

// QueryJobs queries jobs based on criteria
func (jt *JobTracker) QueryJobs(query *JobQuery) ([]*JobInfo, error) {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	var results []*JobInfo

	// Filter jobs based on query criteria
	for _, job := range jt.jobs {
		if query.DealName != "" && job.DealName != query.DealName {
			continue
		}
		if query.Status != "" && job.Status != query.Status {
			continue
		}
		if query.TriggerType != "" && job.TriggerType != query.TriggerType {
			continue
		}

		// Create copy to prevent external modifications
		jobCopy := *job
		results = append(results, &jobCopy)
	}

	// Sort results
	if query.SortBy != "" {
		jt.sortJobs(results, query.SortBy, query.SortOrder)
	}

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		end := start + query.Limit
		if start >= len(results) {
			return []*JobInfo{}, nil
		}
		if end > len(results) {
			end = len(results)
		}
		results = results[start:end]
	}

	return results, nil
}

// GetJobSummary returns aggregated job statistics
func (jt *JobTracker) GetJobSummary() *JobSummary {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	summary := &JobSummary{
		StatusCounts:   make(map[JobStatus]int),
		DealCounts:     make(map[string]int),
		RecentActivity: make([]JobHistoryEntry, 0),
	}

	var totalProgress float64
	var activeJobs int

	// Collect statistics
	for _, job := range jt.jobs {
		summary.TotalJobs++
		summary.StatusCounts[job.Status]++
		summary.DealCounts[job.DealName]++

		if job.Status == JobStatusProcessing || job.Status == JobStatusQueued {
			activeJobs++
			totalProgress += job.Progress
		}

		if job.Status == JobStatusCompleted {
			summary.CompletedJobs++
		}
		if job.Status == JobStatusFailed {
			summary.FailedJobs++
		}

		// Collect recent activity
		for _, historyEntry := range job.ProcessingHistory {
			summary.RecentActivity = append(summary.RecentActivity, historyEntry)
		}
	}

	summary.ActiveJobs = activeJobs
	if activeJobs > 0 {
		summary.AverageProgress = totalProgress / float64(activeJobs)
	}

	// Sort recent activity by timestamp (most recent first)
	sort.Slice(summary.RecentActivity, func(i, j int) bool {
		return summary.RecentActivity[i].Timestamp > summary.RecentActivity[j].Timestamp
	})

	// Keep only the most recent 50 entries
	if len(summary.RecentActivity) > 50 {
		summary.RecentActivity = summary.RecentActivity[:50]
	}

	return summary
}

// CleanupOldJobs removes old completed/failed jobs to prevent memory growth
func (jt *JobTracker) CleanupOldJobs(olderThanHours int) int {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	cutoffTime := time.Now().Add(-time.Duration(olderThanHours) * time.Hour).UnixMilli()
	var removedCount int

	for jobID, job := range jt.jobs {
		if (job.Status == JobStatusCompleted || job.Status == JobStatusFailed) &&
			job.UpdatedAt < cutoffTime {
			delete(jt.jobs, jobID)
			removedCount++
		}
	}

	// Clean up history
	newHistory := make([]string, 0)
	for _, jobID := range jt.history {
		if _, exists := jt.jobs[jobID]; exists {
			newHistory = append(newHistory, jobID)
		}
	}
	jt.history = newHistory

	if removedCount > 0 {
		jt.saveToDisk()
	}

	return removedCount
}

// Internal helper methods

func (jt *JobTracker) addToHistory(jobID string) {
	jt.history = append(jt.history, jobID)

	// Trim history if it gets too long
	if len(jt.history) > jt.maxHistory {
		jt.history = jt.history[len(jt.history)-jt.maxHistory:]
	}
}

func (jt *JobTracker) sortJobs(jobs []*JobInfo, sortBy, sortOrder string) {
	ascending := sortOrder != "desc"

	sort.Slice(jobs, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "createdAt":
			less = jobs[i].CreatedAt < jobs[j].CreatedAt
		case "updatedAt":
			less = jobs[i].UpdatedAt < jobs[j].UpdatedAt
		case "progress":
			less = jobs[i].Progress < jobs[j].Progress
		default:
			less = jobs[i].CreatedAt < jobs[j].CreatedAt
		}

		if ascending {
			return less
		}
		return !less
	})
}

func (jt *JobTracker) saveToDisk() {
	if jt.configPath == "" {
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(jt.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	// Prepare data for saving
	data := map[string]interface{}{
		"jobs":     jt.jobs,
		"history":  jt.history,
		"saved_at": time.Now().UnixMilli(),
	}

	// Write to file
	if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		os.WriteFile(jt.configPath, jsonData, 0644)
	}
}

func (jt *JobTracker) loadFromDisk() {
	if jt.configPath == "" {
		return
	}

	data, err := os.ReadFile(jt.configPath)
	if err != nil {
		return
	}

	var savedData map[string]interface{}
	if err := json.Unmarshal(data, &savedData); err != nil {
		return
	}

	// Load jobs
	if jobsData, ok := savedData["jobs"].(map[string]interface{}); ok {
		for jobID, jobDataInterface := range jobsData {
			if jobDataBytes, err := json.Marshal(jobDataInterface); err == nil {
				var job JobInfo
				if err := json.Unmarshal(jobDataBytes, &job); err == nil {
					jt.jobs[jobID] = &job
				}
			}
		}
	}

	// Load history
	if historyData, ok := savedData["history"].([]interface{}); ok {
		jt.history = make([]string, len(historyData))
		for i, item := range historyData {
			if jobID, ok := item.(string); ok {
				jt.history[i] = jobID
			}
		}
	}
}

// RetryJob resets a failed job for retry
func (jt *JobTracker) RetryJob(jobID string) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != JobStatusFailed {
		return fmt.Errorf("job is not in failed state: %s", job.Status)
	}

	if job.RetryCount >= job.MaxRetries {
		return fmt.Errorf("job has exceeded maximum retries (%d)", job.MaxRetries)
	}

	// Reset job for retry
	job.Status = JobStatusPending
	job.Progress = 0.0
	job.CurrentStep = "Retrying"
	job.UpdatedAt = time.Now().UnixMilli()
	job.RetryCount++
	job.CompletedAt = 0

	// Add retry to history
	job.ProcessingHistory = append(job.ProcessingHistory, JobHistoryEntry{
		Timestamp: job.UpdatedAt,
		Status:    string(JobStatusPending),
		Step:      "Retrying",
		Message:   fmt.Sprintf("Retry attempt %d of %d", job.RetryCount, job.MaxRetries),
		Progress:  0.0,
	})

	jt.saveToDisk()
	return nil
}

// CancelJob cancels a pending or processing job
func (jt *JobTracker) CancelJob(jobID string) error {
	return jt.UpdateJob(jobID, map[string]interface{}{
		"status":      string(JobStatusCanceled),
		"currentStep": "Canceled",
	})
}

// IsHealthy checks if the job tracker is functioning properly
func (jt *JobTracker) IsHealthy(ctx context.Context) error {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	// Check if we can access jobs
	if jt.jobs == nil {
		return fmt.Errorf("job tracker not initialized")
	}

	// Check file system access if config path is set
	if jt.configPath != "" {
		dir := filepath.Dir(jt.configPath)
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("cannot access job tracker storage: %w", err)
		}
	}

	return nil
}
