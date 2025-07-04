package ocr

import (
	"time"
)

// Result represents the result of OCR processing
type Result struct {
	Text           string                 `json:"text"`
	Confidence     float64                `json:"confidence"`
	Language       string                 `json:"language"`
	Pages          []PageResult           `json:"pages"`
	ProcessingTime time.Duration          `json:"processingTime"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PageResult represents OCR results for a single page
type PageResult struct {
	PageNumber int           `json:"pageNumber"`
	Text       string        `json:"text"`
	Confidence float64       `json:"confidence"`
	Width      int           `json:"width"`
	Height     int           `json:"height"`
	Blocks     []TextBlock   `json:"blocks"`
	Tables     []TableResult `json:"tables,omitempty"`
}

// TextBlock represents a block of text detected by OCR
type TextBlock struct {
	Text        string      `json:"text"`
	Confidence  float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"boundingBox"`
	Words       []Word      `json:"words,omitempty"`
	BlockType   string      `json:"blockType"` // "text", "title", "header", "footer"
}

// Word represents a single word detected by OCR
type Word struct {
	Text        string      `json:"text"`
	Confidence  float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"boundingBox"`
}

// BoundingBox represents the location of text in the document
type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TableResult represents a table detected by OCR
type TableResult struct {
	Rows        int         `json:"rows"`
	Columns     int         `json:"columns"`
	Cells       [][]Cell    `json:"cells"`
	BoundingBox BoundingBox `json:"boundingBox"`
	Confidence  float64     `json:"confidence"`
}

// Cell represents a cell in a detected table
type Cell struct {
	Text        string      `json:"text"`
	RowIndex    int         `json:"rowIndex"`
	ColumnIndex int         `json:"columnIndex"`
	RowSpan     int         `json:"rowSpan"`
	ColumnSpan  int         `json:"columnSpan"`
	BoundingBox BoundingBox `json:"boundingBox"`
	Confidence  float64     `json:"confidence"`
}

// Options represents configuration options for OCR processing
type Options struct {
	Language          string        `json:"language"`          // Language hint for OCR
	EnableTables      bool          `json:"enableTables"`      // Enable table detection
	EnableHandwriting bool          `json:"enableHandwriting"` // Enable handwriting recognition
	PageNumbers       []int         `json:"pageNumbers"`       // Specific pages to process
	DPI               int           `json:"dpi"`               // DPI for image conversion
	PreprocessImage   bool          `json:"preprocessImage"`   // Enable image preprocessing
	OutputFormat      string        `json:"outputFormat"`      // "text", "json", "hocr"
	Timeout           time.Duration `json:"timeout"`           // Processing timeout
}

// Provider represents an OCR service provider
type Provider string

const (
	ProviderTesseract Provider = "tesseract"
	ProviderGoogle    Provider = "google"
	ProviderAWS       Provider = "aws"
	ProviderAzure     Provider = "azure"
	ProviderLocal     Provider = "local"
)

// Status represents the status of an OCR job
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// Job represents an OCR processing job
type Job struct {
	ID          string                 `json:"id"`
	FilePath    string                 `json:"filePath"`
	Provider    Provider               `json:"provider"`
	Status      Status                 `json:"status"`
	Progress    float64                `json:"progress"`
	Result      *Result                `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Options     Options                `json:"options"`
	CreatedAt   time.Time              `json:"createdAt"`
	StartedAt   *time.Time             `json:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderConfig represents configuration for an OCR provider
type ProviderConfig struct {
	Name        Provider               `json:"name"`
	APIKey      string                 `json:"apiKey,omitempty"`
	APIEndpoint string                 `json:"apiEndpoint,omitempty"`
	MaxFileSize int64                  `json:"maxFileSize"`
	Timeout     time.Duration          `json:"timeout"`
	Features    []string               `json:"features"`
	Languages   []string               `json:"languages"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
