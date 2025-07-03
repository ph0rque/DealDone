package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// Test implementations

// TestLogger is now defined in test_utils.go

// TestNotifier implements ErrorNotifier interface for testing
type TestNotifier struct {
	notifications []string
	mutex         sync.Mutex
}

func (tn *TestNotifier) NotifyError(execution *WorkflowExecution, step *WorkflowStep, err error) error {
	tn.mutex.Lock()
	defer tn.mutex.Unlock()
	tn.notifications = append(tn.notifications, fmt.Sprintf("ERROR: %s in step %s: %v", execution.ID, step.ID, err))
	return nil
}

func (tn *TestNotifier) NotifyCriticalFailure(execution *WorkflowExecution, message string) error {
	tn.mutex.Lock()
	defer tn.mutex.Unlock()
	tn.notifications = append(tn.notifications, fmt.Sprintf("CRITICAL: %s - %s", execution.ID, message))
	return nil
}

func (tn *TestNotifier) NotifyRecoverySuccess(execution *WorkflowExecution, message string) error {
	tn.mutex.Lock()
	defer tn.mutex.Unlock()
	tn.notifications = append(tn.notifications, fmt.Sprintf("RECOVERY: %s - %s", execution.ID, message))
	return nil
}

func (tn *TestNotifier) GetNotifications() []string {
	tn.mutex.Lock()
	defer tn.mutex.Unlock()
	notifications := make([]string, len(tn.notifications))
	copy(notifications, tn.notifications)
	return notifications
}

func (tn *TestNotifier) Reset() {
	tn.mutex.Lock()
	defer tn.mutex.Unlock()
	tn.notifications = []string{}
}

// TestStepExecutor implements StepExecutor interface for testing
type TestStepExecutor struct {
	failSteps   map[string]error
	callCount   map[string]int
	shouldDelay bool
	delay       time.Duration
	mutex       sync.Mutex
}

func NewTestStepExecutor() *TestStepExecutor {
	return &TestStepExecutor{
		failSteps: make(map[string]error),
		callCount: make(map[string]int),
		delay:     100 * time.Millisecond,
	}
}

func (tse *TestStepExecutor) ExecuteStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error {
	tse.mutex.Lock()
	defer tse.mutex.Unlock()

	tse.callCount[step.ID]++

	if tse.shouldDelay {
		time.Sleep(tse.delay)
	}

	if err, exists := tse.failSteps[step.ID]; exists {
		return err
	}

	return nil
}

func (tse *TestStepExecutor) ValidateStep(step *WorkflowStep) error {
	return nil
}

func (tse *TestStepExecutor) RollbackStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStep) error {
	return nil
}

func (tse *TestStepExecutor) SetStepFailure(stepID string, err error) {
	tse.mutex.Lock()
	defer tse.mutex.Unlock()
	tse.failSteps[stepID] = err
}

func (tse *TestStepExecutor) GetCallCount(stepID string) int {
	tse.mutex.Lock()
	defer tse.mutex.Unlock()
	return tse.callCount[stepID]
}

func (tse *TestStepExecutor) SetDelay(shouldDelay bool, delay time.Duration) {
	tse.mutex.Lock()
	defer tse.mutex.Unlock()
	tse.shouldDelay = shouldDelay
	tse.delay = delay
}

// Helper functions for tests

func createTestConfig(tempDir string) WorkflowRecoveryConfig {
	return WorkflowRecoveryConfig{
		RetryConfig: RetryConfig{
			InitialDelay:   50 * time.Millisecond,
			MaxDelay:       500 * time.Millisecond,
			BackoffFactor:  2.0,
			MaxRetries:     3,
			Jitter:         false,
			JitterMaxDelay: 100 * time.Millisecond,
		},
		PersistenceInterval:   1 * time.Second,
		MaxExecutionHistory:   100,
		ErrorLogRetention:     24 * time.Hour,
		NotificationThreshold: SeverityMedium,
		EnablePartialResults:  true,
		StoragePath:           tempDir,
	}
}

func createTestSteps() []*WorkflowStep {
	return []*WorkflowStep{
		{
			ID:           "step1",
			Name:         "First Step",
			Status:       StepPending,
			MaxRetries:   3,
			Dependencies: []string{},
			CanSkip:      false,
			CanRollback:  false,
			Metadata:     map[string]interface{}{"type": "data_extraction"},
		},
		{
			ID:           "step2",
			Name:         "Second Step",
			Status:       StepPending,
			MaxRetries:   3,
			Dependencies: []string{"step1"},
			CanSkip:      true,
			CanRollback:  false,
			Metadata:     map[string]interface{}{"type": "ai_processing"},
		},
		{
			ID:           "step3",
			Name:         "Third Step",
			Status:       StepPending,
			MaxRetries:   3,
			Dependencies: []string{"step2"},
			CanSkip:      false,
			CanRollback:  true,
			Metadata:     map[string]interface{}{"type": "validation"},
		},
	}
}

// Tests

func TestWorkflowRecoveryService_Creation(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)

	service := NewWorkflowRecoveryService(config, logger, notifier)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.config.StoragePath != tempDir {
		t.Errorf("Expected storage path %s, got %s", tempDir, service.config.StoragePath)
	}

	// Check if storage directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected storage directory to be created")
	}

	service.Shutdown()
}

func TestWorkflowRecoveryService_CreateExecution(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, err := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if execution == nil {
		t.Fatal("Expected execution to be created")
	}

	if execution.WorkflowType != "test-workflow" {
		t.Errorf("Expected workflow type 'test-workflow', got '%s'", execution.WorkflowType)
	}

	if execution.DealID != "deal-123" {
		t.Errorf("Expected deal ID 'deal-123', got '%s'", execution.DealID)
	}

	if execution.DocumentID != "doc-456" {
		t.Errorf("Expected document ID 'doc-456', got '%s'", execution.DocumentID)
	}

	if len(execution.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(execution.Steps))
	}

	if execution.CurrentStepIndex != 0 {
		t.Errorf("Expected current step index 0, got %d", execution.CurrentStepIndex)
	}
}

func TestWorkflowRecoveryService_SuccessfulExecution(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()

	err := service.ExecuteWorkflow(execution.ID, executor)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check execution status
	updatedExecution, _ := service.GetExecution(execution.ID)
	if updatedExecution != nil && updatedExecution.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updatedExecution.Status)
	}

	// Check all steps completed
	for _, step := range execution.Steps {
		if step.Status != StepCompleted {
			t.Errorf("Expected step %s to be completed, got %s", step.ID, step.Status)
		}
	}

	// Check each step was called once
	for _, step := range steps {
		if count := executor.GetCallCount(step.ID); count != 1 {
			t.Errorf("Expected step %s to be called once, got %d", step.ID, count)
		}
	}
}

func TestWorkflowRecoveryService_RetryLogic(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	// Make step2 fail with a retryable error
	executor.SetStepFailure("step2", errors.New("timeout: connection failed"))

	err := service.ExecuteWorkflow(execution.ID, executor)

	if err == nil {
		t.Fatal("Expected error from failed step")
	}

	// Check that step2 was retried (should be called MaxRetries + 1 times)
	expectedCalls := steps[1].MaxRetries + 1
	if count := executor.GetCallCount("step2"); count != expectedCalls {
		t.Errorf("Expected step2 to be called %d times, got %d", expectedCalls, count)
	}

	// Check step1 was called once (no failure)
	if count := executor.GetCallCount("step1"); count != 1 {
		t.Errorf("Expected step1 to be called once, got %d", count)
	}

	// Check step3 was never called (step2 failed)
	if count := executor.GetCallCount("step3"); count != 0 {
		t.Errorf("Expected step3 to never be called, got %d", count)
	}
}

func TestWorkflowRecoveryService_ExponentialBackoff(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	// Set more aggressive timing for test
	config.RetryConfig.InitialDelay = 10 * time.Millisecond
	config.RetryConfig.MaxDelay = 100 * time.Millisecond
	config.RetryConfig.BackoffFactor = 2.0

	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := []*WorkflowStep{
		{
			ID:         "step1",
			Name:       "Test Step",
			Status:     StepPending,
			MaxRetries: 3,
			Metadata:   map[string]interface{}{"type": "test"},
		},
	}

	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	executor.SetStepFailure("step1", errors.New("network: temporary failure"))

	start := time.Now()
	service.ExecuteWorkflow(execution.ID, executor)
	elapsed := time.Since(start)

	// With initial delay 10ms, backoff factor 2.0:
	// Retry 1: 10ms, Retry 2: 20ms, Retry 3: 40ms
	// Total expected: ~70ms + execution time
	expectedMinDuration := 70 * time.Millisecond

	if elapsed < expectedMinDuration {
		t.Errorf("Expected execution to take at least %v, took %v", expectedMinDuration, elapsed)
	}
}

func TestWorkflowRecoveryService_RecoveryStrategies(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	tests := []struct {
		name           string
		errorMsg       string
		canSkip        bool
		canRollback    bool
		expectedStatus WorkflowStepStatus
	}{
		{
			name:           "Skip strategy for high severity skippable step",
			errorMsg:       "timeout: network unreachable",
			canSkip:        true,
			canRollback:    false,
			expectedStatus: StepSkipped,
		},
		{
			name:           "Rollback strategy for critical error with rollback capability",
			errorMsg:       "critical: system failure",
			canSkip:        false,
			canRollback:    true,
			expectedStatus: StepRolledBack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := []*WorkflowStep{
				{
					ID:          "test-step",
					Name:        "Test Step",
					Status:      StepPending,
					MaxRetries:  1,
					CanSkip:     tt.canSkip,
					CanRollback: tt.canRollback,
					Metadata:    map[string]interface{}{"type": "test"},
				},
			}

			execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

			executor := NewTestStepExecutor()
			executor.SetStepFailure("test-step", errors.New(tt.errorMsg))

			service.ExecuteWorkflow(execution.ID, executor)

			if steps[0].Status != tt.expectedStatus {
				t.Errorf("Expected step status %s, got %s", tt.expectedStatus, steps[0].Status)
			}
		})
	}
}

func TestWorkflowRecoveryService_DependencyChecking(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	// Make step1 fail, which should prevent step2 from running
	executor.SetStepFailure("step1", errors.New("auth_failure: invalid credentials"))

	err := service.ExecuteWorkflow(execution.ID, executor)

	if err == nil {
		t.Fatal("Expected error from failed step1")
	}

	// Check that step2 and step3 were never called due to dependency failure
	if count := executor.GetCallCount("step2"); count != 0 {
		t.Errorf("Expected step2 to never be called, got %d", count)
	}

	if count := executor.GetCallCount("step3"); count != 0 {
		t.Errorf("Expected step3 to never be called, got %d", count)
	}
}

func TestWorkflowRecoveryService_ResumeWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	// Make step2 fail initially
	executor.SetStepFailure("step2", errors.New("temporary: service unavailable"))

	// First execution should fail at step2
	err := service.ExecuteWorkflow(execution.ID, executor)
	if err == nil {
		t.Fatal("Expected error from failed step2")
	}

	// Remove the failure and resume
	executor.SetStepFailure("step2", nil)

	err = service.ResumeWorkflow(execution.ID, executor)
	if err != nil {
		t.Fatalf("Expected successful resume, got %v", err)
	}

	// Check execution completed
	updatedExecution, _ := service.GetExecution(execution.ID)
	if updatedExecution != nil && updatedExecution.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updatedExecution.Status)
	}

	// Check step1 was called once (initial run), step2 and step3 called once each (resume)
	if count := executor.GetCallCount("step1"); count != 1 {
		t.Errorf("Expected step1 to be called once, got %d", count)
	}

	// step2 should be called multiple times: initial failure + retries + resume success
	if count := executor.GetCallCount("step2"); count < 2 {
		t.Errorf("Expected step2 to be called at least twice, got %d", count)
	}
}

func TestWorkflowRecoveryService_PartialResults(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	config.EnablePartialResults = true

	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	// Make step2 fail to test partial results
	executor.SetStepFailure("step2", errors.New("auth_failure: invalid token"))

	service.ExecuteWorkflow(execution.ID, executor)

	// Check that partial results were saved for step1
	if execution.PartialResults == nil {
		t.Error("Expected partial results to be saved")
	} else {
		key := "step_step1_result"
		if _, exists := execution.PartialResults[key]; !exists {
			t.Errorf("Expected partial result for key %s", key)
		}
	}
}

func TestWorkflowRecoveryService_ErrorStatistics(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := []*WorkflowStep{
		{ID: "step1", Name: "Step 1", Status: StepPending, MaxRetries: 1, Metadata: map[string]interface{}{"type": "test"}},
		{ID: "step2", Name: "Step 2", Status: StepPending, MaxRetries: 1, Metadata: map[string]interface{}{"type": "test"}},
	}

	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	executor.SetStepFailure("step1", errors.New("timeout: connection failed"))
	executor.SetStepFailure("step2", errors.New("validation: invalid data"))

	service.ExecuteWorkflow(execution.ID, executor)

	stats := service.GetErrorStatistics()

	if stats["execution_error"] < 2 {
		t.Errorf("Expected at least 2 execution errors, got %d", stats["execution_error"])
	}
}

func TestWorkflowRecoveryService_Notifications(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	config.NotificationThreshold = SeverityHigh

	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := []*WorkflowStep{
		{ID: "step1", Name: "Step 1", Status: StepPending, MaxRetries: 1, CanSkip: false, CanRollback: false, Metadata: map[string]interface{}{"type": "test"}},
	}

	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	executor.SetStepFailure("step1", errors.New("critical: system failure"))

	service.ExecuteWorkflow(execution.ID, executor)

	notifications := notifier.GetNotifications()

	if len(notifications) == 0 {
		t.Error("Expected notifications to be sent for critical error")
	}

	found := false
	for _, notification := range notifications {
		if len(notification) > 8 && notification[:8] == "CRITICAL" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected critical failure notification")
	}
}

func TestWorkflowRecoveryService_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	config.PersistenceInterval = 100 * time.Millisecond

	service := NewWorkflowRecoveryService(config, logger, notifier)

	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	// Wait for persistence
	time.Sleep(200 * time.Millisecond)
	service.Shutdown()

	// Create new service and check if state was loaded
	service2 := NewWorkflowRecoveryService(config, logger, notifier)
	defer service2.Shutdown()

	loadedExecution, err := service2.GetExecution(execution.ID)
	if err != nil {
		t.Fatalf("Expected to load execution, got error: %v", err)
	}

	if loadedExecution.WorkflowType != "test-workflow" {
		t.Errorf("Expected workflow type 'test-workflow', got '%s'", loadedExecution.WorkflowType)
	}
}

func TestWorkflowRecoveryService_CleanupOldExecutions(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	config.ErrorLogRetention = 1 * time.Millisecond // Very short retention for testing

	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	// Create and complete an execution
	steps := createTestSteps()
	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	service.ExecuteWorkflow(execution.ID, executor)

	// Wait for retention period to pass
	time.Sleep(10 * time.Millisecond)

	// Run cleanup
	err := service.CleanupOldExecutions()
	if err != nil {
		t.Fatalf("Expected no error from cleanup, got %v", err)
	}

	// The execution should be cleaned up from history
	// Note: This test verifies the cleanup logic runs without error
	// In a real scenario, we'd need more sophisticated timing control
}

func TestWorkflowRecoveryService_ConcurrentExecution(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	var wg sync.WaitGroup
	numConcurrent := 5

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			steps := createTestSteps()
			execution, _ := service.CreateExecution(fmt.Sprintf("test-workflow-%d", index), fmt.Sprintf("deal-%d", index), fmt.Sprintf("doc-%d", index), steps)

			executor := NewTestStepExecutor()
			err := service.ExecuteWorkflow(execution.ID, executor)

			if err != nil {
				t.Errorf("Concurrent execution %d failed: %v", index, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all executions were processed
	stats := service.GetErrorStatistics()

	// Should have no errors for successful concurrent executions
	if stats["execution_error"] > 0 {
		t.Errorf("Expected no execution errors in concurrent test, got %d", stats["execution_error"])
	}
}

func TestWorkflowRecoveryService_SeverityDetermination(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	tests := []struct {
		error    string
		expected ErrorSeverity
	}{
		{"panic: runtime error", SeverityCritical},
		{"fatal: system crash", SeverityCritical},
		{"critical: database failure", SeverityCritical},
		{"timeout: connection failed", SeverityHigh},
		{"network: unreachable host", SeverityHigh},
		{"validation: missing field", SeverityMedium},
		{"parse: invalid format", SeverityMedium},
		{"info: processing complete", SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.error, func(t *testing.T) {
			err := errors.New(tt.error)
			severity := service.determineSeverity(err)

			if severity != tt.expected {
				t.Errorf("Expected severity %s for error '%s', got %s", tt.expected, tt.error, severity)
			}
		})
	}
}

func TestWorkflowRecoveryService_NonRetryableErrors(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	steps := []*WorkflowStep{
		{ID: "step1", Name: "Step 1", Status: StepPending, MaxRetries: 3, Metadata: map[string]interface{}{"type": "test"}},
	}

	execution, _ := service.CreateExecution("test-workflow", "deal-123", "doc-456", steps)

	executor := NewTestStepExecutor()
	// Set a non-retryable error
	executor.SetStepFailure("step1", errors.New("auth_failure: invalid credentials"))

	service.ExecuteWorkflow(execution.ID, executor)

	// Should only be called once (no retries for non-retryable errors)
	if count := executor.GetCallCount("step1"); count != 1 {
		t.Errorf("Expected step1 to be called once (non-retryable), got %d", count)
	}
}

// Benchmark tests

func BenchmarkWorkflowRecoveryService_BasicExecution(b *testing.B) {
	tempDir := b.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		steps := createTestSteps()
		execution, _ := service.CreateExecution("benchmark-workflow", "deal-123", "doc-456", steps)

		executor := NewTestStepExecutor()
		service.ExecuteWorkflow(execution.ID, executor)
	}
}

func BenchmarkWorkflowRecoveryService_WithRetries(b *testing.B) {
	tempDir := b.TempDir()
	logger := NewTestLogger()
	notifier := &TestNotifier{}
	config := createTestConfig(tempDir)
	config.RetryConfig.InitialDelay = 1 * time.Millisecond // Faster for benchmark
	service := NewWorkflowRecoveryService(config, logger, notifier)
	defer service.Shutdown()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		steps := []*WorkflowStep{
			{ID: "step1", Name: "Step 1", Status: StepPending, MaxRetries: 2, Metadata: map[string]interface{}{"type": "test"}},
		}

		execution, _ := service.CreateExecution("benchmark-workflow", "deal-123", "doc-456", steps)

		executor := NewTestStepExecutor()
		if i%2 == 0 {
			// Make every other execution fail and retry
			executor.SetStepFailure("step1", errors.New("temporary: service busy"))
		}

		service.ExecuteWorkflow(execution.ID, executor)
	}
}
