package performance

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// WorkflowPerformanceEnhancer optimizes n8n workflow execution performance
type WorkflowPerformanceEnhancer struct {
	config             *WorkflowEnhancerConfig
	batchProcessor     *BatchProcessor
	connectionPool     *ConnectionPool
	payloadCompressor  *PayloadCompressor
	workflowCache      *WorkflowCache
	performanceMonitor *WorkflowPerformanceMonitor
	loadBalancer       *LoadBalancer
	metrics            *WorkflowPerformanceMetrics
	mu                 sync.RWMutex
}

// WorkflowEnhancerConfig contains configuration for workflow optimization
type WorkflowEnhancerConfig struct {
	MaxConcurrentRequests int           `json:"maxConcurrentRequests"`
	BatchSize             int           `json:"batchSize"`
	BatchTimeout          time.Duration `json:"batchTimeout"`
	CompressionEnabled    bool          `json:"compressionEnabled"`
	CacheEnabled          bool          `json:"cacheEnabled"`
	CacheTTL              time.Duration `json:"cacheTTL"`
	ConnectionPoolSize    int           `json:"connectionPoolSize"`
	RequestTimeout        time.Duration `json:"requestTimeout"`
	RetryAttempts         int           `json:"retryAttempts"`
	RetryDelay            time.Duration `json:"retryDelay"`
	MonitoringEnabled     bool          `json:"monitoringEnabled"`
	LoadBalancingEnabled  bool          `json:"loadBalancingEnabled"`
	N8nEndpoints          []string      `json:"n8nEndpoints"`
}

// BatchProcessor handles batch processing of workflow requests
type BatchProcessor struct {
	batches         map[string]*WorkflowBatch
	batchTimeout    time.Duration
	maxBatchSize    int
	processingQueue chan *WorkflowBatch
	metrics         *BatchProcessingMetrics
	mu              sync.RWMutex
}

// WorkflowBatch represents a batch of workflow requests
type WorkflowBatch struct {
	BatchID      string                 `json:"batchId"`
	WorkflowType string                 `json:"workflowType"`
	Requests     []*WorkflowRequest     `json:"requests"`
	Status       string                 `json:"status"` // "pending", "processing", "completed", "failed"
	CreatedAt    time.Time              `json:"createdAt"`
	ProcessedAt  time.Time              `json:"processedAt"`
	CompletedAt  time.Time              `json:"completedAt"`
	Duration     time.Duration          `json:"duration"`
	Results      []*WorkflowResult      `json:"results"`
	Error        error                  `json:"error"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// WorkflowRequest represents a single workflow request
type WorkflowRequest struct {
	RequestID    string                 `json:"requestId"`
	WorkflowType string                 `json:"workflowType"`
	Payload      map[string]interface{} `json:"payload"`
	Priority     int                    `json:"priority"`
	CreatedAt    time.Time              `json:"createdAt"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// WorkflowResult represents the result of a workflow execution
type WorkflowResult struct {
	RequestID  string                 `json:"requestId"`
	Status     string                 `json:"status"`
	Result     map[string]interface{} `json:"result"`
	Error      error                  `json:"error"`
	Duration   time.Duration          `json:"duration"`
	ExecutedAt time.Time              `json:"executedAt"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ConnectionPool manages HTTP connections to n8n endpoints
type ConnectionPool struct {
	clients      []*http.Client
	endpoints    []string
	currentIndex int
	poolSize     int
	mu           sync.RWMutex
}

// PayloadCompressor handles compression of workflow payloads
type PayloadCompressor struct {
	compressionEnabled bool
	compressionLevel   int
	metrics            *CompressionMetrics
}

// CompressionMetrics tracks compression performance
type CompressionMetrics struct {
	TotalRequests      int64         `json:"totalRequests"`
	CompressedRequests int64         `json:"compressedRequests"`
	CompressionRatio   float64       `json:"compressionRatio"`
	BytesSaved         int64         `json:"bytesSaved"`
	CompressionTime    time.Duration `json:"compressionTime"`
}

// WorkflowCache caches workflow results for performance
type WorkflowCache struct {
	entries   map[string]*WorkflowCacheEntry
	maxSize   int
	ttl       time.Duration
	hitCount  int64
	missCount int64
	mu        sync.RWMutex
}

// WorkflowCacheEntry represents a cached workflow result
type WorkflowCacheEntry struct {
	Key         string                 `json:"key"`
	Result      *WorkflowResult        `json:"result"`
	CreatedAt   time.Time              `json:"createdAt"`
	ExpiresAt   time.Time              `json:"expiresAt"`
	AccessCount int                    `json:"accessCount"`
	LastAccess  time.Time              `json:"lastAccess"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowPerformanceMonitor monitors workflow execution performance
type WorkflowPerformanceMonitor struct {
	enabled            bool
	metrics            *WorkflowPerformanceMetrics
	executionHistory   []*WorkflowExecution
	bottleneckDetector *BottleneckDetector
	alertManager       *AlertManager
	mu                 sync.RWMutex
}

// WorkflowExecution represents a single workflow execution record
type WorkflowExecution struct {
	ExecutionID   string                 `json:"executionId"`
	WorkflowType  string                 `json:"workflowType"`
	StartTime     time.Time              `json:"startTime"`
	EndTime       time.Time              `json:"endTime"`
	Duration      time.Duration          `json:"duration"`
	Status        string                 `json:"status"`
	NodeMetrics   []*NodeMetrics         `json:"nodeMetrics"`
	ResourceUsage *ResourceUsage         `json:"resourceUsage"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NodeMetrics tracks performance of individual workflow nodes
type NodeMetrics struct {
	NodeID        string        `json:"nodeId"`
	NodeType      string        `json:"nodeType"`
	ExecutionTime time.Duration `json:"executionTime"`
	Status        string        `json:"status"`
	InputSize     int64         `json:"inputSize"`
	OutputSize    int64         `json:"outputSize"`
	MemoryUsage   int64         `json:"memoryUsage"`
	ErrorCount    int           `json:"errorCount"`
}

// ResourceUsage tracks resource consumption during workflow execution
type ResourceUsage struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage int64   `json:"memoryUsage"`
	NetworkIO   int64   `json:"networkIO"`
	DiskIO      int64   `json:"diskIO"`
}

// LoadBalancer distributes requests across multiple n8n endpoints
type LoadBalancer struct {
	endpoints       []string
	healthChecks    map[string]*EndpointHealth
	currentIndex    int
	balancingMethod string // "round_robin", "least_connections", "weighted"
	mu              sync.RWMutex
}

// EndpointHealth tracks the health status of n8n endpoints
type EndpointHealth struct {
	Endpoint          string        `json:"endpoint"`
	Status            string        `json:"status"` // "healthy", "unhealthy", "unknown"
	ResponseTime      time.Duration `json:"responseTime"`
	ErrorRate         float64       `json:"errorRate"`
	ActiveConnections int           `json:"activeConnections"`
	LastHealthCheck   time.Time     `json:"lastHealthCheck"`
}

// WorkflowPerformanceMetrics tracks overall workflow performance
type WorkflowPerformanceMetrics struct {
	TotalExecutions      int64         `json:"totalExecutions"`
	SuccessfulExecutions int64         `json:"successfulExecutions"`
	FailedExecutions     int64         `json:"failedExecutions"`
	AverageExecutionTime time.Duration `json:"averageExecutionTime"`
	MedianExecutionTime  time.Duration `json:"medianExecutionTime"`
	P95ExecutionTime     time.Duration `json:"p95ExecutionTime"`
	P99ExecutionTime     time.Duration `json:"p99ExecutionTime"`
	ThroughputPerMinute  float64       `json:"throughputPerMinute"`
	ErrorRate            float64       `json:"errorRate"`
	CacheHitRate         float64       `json:"cacheHitRate"`
	CompressionSavings   float64       `json:"compressionSavings"`
	BatchEfficiency      float64       `json:"batchEfficiency"`
	LastUpdated          time.Time     `json:"lastUpdated"`
}

// BatchProcessingMetrics tracks batch processing performance
type BatchProcessingMetrics struct {
	TotalBatches     int64         `json:"totalBatches"`
	CompletedBatches int64         `json:"completedBatches"`
	AverageBatchSize float64       `json:"averageBatchSize"`
	AverageBatchTime time.Duration `json:"averageBatchTime"`
	BatchThroughput  float64       `json:"batchThroughput"`
	BatchEfficiency  float64       `json:"batchEfficiency"`
}

// BottleneckDetector identifies performance bottlenecks
type BottleneckDetector struct {
	enabled             bool
	thresholds          map[string]float64
	detectedBottlenecks []*Bottleneck
	mu                  sync.RWMutex
}

// Bottleneck represents a detected performance bottleneck
type Bottleneck struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Impact      float64                `json:"impact"`
	Suggestions []string               `json:"suggestions"`
	DetectedAt  time.Time              `json:"detectedAt"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertManager manages performance alerts
type AlertManager struct {
	enabled    bool
	alerts     []*PerformanceAlert
	thresholds map[string]float64
	mu         sync.RWMutex
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	AlertID    string                 `json:"alertId"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Message    string                 `json:"message"`
	Value      float64                `json:"value"`
	Threshold  float64                `json:"threshold"`
	CreatedAt  time.Time              `json:"createdAt"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt time.Time              `json:"resolvedAt"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewWorkflowPerformanceEnhancer creates a new workflow performance enhancer
func NewWorkflowPerformanceEnhancer(config *WorkflowEnhancerConfig) *WorkflowPerformanceEnhancer {
	if config == nil {
		config = &WorkflowEnhancerConfig{
			MaxConcurrentRequests: 50,
			BatchSize:             10,
			BatchTimeout:          5 * time.Second,
			CompressionEnabled:    true,
			CacheEnabled:          true,
			CacheTTL:              30 * time.Minute,
			ConnectionPoolSize:    20,
			RequestTimeout:        30 * time.Second,
			RetryAttempts:         3,
			RetryDelay:            1 * time.Second,
			MonitoringEnabled:     true,
			LoadBalancingEnabled:  true,
			N8nEndpoints:          []string{"http://localhost:5678"},
		}
	}

	enhancer := &WorkflowPerformanceEnhancer{
		config:             config,
		batchProcessor:     NewBatchProcessor(config.BatchSize, config.BatchTimeout),
		connectionPool:     NewConnectionPool(config.ConnectionPoolSize, config.N8nEndpoints),
		payloadCompressor:  NewPayloadCompressor(config.CompressionEnabled),
		workflowCache:      NewWorkflowCache(1000, config.CacheTTL),
		performanceMonitor: NewWorkflowPerformanceMonitor(config.MonitoringEnabled),
		loadBalancer:       NewLoadBalancer(config.N8nEndpoints),
		metrics:            &WorkflowPerformanceMetrics{LastUpdated: time.Now()},
	}

	// Start background processes
	if config.MonitoringEnabled {
		go enhancer.startPerformanceMonitoring()
	}

	return enhancer
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(maxBatchSize int, batchTimeout time.Duration) *BatchProcessor {
	processor := &BatchProcessor{
		batches:         make(map[string]*WorkflowBatch),
		batchTimeout:    batchTimeout,
		maxBatchSize:    maxBatchSize,
		processingQueue: make(chan *WorkflowBatch, 100),
		metrics:         &BatchProcessingMetrics{},
	}

	// Start batch processing worker
	go processor.processBatches()

	return processor
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(poolSize int, endpoints []string) *ConnectionPool {
	pool := &ConnectionPool{
		clients:   make([]*http.Client, poolSize),
		endpoints: endpoints,
		poolSize:  poolSize,
	}

	// Initialize HTTP clients
	for i := 0; i < poolSize; i++ {
		pool.clients[i] = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		}
	}

	return pool
}

// NewPayloadCompressor creates a new payload compressor
func NewPayloadCompressor(enabled bool) *PayloadCompressor {
	return &PayloadCompressor{
		compressionEnabled: enabled,
		compressionLevel:   gzip.DefaultCompression,
		metrics:            &CompressionMetrics{},
	}
}

// NewWorkflowCache creates a new workflow cache
func NewWorkflowCache(maxSize int, ttl time.Duration) *WorkflowCache {
	cache := &WorkflowCache{
		entries: make(map[string]*WorkflowCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cache cleanup goroutine
	go cache.startCleanup()

	return cache
}

// NewWorkflowPerformanceMonitor creates a new workflow performance monitor
func NewWorkflowPerformanceMonitor(enabled bool) *WorkflowPerformanceMonitor {
	return &WorkflowPerformanceMonitor{
		enabled:            enabled,
		metrics:            &WorkflowPerformanceMetrics{LastUpdated: time.Now()},
		executionHistory:   make([]*WorkflowExecution, 0),
		bottleneckDetector: NewBottleneckDetector(),
		alertManager:       NewAlertManager(),
	}
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(endpoints []string) *LoadBalancer {
	lb := &LoadBalancer{
		endpoints:       endpoints,
		healthChecks:    make(map[string]*EndpointHealth),
		balancingMethod: "round_robin",
	}

	// Initialize health checks
	for _, endpoint := range endpoints {
		lb.healthChecks[endpoint] = &EndpointHealth{
			Endpoint:        endpoint,
			Status:          "unknown",
			LastHealthCheck: time.Now(),
		}
	}

	// Start health check monitoring
	go lb.startHealthChecks()

	return lb
}

// NewBottleneckDetector creates a new bottleneck detector
func NewBottleneckDetector() *BottleneckDetector {
	return &BottleneckDetector{
		enabled: true,
		thresholds: map[string]float64{
			"execution_time": 30.0, // seconds
			"error_rate":     0.05, // 5%
			"memory_usage":   0.80, // 80%
			"cpu_usage":      0.75, // 75%
		},
		detectedBottlenecks: make([]*Bottleneck, 0),
	}
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		enabled: true,
		alerts:  make([]*PerformanceAlert, 0),
		thresholds: map[string]float64{
			"error_rate":     0.10, // 10%
			"response_time":  60.0, // seconds
			"throughput":     0.5,  // requests per second
			"cache_hit_rate": 0.70, // 70%
		},
	}
}

// ExecuteWorkflow executes a workflow with performance optimization
func (wpe *WorkflowPerformanceEnhancer) ExecuteWorkflow(ctx context.Context, workflowType string, payload map[string]interface{}) (*WorkflowResult, error) {
	startTime := time.Now()

	// Create workflow request
	request := &WorkflowRequest{
		RequestID:    fmt.Sprintf("req_%d", time.Now().UnixNano()),
		WorkflowType: workflowType,
		Payload:      payload,
		Priority:     1,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Check cache first
	if wpe.config.CacheEnabled {
		if cachedResult := wpe.workflowCache.Get(request.RequestID); cachedResult != nil {
			return cachedResult, nil
		}
	}

	// Execute workflow
	result, err := wpe.executeWorkflowRequest(ctx, request)

	// Cache result if successful
	if err == nil && wpe.config.CacheEnabled {
		wpe.workflowCache.Set(request.RequestID, result)
	}

	// Update metrics
	wpe.updateMetrics(startTime, result, err)

	return result, err
}

// ExecuteBatchWorkflow executes multiple workflows in batch
func (wpe *WorkflowPerformanceEnhancer) ExecuteBatchWorkflow(ctx context.Context, requests []*WorkflowRequest) ([]*WorkflowResult, error) {
	// Group requests by workflow type for optimal batching
	batches := wpe.groupRequestsByType(requests)

	var allResults []*WorkflowResult
	var errors []error

	// Process batches concurrently
	resultChan := make(chan []*WorkflowResult, len(batches))
	errorChan := make(chan error, len(batches))

	for _, batch := range batches {
		go func(b *WorkflowBatch) {
			results, err := wpe.processBatch(ctx, b)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- results
		}(batch)
	}

	// Collect results
	for i := 0; i < len(batches); i++ {
		select {
		case results := <-resultChan:
			allResults = append(allResults, results...)
		case err := <-errorChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if len(errors) > 0 {
		return allResults, fmt.Errorf("batch execution errors: %v", errors)
	}

	return allResults, nil
}

// executeWorkflowRequest executes a single workflow request
func (wpe *WorkflowPerformanceEnhancer) executeWorkflowRequest(ctx context.Context, request *WorkflowRequest) (*WorkflowResult, error) {
	// Get optimal endpoint from load balancer
	endpoint := wpe.loadBalancer.GetNextEndpoint()
	if endpoint == "" {
		return nil, fmt.Errorf("no healthy endpoints available")
	}

	// Compress payload if enabled
	payload := request.Payload
	if wpe.config.CompressionEnabled {
		compressedPayload, err := wpe.payloadCompressor.Compress(payload)
		if err != nil {
			log.Printf("Failed to compress payload: %v", err)
		} else {
			payload = compressedPayload
		}
	}

	// Get HTTP client from pool
	client := wpe.connectionPool.GetClient()

	// Create HTTP request
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if wpe.config.CompressionEnabled {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// Execute request with retry logic
	var response *http.Response
	for attempt := 0; attempt < wpe.config.RetryAttempts; attempt++ {
		response, err = client.Do(req)
		if err == nil && response.StatusCode < 500 {
			break
		}

		if attempt < wpe.config.RetryAttempts-1 {
			time.Sleep(wpe.config.RetryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer response.Body.Close()

	// Read response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Create workflow result
	workflowResult := &WorkflowResult{
		RequestID:  request.RequestID,
		Status:     "completed",
		Result:     result,
		Duration:   time.Since(request.CreatedAt),
		ExecutedAt: time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	if response.StatusCode >= 400 {
		workflowResult.Status = "failed"
		workflowResult.Error = fmt.Errorf("workflow execution failed with status %d", response.StatusCode)
	}

	return workflowResult, nil
}

// groupRequestsByType groups requests by workflow type for batching
func (wpe *WorkflowPerformanceEnhancer) groupRequestsByType(requests []*WorkflowRequest) []*WorkflowBatch {
	batches := make(map[string]*WorkflowBatch)

	for _, request := range requests {
		if batch, exists := batches[request.WorkflowType]; exists {
			batch.Requests = append(batch.Requests, request)
		} else {
			batches[request.WorkflowType] = &WorkflowBatch{
				BatchID:      fmt.Sprintf("batch_%s_%d", request.WorkflowType, time.Now().UnixNano()),
				WorkflowType: request.WorkflowType,
				Requests:     []*WorkflowRequest{request},
				Status:       "pending",
				CreatedAt:    time.Now(),
				Metadata:     make(map[string]interface{}),
			}
		}
	}

	var result []*WorkflowBatch
	for _, batch := range batches {
		result = append(result, batch)
	}

	return result
}

// processBatch processes a batch of workflow requests
func (wpe *WorkflowPerformanceEnhancer) processBatch(ctx context.Context, batch *WorkflowBatch) ([]*WorkflowResult, error) {
	batch.Status = "processing"
	batch.ProcessedAt = time.Now()

	var results []*WorkflowResult
	for _, request := range batch.Requests {
		result, err := wpe.executeWorkflowRequest(ctx, request)
		if err != nil {
			result = &WorkflowResult{
				RequestID:  request.RequestID,
				Status:     "failed",
				Error:      err,
				Duration:   time.Since(request.CreatedAt),
				ExecutedAt: time.Now(),
				Metadata:   make(map[string]interface{}),
			}
		}
		results = append(results, result)
	}

	batch.Results = results
	batch.Status = "completed"
	batch.CompletedAt = time.Now()
	batch.Duration = batch.CompletedAt.Sub(batch.ProcessedAt)

	return results, nil
}

// updateMetrics updates performance metrics
func (wpe *WorkflowPerformanceEnhancer) updateMetrics(startTime time.Time, result *WorkflowResult, err error) {
	wpe.mu.Lock()
	defer wpe.mu.Unlock()

	wpe.metrics.TotalExecutions++
	if err == nil && result.Status == "completed" {
		wpe.metrics.SuccessfulExecutions++
	} else {
		wpe.metrics.FailedExecutions++
	}

	duration := time.Since(startTime)
	wpe.metrics.AverageExecutionTime = (wpe.metrics.AverageExecutionTime + duration) / 2
	wpe.metrics.ErrorRate = float64(wpe.metrics.FailedExecutions) / float64(wpe.metrics.TotalExecutions)
	wpe.metrics.LastUpdated = time.Now()

	// Calculate throughput
	if wpe.metrics.TotalExecutions > 0 {
		elapsed := time.Since(wpe.metrics.LastUpdated)
		if elapsed > 0 {
			wpe.metrics.ThroughputPerMinute = float64(wpe.metrics.TotalExecutions) / elapsed.Minutes()
		}
	}
}

// startPerformanceMonitoring starts background performance monitoring
func (wpe *WorkflowPerformanceEnhancer) startPerformanceMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		wpe.performanceMonitor.collectMetrics(wpe.metrics)
		wpe.performanceMonitor.detectBottlenecks(wpe.metrics)
		wpe.performanceMonitor.checkAlerts(wpe.metrics)
	}
}

// Batch Processor methods

// processBatches processes batches from the queue
func (bp *BatchProcessor) processBatches() {
	for batch := range bp.processingQueue {
		bp.processBatch(batch)
	}
}

// processBatch processes a single batch
func (bp *BatchProcessor) processBatch(batch *WorkflowBatch) {
	bp.mu.Lock()
	bp.metrics.TotalBatches++
	bp.metrics.AverageBatchSize = (bp.metrics.AverageBatchSize + float64(len(batch.Requests))) / 2
	bp.mu.Unlock()

	// Process batch logic would go here
	batch.Status = "completed"
	batch.CompletedAt = time.Now()
	batch.Duration = batch.CompletedAt.Sub(batch.ProcessedAt)

	bp.mu.Lock()
	bp.metrics.CompletedBatches++
	bp.metrics.AverageBatchTime = (bp.metrics.AverageBatchTime + batch.Duration) / 2
	bp.mu.Unlock()
}

// Connection Pool methods

// GetClient returns an HTTP client from the pool
func (cp *ConnectionPool) GetClient() *http.Client {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	client := cp.clients[cp.currentIndex]
	cp.currentIndex = (cp.currentIndex + 1) % cp.poolSize
	return client
}

// Payload Compressor methods

// Compress compresses a payload
func (pc *PayloadCompressor) Compress(payload map[string]interface{}) (map[string]interface{}, error) {
	if !pc.compressionEnabled {
		return payload, nil
	}

	pc.metrics.TotalRequests++

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return payload, err
	}

	// Compress
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, pc.compressionLevel)
	if err != nil {
		return payload, err
	}

	if _, err := writer.Write(jsonData); err != nil {
		return payload, err
	}
	writer.Close()

	// Calculate compression metrics
	originalSize := int64(len(jsonData))
	compressedSize := int64(buf.Len())

	pc.metrics.CompressedRequests++
	pc.metrics.BytesSaved += originalSize - compressedSize
	pc.metrics.CompressionRatio = float64(compressedSize) / float64(originalSize)

	// Return compressed payload metadata
	return map[string]interface{}{
		"compressed":     true,
		"originalSize":   originalSize,
		"compressedSize": compressedSize,
		"data":           buf.Bytes(),
	}, nil
}

// Workflow Cache methods

// Get retrieves a cached workflow result
func (wc *WorkflowCache) Get(key string) *WorkflowResult {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	entry, exists := wc.entries[key]
	if !exists {
		wc.missCount++
		return nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		wc.mu.RUnlock()
		wc.mu.Lock()
		delete(wc.entries, key)
		wc.mu.Unlock()
		wc.mu.RLock()
		wc.missCount++
		return nil
	}

	entry.AccessCount++
	entry.LastAccess = time.Now()
	wc.hitCount++

	return entry.Result
}

// Set stores a workflow result in cache
func (wc *WorkflowCache) Set(key string, result *WorkflowResult) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	// Check cache size and evict if necessary
	if len(wc.entries) >= wc.maxSize {
		wc.evictLRU()
	}

	entry := &WorkflowCacheEntry{
		Key:         key,
		Result:      result,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(wc.ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	wc.entries[key] = entry
}

// evictLRU evicts the least recently used entry
func (wc *WorkflowCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range wc.entries {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(wc.entries, oldestKey)
	}
}

// startCleanup starts cache cleanup goroutine
func (wc *WorkflowCache) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		wc.mu.Lock()
		now := time.Now()

		for key, entry := range wc.entries {
			if now.After(entry.ExpiresAt) {
				delete(wc.entries, key)
			}
		}
		wc.mu.Unlock()
	}
}

// Load Balancer methods

// GetNextEndpoint returns the next endpoint using the configured balancing method
func (lb *LoadBalancer) GetNextEndpoint() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Filter healthy endpoints
	healthyEndpoints := make([]string, 0)
	for endpoint, health := range lb.healthChecks {
		if health.Status == "healthy" {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	if len(healthyEndpoints) == 0 {
		return ""
	}

	// Round robin selection
	endpoint := healthyEndpoints[lb.currentIndex%len(healthyEndpoints)]
	lb.currentIndex++

	return endpoint
}

// startHealthChecks starts health check monitoring
func (lb *LoadBalancer) startHealthChecks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, endpoint := range lb.endpoints {
			go lb.checkEndpointHealth(endpoint)
		}
	}
}

// checkEndpointHealth checks the health of a specific endpoint
func (lb *LoadBalancer) checkEndpointHealth(endpoint string) {
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()

	resp, err := client.Get(endpoint + "/health")
	duration := time.Since(start)

	lb.mu.Lock()
	defer lb.mu.Unlock()

	health := lb.healthChecks[endpoint]
	health.LastHealthCheck = time.Now()
	health.ResponseTime = duration

	if err != nil || resp.StatusCode >= 400 {
		health.Status = "unhealthy"
		health.ErrorRate = 1.0
	} else {
		health.Status = "healthy"
		health.ErrorRate = 0.0
	}

	if resp != nil {
		resp.Body.Close()
	}
}

// Performance Monitor methods

// collectMetrics collects performance metrics
func (wpm *WorkflowPerformanceMonitor) collectMetrics(metrics *WorkflowPerformanceMetrics) {
	if !wpm.enabled {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	// Update internal metrics
	wpm.metrics = metrics
}

// detectBottlenecks detects performance bottlenecks
func (wpm *WorkflowPerformanceMonitor) detectBottlenecks(metrics *WorkflowPerformanceMetrics) {
	if !wpm.enabled {
		return
	}

	wpm.bottleneckDetector.detectBottlenecks(metrics)
}

// checkAlerts checks for performance alerts
func (wpm *WorkflowPerformanceMonitor) checkAlerts(metrics *WorkflowPerformanceMetrics) {
	if !wpm.enabled {
		return
	}

	wpm.alertManager.checkAlerts(metrics)
}

// Bottleneck Detector methods

// detectBottlenecks detects performance bottlenecks
func (bd *BottleneckDetector) detectBottlenecks(metrics *WorkflowPerformanceMetrics) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	// Check execution time bottleneck
	if metrics.AverageExecutionTime.Seconds() > bd.thresholds["execution_time"] {
		bottleneck := &Bottleneck{
			Type:        "execution_time",
			Description: "Average execution time exceeds threshold",
			Severity:    "high",
			Impact:      metrics.AverageExecutionTime.Seconds() / bd.thresholds["execution_time"],
			Suggestions: []string{
				"Optimize workflow nodes",
				"Increase parallel processing",
				"Review resource allocation",
			},
			DetectedAt: time.Now(),
			Metadata:   make(map[string]interface{}),
		}
		bd.detectedBottlenecks = append(bd.detectedBottlenecks, bottleneck)
	}

	// Check error rate bottleneck
	if metrics.ErrorRate > bd.thresholds["error_rate"] {
		bottleneck := &Bottleneck{
			Type:        "error_rate",
			Description: "Error rate exceeds threshold",
			Severity:    "critical",
			Impact:      metrics.ErrorRate / bd.thresholds["error_rate"],
			Suggestions: []string{
				"Review error handling",
				"Check endpoint health",
				"Validate input data",
			},
			DetectedAt: time.Now(),
			Metadata:   make(map[string]interface{}),
		}
		bd.detectedBottlenecks = append(bd.detectedBottlenecks, bottleneck)
	}
}

// Alert Manager methods

// checkAlerts checks for performance alerts
func (am *AlertManager) checkAlerts(metrics *WorkflowPerformanceMetrics) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check error rate alert
	if metrics.ErrorRate > am.thresholds["error_rate"] {
		alert := &PerformanceAlert{
			AlertID:   fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			Type:      "error_rate",
			Severity:  "critical",
			Message:   "Error rate exceeds threshold",
			Value:     metrics.ErrorRate,
			Threshold: am.thresholds["error_rate"],
			CreatedAt: time.Now(),
			Resolved:  false,
			Metadata:  make(map[string]interface{}),
		}
		am.alerts = append(am.alerts, alert)
	}

	// Check response time alert
	if metrics.AverageExecutionTime.Seconds() > am.thresholds["response_time"] {
		alert := &PerformanceAlert{
			AlertID:   fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			Type:      "response_time",
			Severity:  "high",
			Message:   "Average response time exceeds threshold",
			Value:     metrics.AverageExecutionTime.Seconds(),
			Threshold: am.thresholds["response_time"],
			CreatedAt: time.Now(),
			Resolved:  false,
			Metadata:  make(map[string]interface{}),
		}
		am.alerts = append(am.alerts, alert)
	}
}

// GetMetrics returns current workflow performance metrics
func (wpe *WorkflowPerformanceEnhancer) GetMetrics() *WorkflowPerformanceMetrics {
	wpe.mu.RLock()
	defer wpe.mu.RUnlock()

	// Update cache hit rate
	total := wpe.workflowCache.hitCount + wpe.workflowCache.missCount
	if total > 0 {
		wpe.metrics.CacheHitRate = float64(wpe.workflowCache.hitCount) / float64(total)
	}

	// Update compression savings
	if wpe.payloadCompressor.metrics.TotalRequests > 0 {
		wpe.metrics.CompressionSavings = float64(wpe.payloadCompressor.metrics.BytesSaved) / float64(wpe.payloadCompressor.metrics.TotalRequests)
	}

	return wpe.metrics
}

// GetBottlenecks returns detected bottlenecks
func (wpe *WorkflowPerformanceEnhancer) GetBottlenecks() []*Bottleneck {
	return wpe.performanceMonitor.bottleneckDetector.detectedBottlenecks
}

// GetAlerts returns performance alerts
func (wpe *WorkflowPerformanceEnhancer) GetAlerts() []*PerformanceAlert {
	return wpe.performanceMonitor.alertManager.alerts
}
