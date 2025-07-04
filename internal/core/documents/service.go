package documents

import (
	"context"
	"path/filepath"
	"strings"
)

// Service provides a unified interface for document operations
type Service interface {
	// Document Processing
	ProcessDocument(ctx context.Context, filePath string) (*DocumentInfo, error)
	ProcessDocuments(ctx context.Context, filePaths []string, dealName string) ([]*RoutingResult, error)
	ProcessFolder(ctx context.Context, folderPath string, dealName string) ([]*RoutingResult, error)

	// Document Routing
	RouteDocument(ctx context.Context, filePath string, dealName string) (*RoutingResult, error)
	RouteDocuments(ctx context.Context, filePaths []string, dealName string) ([]*RoutingResult, error)
	RouteFolder(ctx context.Context, folderPath string, dealName string) ([]*RoutingResult, error)

	// Document Analysis
	AnalyzeDocument(ctx context.Context, filePath string) (*DocumentInfo, error)
	ExtractText(ctx context.Context, filePath string) (string, error)
	GetDocumentMetadata(ctx context.Context, filePath string) (map[string]interface{}, error)

	// AI-Powered Analysis
	AnalyzeDocumentRisks(ctx context.Context, filePath string) (*RiskAnalysis, error)
	GenerateDocumentInsights(ctx context.Context, filePath string) (*DocumentInsights, error)
	ExtractDocumentEntities(ctx context.Context, filePath string) (*EntityExtraction, error)
	ExtractFinancialData(ctx context.Context, filePath string) (*FinancialAnalysis, error)

	// Utility Functions
	GetSupportedFileTypes() []string
	IsSupported(filePath string) bool
	GetDocumentType(filePath string) DocumentType
	GetRoutingSummary(results []*RoutingResult) map[string]interface{}
}

// Manager combines document processing and routing services
type Manager struct {
	processor *DocumentProcessor
	router    *DocumentRouter
}

// NewManager creates a new unified document manager
func NewManager(processor *DocumentProcessor, router *DocumentRouter) *Manager {
	return &Manager{
		processor: processor,
		router:    router,
	}
}

// Document Processing
func (m *Manager) ProcessDocument(ctx context.Context, filePath string) (*DocumentInfo, error) {
	return m.processor.ProcessDocument(filePath)
}

func (m *Manager) ProcessDocuments(ctx context.Context, filePaths []string, dealName string) ([]*RoutingResult, error) {
	return m.router.RouteDocuments(filePaths, dealName)
}

func (m *Manager) ProcessFolder(ctx context.Context, folderPath string, dealName string) ([]*RoutingResult, error) {
	return m.router.RouteFolder(folderPath, dealName)
}

// Document Routing
func (m *Manager) RouteDocument(ctx context.Context, filePath string, dealName string) (*RoutingResult, error) {
	return m.router.RouteDocument(filePath, dealName)
}

func (m *Manager) RouteDocuments(ctx context.Context, filePaths []string, dealName string) ([]*RoutingResult, error) {
	return m.router.RouteDocuments(filePaths, dealName)
}

func (m *Manager) RouteFolder(ctx context.Context, folderPath string, dealName string) ([]*RoutingResult, error) {
	return m.router.RouteFolder(folderPath, dealName)
}

// Document Analysis
func (m *Manager) AnalyzeDocument(ctx context.Context, filePath string) (*DocumentInfo, error) {
	return m.processor.ProcessDocument(filePath)
}

func (m *Manager) ExtractText(ctx context.Context, filePath string) (string, error) {
	return m.processor.ExtractText(filePath)
}

func (m *Manager) GetDocumentMetadata(ctx context.Context, filePath string) (map[string]interface{}, error) {
	return m.processor.GetDocumentMetadata(filePath)
}

// AI-Powered Analysis
func (m *Manager) AnalyzeDocumentRisks(ctx context.Context, filePath string) (*RiskAnalysis, error) {
	// These methods don't exist yet in DocumentProcessor, so we'll return nil for now
	// They would need to be implemented in the actual refactoring
	return nil, nil
}

func (m *Manager) GenerateDocumentInsights(ctx context.Context, filePath string) (*DocumentInsights, error) {
	// These methods don't exist yet in DocumentProcessor, so we'll return nil for now
	// They would need to be implemented in the actual refactoring
	return nil, nil
}

func (m *Manager) ExtractDocumentEntities(ctx context.Context, filePath string) (*EntityExtraction, error) {
	// These methods don't exist yet in DocumentProcessor, so we'll return nil for now
	// They would need to be implemented in the actual refactoring
	return nil, nil
}

func (m *Manager) ExtractFinancialData(ctx context.Context, filePath string) (*FinancialAnalysis, error) {
	// These methods don't exist yet in DocumentProcessor, so we'll return nil for now
	// They would need to be implemented in the actual refactoring
	return nil, nil
}

// Utility Functions
func (m *Manager) GetSupportedFileTypes() []string {
	// Return commonly supported file types
	return []string{".pdf", ".docx", ".doc", ".xlsx", ".xls", ".pptx", ".ppt", ".txt", ".csv", ".jpg", ".jpeg", ".png"}
}

func (m *Manager) IsSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedTypes := m.GetSupportedFileTypes()
	for _, supported := range supportedTypes {
		if ext == supported {
			return true
		}
	}
	return false
}

func (m *Manager) GetDocumentType(filePath string) DocumentType {
	// Implementation based on file extension
	// This is a simplified version - the actual implementation might be more complex
	ext := filepath.Ext(strings.ToLower(filePath))
	switch ext {
	case ".pdf":
		return DocumentTypePDF
	case ".docx", ".doc":
		return DocumentTypeWord
	case ".xlsx", ".xls":
		return DocumentTypeExcel
	case ".pptx", ".ppt":
		return DocumentTypePowerPoint
	case ".txt":
		return DocumentTypeText
	case ".csv":
		return DocumentTypeCSV
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return DocumentTypeImage
	default:
		return DocumentTypeUnknown
	}
}

func (m *Manager) GetRoutingSummary(results []*RoutingResult) map[string]interface{} {
	return m.router.GetRoutingSummary(results)
}
