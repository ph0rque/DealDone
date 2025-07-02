package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QueueManager handles document processing queue with persistence and state tracking
type QueueManager struct {
	queue             []*QueueItem
	dealFolders       map[string]*DealFolderMirror
	processingHistory map[string]*ProcessingHistory
	config            QueueConfiguration
	mutex             sync.RWMutex
	folderMutex       sync.RWMutex
	historyMutex      sync.RWMutex
	processingCount   int
	persistenceFile   string
	ctx               context.Context
	cancel            context.CancelFunc
	isRunning         bool
}

// NewQueueManager creates a new queue manager with default configuration
func NewQueueManager(dataDir string) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	config := QueueConfiguration{
		MaxConcurrentJobs:      3,
		MaxRetryAttempts:       3,
		RetryBackoffMultiplier: 2.0,
		MaxRetryBackoff:        30 * time.Minute,
		QueueTimeout:           5 * time.Minute,
		ProcessingTimeout:      30 * time.Minute,
		HealthCheckInterval:    1 * time.Minute,
		PersistenceInterval:    5 * time.Minute,
		CleanupInterval:        1 * time.Hour,
		MaxHistoryDays:         30,
	}

	qm := &QueueManager{
		queue:             make([]*QueueItem, 0),
		dealFolders:       make(map[string]*DealFolderMirror),
		processingHistory: make(map[string]*ProcessingHistory),
		config:            config,
		persistenceFile:   filepath.Join(dataDir, "queue_state.json"),
		ctx:               ctx,
		cancel:            cancel,
	}

	// Load persisted state
	qm.loadPersistedState()

	return qm
}

// Task 3.1: FIFO processing with job metadata tracking
func (qm *QueueManager) EnqueueDocument(dealName, documentPath, documentName string, priority ProcessingPriority, metadata map[string]interface{}) (*QueueItem, error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Check for duplicate processing (race condition prevention - Task 3.4)
	for _, item := range qm.queue {
		if item.DealName == dealName && item.DocumentPath == documentPath &&
			(item.Status == QueueStatusPending || item.Status == QueueStatusProcessing) {
			return nil, fmt.Errorf("document already queued for processing: %s", documentPath)
		}
	}

	item := &QueueItem{
		ID:                uuid.New().String(),
		JobID:             uuid.New().String(),
		DealName:          dealName,
		DocumentPath:      documentPath,
		DocumentName:      documentName,
		Priority:          priority,
		QueuedAt:          time.Now(),
		Status:            QueueStatusPending,
		Metadata:          metadata,
		RetryCount:        0,
		EstimatedDuration: qm.estimateProcessingDuration(documentName),
	}

	// Insert in priority order while maintaining FIFO within same priority
	insertIndex := len(qm.queue)
	for i, existing := range qm.queue {
		if qm.shouldInsertBefore(item, existing) {
			insertIndex = i
			break
		}
	}

	// Insert at calculated position
	qm.queue = append(qm.queue[:insertIndex], append([]*QueueItem{item}, qm.queue[insertIndex:]...)...)

	// Update deal folder mirror (Task 3.2)
	qm.updateDealFolderMirror(dealName, documentPath)

	// Persist state (Task 3.3)
	go qm.persistState()

	return item, nil
}

// Task 3.5: Queue status queries and progress tracking
func (qm *QueueManager) GetQueueStatus() QueueStats {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	stats := QueueStats{
		StatusBreakdown:   make(map[QueueItemStatus]int),
		PriorityBreakdown: make(map[ProcessingPriority]int),
		LastUpdated:       time.Now(),
	}

	var totalWaitTime, totalProcessTime time.Duration
	waitTimeCount, processTimeCount := 0, 0

	for _, item := range qm.queue {
		stats.TotalItems++
		stats.StatusBreakdown[item.Status]++
		stats.PriorityBreakdown[item.Priority]++

		switch item.Status {
		case QueueStatusPending:
			stats.PendingItems++
			if !item.QueuedAt.IsZero() {
				waitTime := time.Since(item.QueuedAt)
				totalWaitTime += waitTime
				waitTimeCount++
			}
		case QueueStatusProcessing:
			stats.ProcessingItems++
		case QueueStatusCompleted:
			stats.CompletedItems++
			if item.ActualDuration > 0 {
				totalProcessTime += item.ActualDuration
				processTimeCount++
			}
		case QueueStatusFailed:
			stats.FailedItems++
		}
	}

	if waitTimeCount > 0 {
		stats.AverageWaitTime = totalWaitTime / time.Duration(waitTimeCount)
	}
	if processTimeCount > 0 {
		stats.AverageProcessTime = totalProcessTime / time.Duration(processTimeCount)
		stats.ThroughputPerHour = float64(processTimeCount) / stats.AverageProcessTime.Hours()
	}

	return stats
}

// Helper methods
func (qm *QueueManager) shouldInsertBefore(newItem, existing *QueueItem) bool {
	// Higher priority comes first
	if newItem.Priority != existing.Priority {
		return qm.getPriorityValue(newItem.Priority) > qm.getPriorityValue(existing.Priority)
	}
	// Same priority: FIFO (earlier queued time comes first)
	return newItem.QueuedAt.Before(existing.QueuedAt)
}

func (qm *QueueManager) getPriorityValue(priority ProcessingPriority) int {
	switch priority {
	case PriorityHigh:
		return 3
	case PriorityNormal:
		return 2
	case PriorityLow:
		return 1
	default:
		return 1
	}
}

func (qm *QueueManager) estimateProcessingDuration(documentName string) time.Duration {
	ext := filepath.Ext(documentName)
	switch ext {
	case ".pdf":
		return 5 * time.Minute
	case ".docx", ".doc":
		return 3 * time.Minute
	case ".xlsx", ".xls":
		return 4 * time.Minute
	default:
		return 2 * time.Minute
	}
}

func (qm *QueueManager) Start() error {
	if qm.isRunning {
		return fmt.Errorf("queue manager already running")
	}
	qm.isRunning = true
	return nil
}

func (qm *QueueManager) Stop() error {
	if !qm.isRunning {
		return fmt.Errorf("queue manager not running")
	}
	qm.cancel()
	qm.isRunning = false
	return qm.persistState()
}

func (qm *QueueManager) persistState() error {
	return nil // Placeholder implementation
}

func (qm *QueueManager) loadPersistedState() error {
	return nil // Placeholder implementation
}

// QueryQueue searches queue items based on criteria
func (qm *QueueManager) QueryQueue(query QueueQuery) ([]*QueueItem, error) {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	var results []*QueueItem

	for _, item := range qm.queue {
		if qm.matchesQuery(item, query) {
			results = append(results, item)
		}
	}

	// Sort results
	if query.SortBy != "" {
		qm.sortQueueItems(results, query.SortBy, query.SortOrder)
	}

	// Apply pagination
	start := query.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		return []*QueueItem{}, nil
	}

	end := len(results)
	if query.Limit > 0 && start+query.Limit < end {
		end = start + query.Limit
	}

	return results[start:end], nil
}

// SyncDealFolder synchronizes a deal folder structure
func (qm *QueueManager) SyncDealFolder(dealName string) error {
	return fmt.Errorf("sync deal folder not implemented yet")
}

// SynchronizeWorkflowState updates queue item status based on workflow progress
func (qm *QueueManager) SynchronizeWorkflowState(jobId, workflowStatus string) error {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, item := range qm.queue {
		if item.JobID == jobId {
			switch workflowStatus {
			case "processing":
				item.Status = QueueStatusProcessing
				if item.ProcessingStarted == nil {
					now := time.Now()
					item.ProcessingStarted = &now
				}
			case "completed":
				item.Status = QueueStatusCompleted
				if item.ProcessingEnded == nil {
					now := time.Now()
					item.ProcessingEnded = &now
					if item.ProcessingStarted != nil {
						item.ActualDuration = now.Sub(*item.ProcessingStarted)
					}
				}
			case "failed":
				item.Status = QueueStatusFailed
				item.RetryCount++
			case "retry":
				item.Status = QueueStatusRetrying
			}

			go qm.persistState()
			return nil
		}
	}

	return fmt.Errorf("job not found: %s", jobId)
}

// GetProcessingHistory returns processing history for a deal
func (qm *QueueManager) GetProcessingHistory(dealName string, limit int) []*ProcessingHistory {
	qm.historyMutex.RLock()
	defer qm.historyMutex.RUnlock()

	var history []*ProcessingHistory
	for _, h := range qm.processingHistory {
		if h.DealName == dealName {
			history = append(history, h)
		}
	}

	// Sort by start time (most recent first)
	for i := 0; i < len(history)-1; i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].StartTime.Before(history[j].StartTime) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history
}

// RecordProcessingHistory adds a processing history entry
func (qm *QueueManager) RecordProcessingHistory(dealName, documentPath, processingType string, results map[string]interface{}) {
	qm.historyMutex.Lock()
	defer qm.historyMutex.Unlock()

	historyItem := &ProcessingHistory{
		ID:              uuid.New().String(),
		DealName:        dealName,
		DocumentPath:    documentPath,
		ProcessingType:  processingType,
		StartTime:       time.Now(),
		Status:          "completed",
		Results:         results,
		Version:         1,
		UserCorrections: []UserCorrection{},
	}

	// Extract specific fields from results
	if templatesUsed, ok := results["templatesUsed"].([]string); ok {
		historyItem.TemplatesUsed = templatesUsed
	} else if templatesUsedInterface, ok := results["templatesUsed"].([]interface{}); ok {
		for _, template := range templatesUsedInterface {
			if templateStr, ok := template.(string); ok {
				historyItem.TemplatesUsed = append(historyItem.TemplatesUsed, templateStr)
			}
		}
	}

	if fieldsExtracted, ok := results["fieldsExtracted"].(float64); ok {
		historyItem.FieldsExtracted = int(fieldsExtracted)
	}

	if confidenceScore, ok := results["confidenceScore"].(float64); ok {
		historyItem.ConfidenceScore = confidenceScore
	}

	// Set end time
	endTime := time.Now()
	historyItem.EndTime = &endTime

	qm.processingHistory[historyItem.ID] = historyItem

	// Persist state
	go qm.persistState()
}

func (qm *QueueManager) matchesQuery(item *QueueItem, query QueueQuery) bool {
	if query.DealName != "" && item.DealName != query.DealName {
		return false
	}
	if query.Status != "" && item.Status != query.Status {
		return false
	}
	if query.Priority != 0 && item.Priority != query.Priority {
		return false
	}
	if query.FromTime != nil && item.QueuedAt.Before(*query.FromTime) {
		return false
	}
	if query.ToTime != nil && item.QueuedAt.After(*query.ToTime) {
		return false
	}
	return true
}

func (qm *QueueManager) sortQueueItems(items []*QueueItem, sortBy, sortOrder string) {
	// Placeholder implementation
}

func (qm *QueueManager) performHealthCheck() {
	// Placeholder implementation
}

func (qm *QueueManager) cleanupCompletedItems() {
	// Placeholder implementation
}

func (qm *QueueManager) updateDealFolderMirror(dealName, documentPath string) {
	qm.folderMutex.Lock()
	defer qm.folderMutex.Unlock()

	// Get or create deal folder mirror
	mirror, exists := qm.dealFolders[dealName]
	if !exists {
		mirror = &DealFolderMirror{
			DealName:      dealName,
			FolderPath:    filepath.Dir(documentPath),
			FileStructure: make(map[string]FileStructInfo),
			SyncStatus:    SyncStatusOutOfSync,
			LastSynced:    time.Now(),
		}
		qm.dealFolders[dealName] = mirror
	}

	// Add file to structure
	fileInfo := FileStructInfo{
		Path:            documentPath,
		ModifiedAt:      time.Now(),
		Size:            0,  // Would need to stat file for real size
		Checksum:        "", // Would compute actual checksum in production
		ProcessingState: "queued",
		QueueItemID:     "", // Would be set when processing starts
	}

	mirror.FileStructure[documentPath] = fileInfo
	mirror.FileCount = len(mirror.FileStructure)
	mirror.LastSynced = time.Now()

	// Update overall sync status based on any processing states
	mirror.SyncStatus = SyncStatusSynced
	for _, info := range mirror.FileStructure {
		if info.ProcessingState == "queued" || info.ProcessingState == "processing" {
			mirror.SyncStatus = SyncStatusOutOfSync
			break
		}
	}
}
