package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestQueueManager_Creation(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	if qm == nil {
		t.Fatal("QueueManager should not be nil")
	}

	if len(qm.queue) != 0 {
		t.Errorf("Expected empty queue, got %d items", len(qm.queue))
	}

	if len(qm.dealFolders) != 0 {
		t.Errorf("Expected empty deal folders, got %d", len(qm.dealFolders))
	}
}

func TestQueueManager_EnqueueDocument(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	item, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, map[string]interface{}{"test": "value"})
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	if item.ID == "" {
		t.Error("Item ID should not be empty")
	}

	if item.JobID == "" {
		t.Error("Job ID should not be empty")
	}

	if item.Status != QueueStatusPending {
		t.Errorf("Expected status %s, got %s", QueueStatusPending, item.Status)
	}

	// Test duplicate prevention
	_, err = qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err == nil {
		t.Error("Expected error when enqueueing duplicate document")
	}
}

func TestQueueManager_PriorityOrdering(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test files
	testFiles := []string{"normal.pdf", "high.pdf", "low.pdf"}
	for _, filename := range testFiles {
		testFile := filepath.Join(tempDir, filename)
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Enqueue with different priorities
	_, err := qm.EnqueueDocument("TestDeal", filepath.Join(tempDir, "normal.pdf"), "normal.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue normal priority: %v", err)
	}

	_, err = qm.EnqueueDocument("TestDeal", filepath.Join(tempDir, "high.pdf"), "high.pdf", PriorityHigh, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue high priority: %v", err)
	}

	_, err = qm.EnqueueDocument("TestDeal", filepath.Join(tempDir, "low.pdf"), "low.pdf", PriorityLow, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue low priority: %v", err)
	}

	// Check order - high priority should be first
	qm.mutex.RLock()
	if qm.queue[0].Priority != PriorityHigh {
		t.Errorf("Expected first item to be high priority, got %v", qm.queue[0].Priority)
	}
	if qm.queue[1].Priority != PriorityNormal {
		t.Errorf("Expected second item to be normal priority, got %v", qm.queue[1].Priority)
	}
	if qm.queue[2].Priority != PriorityLow {
		t.Errorf("Expected third item to be low priority, got %v", qm.queue[2].Priority)
	}
	qm.mutex.RUnlock()
}

func TestQueueManager_QueueStatus(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create and enqueue test documents
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tempDir, "test"+string(rune('0'+i))+".pdf")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
		if err != nil {
			t.Fatalf("Failed to enqueue document: %v", err)
		}
	}

	stats := qm.GetQueueStatus()

	if stats.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", stats.TotalItems)
	}

	if stats.PendingItems != 3 {
		t.Errorf("Expected 3 pending items, got %d", stats.PendingItems)
	}

	if stats.StatusBreakdown[QueueStatusPending] != 3 {
		t.Errorf("Expected 3 pending in breakdown, got %d", stats.StatusBreakdown[QueueStatusPending])
	}
}

func TestQueueManager_QueryQueue(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test documents for different deals
	deals := []string{"Deal1", "Deal2"}
	for _, deal := range deals {
		for i := 0; i < 2; i++ {
			testFile := filepath.Join(tempDir, deal+"_test"+string(rune('0'+i))+".pdf")
			if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err := qm.EnqueueDocument(deal, testFile, "test.pdf", PriorityNormal, nil)
			if err != nil {
				t.Fatalf("Failed to enqueue document: %v", err)
			}
		}
	}

	// Query for Deal1 only
	query := QueueQuery{
		DealName: "Deal1",
		Limit:    10,
	}

	results, err := qm.QueryQueue(query)
	if err != nil {
		t.Fatalf("Failed to query queue: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for Deal1, got %d", len(results))
	}

	for _, item := range results {
		if item.DealName != "Deal1" {
			t.Errorf("Expected Deal1, got %s", item.DealName)
		}
	}
}

func TestQueueManager_DealFolderMirror(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enqueue document
	_, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	// Check deal folder mirror was created
	qm.folderMutex.RLock()
	mirror, exists := qm.dealFolders["TestDeal"]
	qm.folderMutex.RUnlock()

	if !exists {
		t.Fatal("Deal folder mirror should have been created")
	}

	if mirror.DealName != "TestDeal" {
		t.Errorf("Expected deal name TestDeal, got %s", mirror.DealName)
	}

	if len(mirror.FileStructure) != 1 {
		t.Errorf("Expected 1 file in structure, got %d", len(mirror.FileStructure))
	}

	fileInfo, hasFile := mirror.FileStructure[testFile]
	if !hasFile {
		t.Error("Test file should be in file structure")
	}

	if fileInfo.ProcessingState != "queued" {
		t.Errorf("Expected processing state 'queued', got %s", fileInfo.ProcessingState)
	}
}

func TestQueueManager_StateSync(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enqueue document
	item, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	// Simulate workflow state change
	err = qm.SynchronizeWorkflowState(item.JobID, "processing")
	if err != nil {
		t.Fatalf("Failed to synchronize workflow state: %v", err)
	}

	// Check status was updated
	qm.mutex.RLock()
	if item.Status != QueueStatusProcessing {
		t.Errorf("Expected status %s, got %s", QueueStatusProcessing, item.Status)
	}
	qm.mutex.RUnlock()

	// Test completion
	err = qm.SynchronizeWorkflowState(item.JobID, "completed")
	if err != nil {
		t.Fatalf("Failed to synchronize completion: %v", err)
	}

	qm.mutex.RLock()
	if item.Status != QueueStatusCompleted {
		t.Errorf("Expected status %s, got %s", QueueStatusCompleted, item.Status)
	}
	if item.ProcessingEnded == nil {
		t.Error("Processing end time should be set")
	}
	qm.mutex.RUnlock()
}

func TestQueueManager_ProcessingHistory(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Record processing history
	results := map[string]interface{}{
		"templatesUsed":   []string{"template1", "template2"},
		"fieldsExtracted": float64(5),
		"confidenceScore": 0.85,
	}

	qm.RecordProcessingHistory("TestDeal", "/path/to/doc.pdf", "document-analysis", results)

	// Get history
	history := qm.GetProcessingHistory("TestDeal", 10)

	if len(history) != 1 {
		t.Errorf("Expected 1 history item, got %d", len(history))
	}

	item := history[0]
	if item.DealName != "TestDeal" {
		t.Errorf("Expected deal name TestDeal, got %s", item.DealName)
	}

	if item.FieldsExtracted != 5 {
		t.Errorf("Expected 5 fields extracted, got %d", item.FieldsExtracted)
	}

	if item.ConfidenceScore != 0.85 {
		t.Errorf("Expected confidence score 0.85, got %f", item.ConfidenceScore)
	}

	if len(item.TemplatesUsed) != 2 {
		t.Errorf("Expected 2 templates used, got %d", len(item.TemplatesUsed))
	}
}

func TestQueueManager_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enqueue document
	item, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	// Persist state
	err = qm.persistState()
	if err != nil {
		t.Fatalf("Failed to persist state: %v", err)
	}

	// Create new queue manager and load state
	qm2 := NewQueueManager(tempDir)

	// Check that state was loaded
	if len(qm2.queue) != 1 {
		t.Errorf("Expected 1 queue item after loading, got %d", len(qm2.queue))
	}

	if qm2.queue[0].ID != item.ID {
		t.Errorf("Expected item ID %s, got %s", item.ID, qm2.queue[0].ID)
	}

	if len(qm2.dealFolders) != 1 {
		t.Errorf("Expected 1 deal folder after loading, got %d", len(qm2.dealFolders))
	}
}

func TestQueueManager_StartStop(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Test starting
	err := qm.Start()
	if err != nil {
		t.Fatalf("Failed to start queue manager: %v", err)
	}

	if !qm.isRunning {
		t.Error("Queue manager should be running")
	}

	// Test double start
	err = qm.Start()
	if err == nil {
		t.Error("Expected error when starting already running queue manager")
	}

	// Test stopping
	err = qm.Stop()
	if err != nil {
		t.Fatalf("Failed to stop queue manager: %v", err)
	}

	if qm.isRunning {
		t.Error("Queue manager should not be running")
	}
}

func TestQueueManager_SyncDealFolder(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test files in deal folder
	dealDir := filepath.Join(tempDir, "TestDeal")
	if err := os.MkdirAll(dealDir, 0755); err != nil {
		t.Fatalf("Failed to create deal directory: %v", err)
	}

	testFile1 := filepath.Join(dealDir, "doc1.pdf")
	testFile2 := filepath.Join(dealDir, "doc2.pdf")

	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Initialize deal folder mirror
	qm.dealFolders["TestDeal"] = &DealFolderMirror{
		DealName:      "TestDeal",
		FolderPath:    dealDir,
		FileStructure: make(map[string]FileStructInfo),
		SyncStatus:    SyncStatusOutOfSync,
	}

	// Sync folder
	err := qm.SyncDealFolder("TestDeal")
	if err != nil {
		t.Fatalf("Failed to sync deal folder: %v", err)
	}

	// Check sync results
	qm.folderMutex.RLock()
	mirror := qm.dealFolders["TestDeal"]

	if mirror.SyncStatus != SyncStatusSynced {
		t.Errorf("Expected sync status %s, got %s", SyncStatusSynced, mirror.SyncStatus)
	}

	if mirror.FileCount != 2 {
		t.Errorf("Expected 2 files, got %d", mirror.FileCount)
	}

	if _, hasFile1 := mirror.FileStructure[testFile1]; !hasFile1 {
		t.Error("File 1 should be in file structure")
	}

	if _, hasFile2 := mirror.FileStructure[testFile2]; !hasFile2 {
		t.Error("File 2 should be in file structure")
	}
	qm.folderMutex.RUnlock()
}

func TestQueueManager_HealthCheck(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Override timeout for testing
	qm.config.ProcessingTimeout = 100 * time.Millisecond

	// Create test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enqueue document
	item, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	// Simulate processing start
	now := time.Now().Add(-200 * time.Millisecond) // Already timed out
	item.Status = QueueStatusProcessing
	item.ProcessingStarted = &now
	qm.processingCount = 1

	// Run health check
	qm.performHealthCheck()

	// Check that item was marked as failed
	if item.Status != QueueStatusFailed {
		t.Errorf("Expected status %s, got %s", QueueStatusFailed, item.Status)
	}

	if item.LastError == nil {
		t.Error("Last error should be set")
	}

	if item.LastError.ErrorType != "timeout" {
		t.Errorf("Expected error type 'timeout', got %s", item.LastError.ErrorType)
	}

	if qm.processingCount != 0 {
		t.Errorf("Expected processing count 0, got %d", qm.processingCount)
	}
}

func TestQueueManager_CleanupCompletedItems(t *testing.T) {
	tempDir := t.TempDir()
	qm := NewQueueManager(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enqueue and complete document
	item, err := qm.EnqueueDocument("TestDeal", testFile, "test.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document: %v", err)
	}

	// Mark as completed with old timestamp
	oldTime := time.Now().Add(-25 * time.Hour)
	item.Status = QueueStatusCompleted
	item.ProcessingEnded = &oldTime

	// Add a recent item
	testFile2 := filepath.Join(tempDir, "test2.pdf")
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	item2, err := qm.EnqueueDocument("TestDeal", testFile2, "test2.pdf", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Failed to enqueue document 2: %v", err)
	}

	recentTime := time.Now().Add(-1 * time.Hour)
	item2.Status = QueueStatusCompleted
	item2.ProcessingEnded = &recentTime

	// Run cleanup
	qm.cleanupCompletedItems()

	// Check that old item was removed and recent item was kept
	qm.mutex.RLock()
	foundOld := false
	foundRecent := false

	for _, queueItem := range qm.queue {
		if queueItem.ID == item.ID {
			foundOld = true
		}
		if queueItem.ID == item2.ID {
			foundRecent = true
		}
	}
	qm.mutex.RUnlock()

	if foundOld {
		t.Error("Old completed item should have been cleaned up")
	}

	if !foundRecent {
		t.Error("Recent completed item should not have been cleaned up")
	}
}
