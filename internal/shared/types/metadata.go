package types

import (
	"time"
)

// Metadata represents common metadata for entities
type Metadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   string                 `json:"createdBy,omitempty"`
	UpdatedBy   string                 `json:"updatedBy,omitempty"`
	Version     int                    `json:"version"`
}

// NewMetadata creates new metadata with ID and name
func NewMetadata(id, name string) *Metadata {
	now := time.Now()
	return &Metadata{
		ID:         id,
		Name:       name,
		Tags:       make([]string, 0),
		Properties: make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// FileMetadata represents metadata for files
type FileMetadata struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Extension   string    `json:"extension,omitempty"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mimeType,omitempty"`
	Checksum    string    `json:"checksum,omitempty"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	AccessedAt  time.Time `json:"accessedAt,omitempty"`
	Permissions string    `json:"permissions,omitempty"`
	IsDirectory bool      `json:"isDirectory"`
	IsHidden    bool      `json:"isHidden"`
	IsReadOnly  bool      `json:"isReadOnly"`
}

// ProcessingMetadata represents metadata for processing operations
type ProcessingMetadata struct {
	JobID          string                 `json:"jobId"`
	ProcessingType string                 `json:"processingType"`
	Status         string                 `json:"status"`
	Priority       int                    `json:"priority"`
	StartTime      time.Time              `json:"startTime"`
	EndTime        *time.Time             `json:"endTime,omitempty"`
	Duration       time.Duration          `json:"duration,omitempty"`
	Progress       float64                `json:"progress"`
	RetryCount     int                    `json:"retryCount"`
	MaxRetries     int                    `json:"maxRetries"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Results        map[string]interface{} `json:"results,omitempty"`
	Errors         []string               `json:"errors,omitempty"`
}

// NewProcessingMetadata creates new processing metadata
func NewProcessingMetadata(jobID, processingType string, priority int) *ProcessingMetadata {
	return &ProcessingMetadata{
		JobID:          jobID,
		ProcessingType: processingType,
		Status:         "pending",
		Priority:       priority,
		StartTime:      time.Now(),
		Progress:       0,
		RetryCount:     0,
		MaxRetries:     3,
		Parameters:     make(map[string]interface{}),
		Results:        make(map[string]interface{}),
		Errors:         make([]string, 0),
	}
}

// DocumentMetadata represents metadata for documents
type DocumentMetadata struct {
	DocumentID     string                 `json:"documentId"`
	DocumentType   string                 `json:"documentType"`
	Category       string                 `json:"category,omitempty"`
	Classification string                 `json:"classification,omitempty"`
	Confidence     float64                `json:"confidence"`
	Language       string                 `json:"language,omitempty"`
	PageCount      int                    `json:"pageCount,omitempty"`
	WordCount      int                    `json:"wordCount,omitempty"`
	Author         string                 `json:"author,omitempty"`
	Subject        string                 `json:"subject,omitempty"`
	Keywords       []string               `json:"keywords,omitempty"`
	ExtractedData  map[string]interface{} `json:"extractedData,omitempty"`
	ProcessingInfo ProcessingMetadata     `json:"processingInfo,omitempty"`
}

// AnalysisMetadata represents metadata for analysis operations
type AnalysisMetadata struct {
	AnalysisID   string                 `json:"analysisId"`
	AnalysisType string                 `json:"analysisType"`
	DealName     string                 `json:"dealName"`
	Documents    []string               `json:"documents"`
	Templates    []string               `json:"templates,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Results      map[string]interface{} `json:"results,omitempty"`
	Confidence   float64                `json:"confidence"`
	Quality      float64                `json:"quality"`
	StartTime    time.Time              `json:"startTime"`
	EndTime      *time.Time             `json:"endTime,omitempty"`
	Duration     time.Duration          `json:"duration,omitempty"`
}
