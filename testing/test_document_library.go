package testing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestDocumentLibrary manages test documents for comprehensive workflow testing
type TestDocumentLibrary struct {
	LibraryPath     string
	DocumentSets    map[string]*TestDocumentSet
	SyntheticDocs   map[string]*SyntheticDocument
	TestScenarios   map[string]*TestScenario
	PerformanceData map[string]*PerformanceMetrics
}

// TestDocumentSet represents a collection of related test documents
type TestDocumentSet struct {
	SetID          string                 `json:"setId"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Industry       string                 `json:"industry"`
	DealType       string                 `json:"dealType"`
	DealSize       string                 `json:"dealSize"`
	Documents      []TestDocument         `json:"documents"`
	ExpectedData   map[string]interface{} `json:"expectedData"`
	Complexity     string                 `json:"complexity"` // "simple", "medium", "complex"
	CreatedAt      time.Time              `json:"createdAt"`
	LastUpdated    time.Time              `json:"lastUpdated"`
}

// TestDocument represents a single test document
type TestDocument struct {
	DocumentID    string                 `json:"documentId"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"` // "cim", "financial", "legal", "pitch_deck", etc.
	FilePath      string                 `json:"filePath"`
	Size          int64                  `json:"size"`
	Pages         int                    `json:"pages"`
	Quality       string                 `json:"quality"` // "high", "medium", "low", "corrupted"
	Language      string                 `json:"language"`
	ExpectedData  map[string]interface{} `json:"expectedData"`
	TestTags      []string               `json:"testTags"`
	CreatedAt     time.Time              `json:"createdAt"`
}

// SyntheticDocument represents artificially generated test documents
type SyntheticDocument struct {
	DocumentID      string                 `json:"documentId"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Content         string                 `json:"content"`
	GenerationRules map[string]interface{} `json:"generationRules"`
	ExpectedData    map[string]interface{} `json:"expectedData"`
	TestPurpose     string                 `json:"testPurpose"`
	Complexity      string                 `json:"complexity"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// TestScenario represents a specific testing scenario
type TestScenario struct {
	ScenarioID      string                 `json:"scenarioId"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"` // "happy_path", "error_case", "edge_case", "performance"
	DocumentSetID   string                 `json:"documentSetId"`
	ExpectedResults map[string]interface{} `json:"expectedResults"`
	TestSteps       []TestStep             `json:"testSteps"`
	PassCriteria    []string               `json:"passCriteria"`
	FailureModes    []string               `json:"failureModes"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	StepID          string                 `json:"stepId"`
	Name            string                 `json:"name"`
	Action          string                 `json:"action"`
	Parameters      map[string]interface{} `json:"parameters"`
	ExpectedOutcome string                 `json:"expectedOutcome"`
	Timeout         time.Duration          `json:"timeout"`
	RetryCount      int                    `json:"retryCount"`
}

// PerformanceMetrics tracks performance data for test scenarios
type PerformanceMetrics struct {
	ScenarioID        string        `json:"scenarioId"`
	ExecutionTime     time.Duration `json:"executionTime"`
	MemoryUsage       int64         `json:"memoryUsage"`
	AIAPICalls        int           `json:"aiApiCalls"`
	DocumentsProcessed int           `json:"documentsProcessed"`
	SuccessRate       float64       `json:"successRate"`
	ErrorCount        int           `json:"errorCount"`
	Timestamp         time.Time     `json:"timestamp"`
}

// NewTestDocumentLibrary creates a new test document library
func NewTestDocumentLibrary(libraryPath string) *TestDocumentLibrary {
	return &TestDocumentLibrary{
		LibraryPath:     libraryPath,
		DocumentSets:    make(map[string]*TestDocumentSet),
		SyntheticDocs:   make(map[string]*SyntheticDocument),
		TestScenarios:   make(map[string]*TestScenario),
		PerformanceData: make(map[string]*PerformanceMetrics),
	}
}

// GetDocumentSet retrieves a document set by ID
func (tdl *TestDocumentLibrary) GetDocumentSet(setID string) (*TestDocumentSet, bool) {
	docSet, exists := tdl.DocumentSets[setID]
	return docSet, exists
}

// GetSyntheticDocument retrieves a synthetic document by ID
func (tdl *TestDocumentLibrary) GetSyntheticDocument(docID string) (*SyntheticDocument, bool) {
	synDoc, exists := tdl.SyntheticDocs[docID]
	return synDoc, exists
}

// GetTestScenario retrieves a test scenario by ID
func (tdl *TestDocumentLibrary) GetTestScenario(scenarioID string) (*TestScenario, bool) {
	scenario, exists := tdl.TestScenarios[scenarioID]
	return scenario, exists
}

// ListDocumentSets returns all available document sets
func (tdl *TestDocumentLibrary) ListDocumentSets() []string {
	sets := make([]string, 0, len(tdl.DocumentSets))
	for setID := range tdl.DocumentSets {
		sets = append(sets, setID)
	}
	return sets
}

// ListTestScenarios returns all available test scenarios
func (tdl *TestDocumentLibrary) ListTestScenarios() []string {
	scenarios := make([]string, 0, len(tdl.TestScenarios))
	for scenarioID := range tdl.TestScenarios {
		scenarios = append(scenarios, scenarioID)
	}
	return scenarios
}

// RecordPerformanceMetrics records performance data for a test scenario
func (tdl *TestDocumentLibrary) RecordPerformanceMetrics(metrics *PerformanceMetrics) {
	tdl.PerformanceData[metrics.ScenarioID] = metrics
}

// GetPerformanceMetrics retrieves performance metrics for a scenario
func (tdl *TestDocumentLibrary) GetPerformanceMetrics(scenarioID string) (*PerformanceMetrics, bool) {
	metrics, exists := tdl.PerformanceData[scenarioID]
	return metrics, exists
}
