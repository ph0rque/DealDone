package monitoring

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// SystemMonitor provides real-time system performance and health monitoring
type SystemMonitor struct {
	mu               sync.RWMutex
	alertManager     *AlertManager
	metricsCollector *MetricsCollector
	healthStatus     SystemHealthStatus
	metrics          map[string]*SystemMetrics
	alerts           []Alert
	config           *MonitoringConfig
	lastUpdate       time.Time
	running          bool
	stopChan         chan bool
	logger           *log.Logger
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	Timestamp          time.Time          `json:"timestamp"`
	CPUUsage           float64            `json:"cpu_usage"`
	MemoryUsage        float64            `json:"memory_usage"`
	DiskUsage          float64            `json:"disk_usage"`
	NetworkIO          NetworkMetrics     `json:"network_io"`
	ResponseTime       time.Duration      `json:"response_time"`
	ThroughputRPS      float64            `json:"throughput_rps"`
	ErrorRate          float64            `json:"error_rate"`
	ActiveConnections  int64              `json:"active_connections"`
	QueueDepth         int64              `json:"queue_depth"`
	AIProviderMetrics  AIProviderMetrics  `json:"ai_provider_metrics"`
	N8NWorkflowMetrics N8NMetrics         `json:"n8n_workflow_metrics"`
	DatabaseMetrics    DatabaseMetrics    `json:"database_metrics"`
	CustomMetrics      map[string]float64 `json:"custom_metrics"`
}

// NetworkMetrics tracks network performance
type NetworkMetrics struct {
	BytesIn       int64   `json:"bytes_in"`
	BytesOut      int64   `json:"bytes_out"`
	PacketsIn     int64   `json:"packets_in"`
	PacketsOut    int64   `json:"packets_out"`
	ConnectionsIn int64   `json:"connections_in"`
	Latency       float64 `json:"latency"`
}

// AIProviderMetrics tracks AI provider performance
type AIProviderMetrics struct {
	OpenAIResponseTime time.Duration `json:"openai_response_time"`
	ClaudeResponseTime time.Duration `json:"claude_response_time"`
	OpenAIErrorRate    float64       `json:"openai_error_rate"`
	ClaudeErrorRate    float64       `json:"claude_error_rate"`
	TokensUsed         int64         `json:"tokens_used"`
	RequestsPerMinute  int64         `json:"requests_per_minute"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	CostPerHour        float64       `json:"cost_per_hour"`
}

// N8NMetrics tracks n8n workflow performance
type N8NMetrics struct {
	WorkflowExecutions   int64         `json:"workflow_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	QueuedExecutions     int64         `json:"queued_executions"`
	WorkflowUptime       float64       `json:"workflow_uptime"`
}

// DatabaseMetrics tracks database performance
type DatabaseMetrics struct {
	ConnectionCount    int64         `json:"connection_count"`
	QueryResponseTime  time.Duration `json:"query_response_time"`
	DeadlockCount      int64         `json:"deadlock_count"`
	CacheHitRatio      float64       `json:"cache_hit_ratio"`
	TransactionsPerSec float64       `json:"transactions_per_sec"`
	DatabaseSize       int64         `json:"database_size"`
}

// Alert represents a system alert
type Alert struct {
	ID           string        `json:"id"`
	Type         AlertType     `json:"type"`
	Severity     AlertSeverity `json:"severity"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Source       string        `json:"source"`
	Metric       string        `json:"metric"`
	Threshold    float64       `json:"threshold"`
	CurrentValue float64       `json:"current_value"`
	Timestamp    time.Time     `json:"timestamp"`
	Resolved     bool          `json:"resolved"`
	ResolvedAt   time.Time     `json:"resolved_at"`
	Actions      []AlertAction `json:"actions"`
}

// AlertAction represents an action to take for an alert
type AlertAction struct {
	Type        ActionType `json:"type"`
	Description string     `json:"description"`
	Executed    bool       `json:"executed"`
	ExecutedAt  time.Time  `json:"executed_at"`
	Result      string     `json:"result"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	UpdateInterval     time.Duration       `json:"update_interval"`
	RetentionPeriod    time.Duration       `json:"retention_period"`
	AlertThresholds    map[string]float64  `json:"alert_thresholds"`
	NotificationConfig *NotificationConfig `json:"notification_config"`
	MetricsEndpoints   []string            `json:"metrics_endpoints"`
	HealthCheckConfig  *HealthCheckConfig  `json:"health_check_config"`
	DashboardConfig    *DashboardConfig    `json:"dashboard_config"`
}

// NotificationConfig configures alert notifications
type NotificationConfig struct {
	SlackWebhook    string   `json:"slack_webhook"`
	EmailRecipients []string `json:"email_recipients"`
	SMSRecipients   []string `json:"sms_recipients"`
	EnabledChannels []string `json:"enabled_channels"`
}

// HealthCheckConfig configures health checking
type HealthCheckConfig struct {
	Endpoints []HealthEndpoint `json:"endpoints"`
	Interval  time.Duration    `json:"interval"`
	Timeout   time.Duration    `json:"timeout"`
	Retries   int              `json:"retries"`
}

// HealthEndpoint represents a health check endpoint
type HealthEndpoint struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Method   string `json:"method"`
	Expected int    `json:"expected"`
	Critical bool   `json:"critical"`
}

// DashboardConfig configures monitoring dashboards
type DashboardConfig struct {
	ExecutiveDashboard   *DashboardSettings `json:"executive_dashboard"`
	OperationalDashboard *DashboardSettings `json:"operational_dashboard"`
	TechnicalDashboard   *DashboardSettings `json:"technical_dashboard"`
	RefreshInterval      time.Duration      `json:"refresh_interval"`
}

// DashboardSettings represents dashboard configuration
type DashboardSettings struct {
	Enabled     bool     `json:"enabled"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Widgets     []string `json:"widgets"`
	Layout      string   `json:"layout"`
}

// Enums
type SystemHealthStatus string

const (
	HealthStatusHealthy   SystemHealthStatus = "healthy"
	HealthStatusDegraded  SystemHealthStatus = "degraded"
	HealthStatusUnhealthy SystemHealthStatus = "unhealthy"
	HealthStatusUnknown   SystemHealthStatus = "unknown"
)

type AlertType string

const (
	AlertTypePerformance AlertType = "performance"
	AlertTypeError       AlertType = "error"
	AlertTypeHealth      AlertType = "health"
	AlertTypeResource    AlertType = "resource"
	AlertTypeSecurity    AlertType = "security"
	AlertTypeCustom      AlertType = "custom"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

type ActionType string

const (
	ActionTypeNotify      ActionType = "notify"
	ActionTypeRestart     ActionType = "restart"
	ActionTypeScale       ActionType = "scale"
	ActionTypeRollback    ActionType = "rollback"
	ActionTypeInvestigate ActionType = "investigate"
)

// NewSystemMonitor creates a new system monitor instance
func NewSystemMonitor(logger *log.Logger) *SystemMonitor {
	return &SystemMonitor{
		alertManager:     NewAlertManager(),
		metricsCollector: NewMetricsCollector(),
		healthStatus:     HealthStatusUnknown,
		metrics:          make(map[string]*SystemMetrics),
		alerts:           make([]Alert, 0),
		config: &MonitoringConfig{
			UpdateInterval:  30 * time.Second,
			RetentionPeriod: 7 * 24 * time.Hour,
			AlertThresholds: map[string]float64{
				"cpu_usage":     80.0,
				"memory_usage":  85.0,
				"disk_usage":    90.0,
				"error_rate":    5.0,
				"response_time": 5000.0,
				"queue_depth":   100.0,
			},
			NotificationConfig: &NotificationConfig{
				EnabledChannels: []string{"slack", "email"},
			},
			HealthCheckConfig: &HealthCheckConfig{
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Retries:  3,
			},
			DashboardConfig: &DashboardConfig{
				ExecutiveDashboard: &DashboardSettings{
					Enabled: true,
					Title:   "Executive Dashboard",
					Widgets: []string{"system_health", "key_metrics", "alerts_summary"},
				},
				OperationalDashboard: &DashboardSettings{
					Enabled: true,
					Title:   "Operational Dashboard",
					Widgets: []string{"detailed_metrics", "performance_charts", "alert_details"},
				},
				RefreshInterval: 5 * time.Second,
			},
		},
		stopChan:   make(chan bool),
		logger:     logger,
		lastUpdate: time.Now(),
	}
}

// Start begins monitoring
func (sm *SystemMonitor) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("system monitor is already running")
	}

	sm.running = true
	sm.logger.Printf("Starting system monitor with %d second intervals", int(sm.config.UpdateInterval.Seconds()))

	// Start monitoring goroutines
	go sm.monitoringLoop(ctx)
	go sm.healthCheckLoop(ctx)
	go sm.alertProcessingLoop(ctx)

	return nil
}

// Stop stops monitoring
func (sm *SystemMonitor) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return fmt.Errorf("system monitor is not running")
	}

	sm.running = false
	close(sm.stopChan)
	sm.logger.Printf("System monitor stopped")

	return nil
}

// monitoringLoop runs the main monitoring loop
func (sm *SystemMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(sm.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.collectMetrics()
		}
	}
}

// collectMetrics collects system metrics
func (sm *SystemMonitor) collectMetrics() {
	timestamp := time.Now()

	// Simulate metric collection (in production, use actual system calls)
	metrics := &SystemMetrics{
		Timestamp:         timestamp,
		CPUUsage:          sm.simulateCPUUsage(),
		MemoryUsage:       sm.simulateMemoryUsage(),
		DiskUsage:         sm.simulateDiskUsage(),
		ResponseTime:      sm.simulateResponseTime(),
		ThroughputRPS:     sm.simulateThroughput(),
		ErrorRate:         sm.simulateErrorRate(),
		ActiveConnections: sm.simulateActiveConnections(),
		QueueDepth:        sm.simulateQueueDepth(),
		NetworkIO: NetworkMetrics{
			BytesIn:  int64(1024 * 1024 * (50 + rand.Intn(50))),
			BytesOut: int64(1024 * 1024 * (30 + rand.Intn(30))),
			Latency:  float64(10 + rand.Intn(40)),
		},
		AIProviderMetrics: AIProviderMetrics{
			OpenAIResponseTime: time.Duration(500+rand.Intn(1000)) * time.Millisecond,
			ClaudeResponseTime: time.Duration(600+rand.Intn(1200)) * time.Millisecond,
			OpenAIErrorRate:    float64(rand.Intn(5)),
			ClaudeErrorRate:    float64(rand.Intn(3)),
			TokensUsed:         int64(1000 + rand.Intn(2000)),
			RequestsPerMinute:  int64(50 + rand.Intn(100)),
			CacheHitRate:       float64(70 + rand.Intn(25)),
			CostPerHour:        float64(5 + rand.Intn(10)),
		},
		N8NWorkflowMetrics: N8NMetrics{
			WorkflowExecutions:   int64(100 + rand.Intn(50)),
			SuccessfulExecutions: int64(95 + rand.Intn(5)),
			FailedExecutions:     int64(rand.Intn(5)),
			AverageExecutionTime: time.Duration(2000+rand.Intn(3000)) * time.Millisecond,
			QueuedExecutions:     int64(rand.Intn(10)),
			WorkflowUptime:       float64(95 + rand.Intn(5)),
		},
		DatabaseMetrics: DatabaseMetrics{
			ConnectionCount:    int64(10 + rand.Intn(40)),
			QueryResponseTime:  time.Duration(50+rand.Intn(200)) * time.Millisecond,
			DeadlockCount:      int64(rand.Intn(2)),
			CacheHitRatio:      float64(80 + rand.Intn(15)),
			TransactionsPerSec: float64(100 + rand.Intn(200)),
			DatabaseSize:       int64(1024 * 1024 * 1024 * (5 + rand.Intn(5))),
		},
		CustomMetrics: make(map[string]float64),
	}

	// Store metrics
	sm.mu.Lock()
	sm.metrics[timestamp.Format(time.RFC3339)] = metrics
	sm.lastUpdate = timestamp

	// Clean up old metrics
	sm.cleanupOldMetrics()
	sm.mu.Unlock()

	// Check for alerts
	sm.checkAlertConditions(metrics)

	// Update health status
	sm.updateHealthStatus(metrics)

	sm.logger.Printf("Collected metrics: CPU=%.1f%%, Memory=%.1f%%, Disk=%.1f%%, ErrorRate=%.1f%%",
		metrics.CPUUsage, metrics.MemoryUsage, metrics.DiskUsage, metrics.ErrorRate)
}

// Metric simulation methods (replace with actual system calls in production)
func (sm *SystemMonitor) simulateCPUUsage() float64 {
	return float64(30 + rand.Intn(40)) // 30-70%
}

func (sm *SystemMonitor) simulateMemoryUsage() float64 {
	return float64(40 + rand.Intn(35)) // 40-75%
}

func (sm *SystemMonitor) simulateDiskUsage() float64 {
	return float64(20 + rand.Intn(50)) // 20-70%
}

func (sm *SystemMonitor) simulateResponseTime() time.Duration {
	return time.Duration(100+rand.Intn(400)) * time.Millisecond // 100-500ms
}

func (sm *SystemMonitor) simulateThroughput() float64 {
	return float64(50 + rand.Intn(100)) // 50-150 RPS
}

func (sm *SystemMonitor) simulateErrorRate() float64 {
	return float64(rand.Intn(8)) // 0-8%
}

func (sm *SystemMonitor) simulateActiveConnections() int64 {
	return int64(50 + rand.Intn(200)) // 50-250 connections
}

func (sm *SystemMonitor) simulateQueueDepth() int64 {
	return int64(rand.Intn(50)) // 0-50 items
}

// checkAlertConditions checks if any alert conditions are met
func (sm *SystemMonitor) checkAlertConditions(metrics *SystemMetrics) {
	thresholds := sm.config.AlertThresholds

	// Check CPU usage
	if metrics.CPUUsage > thresholds["cpu_usage"] {
		sm.createAlert("high_cpu_usage", AlertTypePerformance, AlertSeverityHigh,
			"High CPU Usage", fmt.Sprintf("CPU usage is %.1f%%, exceeding threshold of %.1f%%",
				metrics.CPUUsage, thresholds["cpu_usage"]), "cpu_usage",
			thresholds["cpu_usage"], metrics.CPUUsage)
	}

	// Check memory usage
	if metrics.MemoryUsage > thresholds["memory_usage"] {
		sm.createAlert("high_memory_usage", AlertTypeResource, AlertSeverityHigh,
			"High Memory Usage", fmt.Sprintf("Memory usage is %.1f%%, exceeding threshold of %.1f%%",
				metrics.MemoryUsage, thresholds["memory_usage"]), "memory_usage",
			thresholds["memory_usage"], metrics.MemoryUsage)
	}

	// Check error rate
	if metrics.ErrorRate > thresholds["error_rate"] {
		sm.createAlert("high_error_rate", AlertTypeError, AlertSeverityCritical,
			"High Error Rate", fmt.Sprintf("Error rate is %.1f%%, exceeding threshold of %.1f%%",
				metrics.ErrorRate, thresholds["error_rate"]), "error_rate",
			thresholds["error_rate"], metrics.ErrorRate)
	}

	// Check response time
	responseTimeMs := float64(metrics.ResponseTime.Milliseconds())
	if responseTimeMs > thresholds["response_time"] {
		sm.createAlert("high_response_time", AlertTypePerformance, AlertSeverityMedium,
			"High Response Time", fmt.Sprintf("Response time is %.0fms, exceeding threshold of %.0fms",
				responseTimeMs, thresholds["response_time"]), "response_time",
			thresholds["response_time"], responseTimeMs)
	}
}

// createAlert creates a new alert
func (sm *SystemMonitor) createAlert(id string, alertType AlertType, severity AlertSeverity,
	title, description, metric string, threshold, currentValue float64) {

	// Check if alert already exists and is unresolved
	for _, alert := range sm.alerts {
		if alert.ID == id && !alert.Resolved {
			return // Alert already active
		}
	}

	alert := Alert{
		ID:           id,
		Type:         alertType,
		Severity:     severity,
		Title:        title,
		Description:  description,
		Source:       "system_monitor",
		Metric:       metric,
		Threshold:    threshold,
		CurrentValue: currentValue,
		Timestamp:    time.Now(),
		Resolved:     false,
		Actions:      sm.getActionsForAlert(alertType, severity),
	}

	sm.mu.Lock()
	sm.alerts = append(sm.alerts, alert)
	sm.mu.Unlock()

	// Send alert notification
	sm.alertManager.ProcessAlert(&alert)

	sm.logger.Printf("ALERT: %s - %s", alert.Title, alert.Description)
}

// getActionsForAlert returns recommended actions for an alert
func (sm *SystemMonitor) getActionsForAlert(alertType AlertType, severity AlertSeverity) []AlertAction {
	actions := []AlertAction{}

	switch alertType {
	case AlertTypePerformance:
		actions = append(actions, AlertAction{
			Type:        ActionTypeNotify,
			Description: "Notify operations team",
		})
		if severity == AlertSeverityCritical {
			actions = append(actions, AlertAction{
				Type:        ActionTypeScale,
				Description: "Consider scaling resources",
			})
		}
	case AlertTypeError:
		actions = append(actions, AlertAction{
			Type:        ActionTypeNotify,
			Description: "Notify development team",
		})
		actions = append(actions, AlertAction{
			Type:        ActionTypeInvestigate,
			Description: "Investigate error logs",
		})
	case AlertTypeResource:
		actions = append(actions, AlertAction{
			Type:        ActionTypeNotify,
			Description: "Notify infrastructure team",
		})
		if severity >= AlertSeverityHigh {
			actions = append(actions, AlertAction{
				Type:        ActionTypeScale,
				Description: "Scale infrastructure resources",
			})
		}
	}

	return actions
}

// updateHealthStatus updates overall system health status
func (sm *SystemMonitor) updateHealthStatus(metrics *SystemMetrics) {
	criticalIssues := 0
	warningIssues := 0

	// Count unresolved alerts by severity
	for _, alert := range sm.alerts {
		if !alert.Resolved {
			switch alert.Severity {
			case AlertSeverityCritical:
				criticalIssues++
			case AlertSeverityHigh, AlertSeverityMedium:
				warningIssues++
			}
		}
	}

	// Determine health status
	var newStatus SystemHealthStatus
	if criticalIssues > 0 {
		newStatus = HealthStatusUnhealthy
	} else if warningIssues > 0 {
		newStatus = HealthStatusDegraded
	} else {
		newStatus = HealthStatusHealthy
	}

	sm.mu.Lock()
	if sm.healthStatus != newStatus {
		sm.logger.Printf("System health status changed from %s to %s", sm.healthStatus, newStatus)
		sm.healthStatus = newStatus
	}
	sm.mu.Unlock()
}

// cleanupOldMetrics removes old metrics based on retention period
func (sm *SystemMonitor) cleanupOldMetrics() {
	cutoff := time.Now().Add(-sm.config.RetentionPeriod)

	for timestamp := range sm.metrics {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil && t.Before(cutoff) {
			delete(sm.metrics, timestamp)
		}
	}
}

// healthCheckLoop runs health checks
func (sm *SystemMonitor) healthCheckLoop(ctx context.Context) {
	if sm.config.HealthCheckConfig == nil {
		return
	}

	ticker := time.NewTicker(sm.config.HealthCheckConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on configured endpoints
func (sm *SystemMonitor) performHealthChecks() {
	if sm.config.HealthCheckConfig == nil || len(sm.config.HealthCheckConfig.Endpoints) == 0 {
		return
	}

	for _, endpoint := range sm.config.HealthCheckConfig.Endpoints {
		healthy := sm.checkEndpointHealth(endpoint)

		if !healthy && endpoint.Critical {
			sm.createAlert(fmt.Sprintf("health_check_%s", endpoint.Name),
				AlertTypeHealth, AlertSeverityCritical,
				"Critical Health Check Failed",
				fmt.Sprintf("Health check for %s (%s) failed", endpoint.Name, endpoint.URL),
				"health_check", 1.0, 0.0)
		}
	}
}

// checkEndpointHealth checks the health of a single endpoint
func (sm *SystemMonitor) checkEndpointHealth(endpoint HealthEndpoint) bool {
	// Simulate health check (in production, make actual HTTP request)
	// For now, randomly return true/false with bias towards healthy
	return rand.Float64() > 0.1 // 90% healthy
}

// alertProcessingLoop processes alerts
func (sm *SystemMonitor) alertProcessingLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.processAlerts()
		}
	}
}

// processAlerts processes and resolves alerts
func (sm *SystemMonitor) processAlerts() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Auto-resolve alerts that are no longer triggered
	for i := range sm.alerts {
		if !sm.alerts[i].Resolved && time.Since(sm.alerts[i].Timestamp) > 10*time.Minute {
			// Check if condition is still met
			if !sm.isAlertConditionStillMet(&sm.alerts[i]) {
				sm.alerts[i].Resolved = true
				sm.alerts[i].ResolvedAt = time.Now()
				sm.logger.Printf("Auto-resolved alert: %s", sm.alerts[i].Title)
			}
		}
	}
}

// isAlertConditionStillMet checks if an alert condition is still met
func (sm *SystemMonitor) isAlertConditionStillMet(alert *Alert) bool {
	// Get latest metrics
	var latestMetrics *SystemMetrics
	latestTime := time.Time{}

	for timestamp, metrics := range sm.metrics {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil && t.After(latestTime) {
			latestTime = t
			latestMetrics = metrics
		}
	}

	if latestMetrics == nil {
		return false
	}

	// Check if current value still exceeds threshold
	switch alert.Metric {
	case "cpu_usage":
		return latestMetrics.CPUUsage > alert.Threshold
	case "memory_usage":
		return latestMetrics.MemoryUsage > alert.Threshold
	case "error_rate":
		return latestMetrics.ErrorRate > alert.Threshold
	case "response_time":
		return float64(latestMetrics.ResponseTime.Milliseconds()) > alert.Threshold
	}

	return false
}

// GetCurrentMetrics returns the latest system metrics
func (sm *SystemMonitor) GetCurrentMetrics() (*SystemMetrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.metrics) == 0 {
		return nil, fmt.Errorf("no metrics available")
	}

	// Find latest metrics
	var latestMetrics *SystemMetrics
	latestTime := time.Time{}

	for timestamp, metrics := range sm.metrics {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil && t.After(latestTime) {
			latestTime = t
			latestMetrics = metrics
		}
	}

	return latestMetrics, nil
}

// GetHealthStatus returns the current system health status
func (sm *SystemMonitor) GetHealthStatus() SystemHealthStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.healthStatus
}

// GetActiveAlerts returns all active (unresolved) alerts
func (sm *SystemMonitor) GetActiveAlerts() []Alert {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	activeAlerts := make([]Alert, 0)
	for _, alert := range sm.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetMetricsHistory returns metrics within a time range
func (sm *SystemMonitor) GetMetricsHistory(start, end time.Time) []*SystemMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	history := make([]*SystemMetrics, 0)
	for timestamp, metrics := range sm.metrics {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			if t.After(start) && t.Before(end) {
				history = append(history, metrics)
			}
		}
	}

	return history
}

// Supporting component constructors
func NewAlertManager() *AlertManager {
	return &AlertManager{
		notifications: make([]AlertNotification, 0),
		config:        &AlertManagerConfig{},
	}
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		collectors: make(map[string]MetricCollector),
		config:     &MetricsConfig{},
	}
}

// Placeholder types for supporting components
type AlertManager struct {
	notifications []AlertNotification
	config        *AlertManagerConfig
}

type AlertNotification struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	Channel   string    `json:"channel"`
	Sent      bool      `json:"sent"`
	Timestamp time.Time `json:"timestamp"`
}

type AlertManagerConfig struct {
	NotificationTimeout time.Duration `json:"notification_timeout"`
	RetryAttempts       int           `json:"retry_attempts"`
}

type MetricsCollector struct {
	collectors map[string]MetricCollector
	config     *MetricsConfig
}

type MetricCollector interface {
	Collect() (map[string]float64, error)
}

type MetricsConfig struct {
	CollectionInterval time.Duration `json:"collection_interval"`
	BufferSize         int           `json:"buffer_size"`
}

// ProcessAlert processes an alert (placeholder implementation)
func (am *AlertManager) ProcessAlert(alert *Alert) {
	// In a real implementation, this would send notifications via configured channels
	notification := AlertNotification{
		ID:        fmt.Sprintf("notif_%d", time.Now().Unix()),
		AlertID:   alert.ID,
		Channel:   "slack",
		Sent:      true,
		Timestamp: time.Now(),
	}

	am.notifications = append(am.notifications, notification)
}
