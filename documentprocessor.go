package main

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentType represents the type of document
type DocumentType string

const (
	DocTypeLegal     DocumentType = "legal"
	DocTypeFinancial DocumentType = "financial"
	DocTypeGeneral   DocumentType = "general"
	DocTypeUnknown   DocumentType = "unknown"
)

// DocumentInfo contains information about a processed document
type DocumentInfo struct {
	Path         string       `json:"path"`
	Name         string       `json:"name"`
	Type         DocumentType `json:"type"`
	MimeType     string       `json:"mimeType"`
	Size         int64        `json:"size"`
	Extension    string       `json:"extension"`
	IsScanned    bool         `json:"isScanned"`
	Confidence   float64      `json:"confidence"`
	Keywords     []string     `json:"keywords"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
}

// DocumentProcessor handles document analysis and classification
type DocumentProcessor struct {
	aiService     *AIService
	ocrService    *OCRService
	supportedExts map[string]bool
}

// NewDocumentProcessor creates a new document processor
func NewDocumentProcessor(aiService *AIService) *DocumentProcessor {
	return &DocumentProcessor{
		aiService:  aiService,
		ocrService: NewOCRService(""), // Will be configured later
		supportedExts: map[string]bool{
			".pdf":  true,
			".doc":  true,
			".docx": true,
			".xls":  true,
			".xlsx": true,
			".ppt":  true,
			".pptx": true,
			".txt":  true,
			".md":   true,
			".rtf":  true,
			".csv":  true,
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".tiff": true,
			".bmp":  true,
		},
	}
}

// SetOCRService sets the OCR service for the document processor
func (dp *DocumentProcessor) SetOCRService(ocrService *OCRService) {
	dp.ocrService = ocrService
}

// ProcessDocument analyzes a document and returns its information
func (dp *DocumentProcessor) ProcessDocument(filePath string) (*DocumentInfo, error) {
	// Basic file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Get file details
	fileName := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType := mime.TypeByExtension(ext)

	docInfo := &DocumentInfo{
		Path:      filePath,
		Name:      fileName,
		Size:      info.Size(),
		Extension: ext,
		MimeType:  mimeType,
		Type:      DocTypeUnknown,
	}

	// Check if supported
	if !dp.IsSupportedFile(fileName) {
		docInfo.ErrorMessage = "Unsupported file type"
		return docInfo, nil
	}

	// Check if it's an image (potentially scanned document)
	if dp.isImageFile(ext) {
		docInfo.IsScanned = true
	}

	// First try rule-based classification for obvious cases
	ruleBasedType := dp.detectDocumentTypeByRules(fileName, ext)
	fmt.Printf("Document classification for %s: rule-based=%s\n", fileName, ruleBasedType)

	// If we have a strong rule-based match (not general), use it
	if ruleBasedType != DocTypeGeneral {
		docInfo.Type = ruleBasedType
		docInfo.Confidence = 0.9 // High confidence for rule-based matches
		fmt.Printf("Using rule-based classification for %s: %s\n", fileName, ruleBasedType)
		return docInfo, nil
	}

	// Otherwise, try AI classification for general documents
	if dp.aiService != nil {
		aiResult, err := dp.detectDocumentTypeWithAI(filePath, docInfo)
		if err == nil {
			docInfo.Type = aiResult.Type
			docInfo.Confidence = aiResult.Confidence
			docInfo.Keywords = aiResult.Keywords
			fmt.Printf("Using AI classification for %s: %s (confidence: %.2f)\n", fileName, aiResult.Type, aiResult.Confidence)
		} else {
			// Fall back to rule-based detection
			docInfo.Type = ruleBasedType
			docInfo.Confidence = 0.7 // Lower confidence for rule-based
			fmt.Printf("AI classification failed for %s, using rule-based: %s\n", fileName, ruleBasedType)
		}
	} else {
		// No AI service, use rules
		docInfo.Type = ruleBasedType
		docInfo.Confidence = 0.7
		fmt.Printf("No AI service, using rule-based for %s: %s\n", fileName, ruleBasedType)
	}

	return docInfo, nil
}

// AIDetectionResult represents the result from AI analysis
type AIDetectionResult struct {
	Type       DocumentType
	Confidence float64
	Keywords   []string
}

// detectDocumentTypeWithAI uses AI service to detect document type
func (dp *DocumentProcessor) detectDocumentTypeWithAI(filePath string, docInfo *DocumentInfo) (*AIDetectionResult, error) {
	// Extract text from the document
	text, err := dp.ExtractTextWithOCR(filePath)
	if err != nil {
		// If text extraction fails, try without OCR
		text, err = dp.ExtractText(filePath)
		if err != nil {
			return nil, err
		}
	}

	// Use AI to classify based on extracted text
	if text != "" && dp.aiService != nil && dp.aiService.IsAvailable() {
		metadata := map[string]interface{}{
			"filename":  docInfo.Name,
			"extension": docInfo.Extension,
			"size":      docInfo.Size,
		}

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		result, err := dp.aiService.ClassifyDocument(ctx, text, metadata)
		if err != nil {
			return nil, err
		}

		// Convert AI result to our format
		docType := DocTypeGeneral
		switch result.DocumentType {
		case "legal":
			docType = DocTypeLegal
		case "financial":
			docType = DocTypeFinancial
		}

		return &AIDetectionResult{
			Type:       docType,
			Confidence: result.Confidence,
			Keywords:   result.Keywords,
		}, nil
	}

	// Fallback to rule-based
	return &AIDetectionResult{
		Type:       dp.detectDocumentTypeByRules(docInfo.Name, docInfo.Extension),
		Confidence: 0.7,
		Keywords:   []string{},
	}, nil
}

// detectDocumentTypeByRules uses rule-based detection as fallback
func (dp *DocumentProcessor) detectDocumentTypeByRules(fileName, ext string) DocumentType {
	lowerName := strings.ToLower(fileName)

	// Legal document patterns
	legalPatterns := []string{
		"nda", "non-disclosure", "non_disclosure",
		"loi", "letter_of_intent", "letter-of-intent",
		"purchase_agreement", "purchase-agreement", "spa",
		"contract", "agreement", "legal",
		"terms", "conditions", "bylaws",
		"articles_of_incorporation", "charter",
		"intellectual_property", "patent", "trademark",
		"litigation", "lawsuit", "court",
		"regulatory", "compliance", "license",
	}

	for _, pattern := range legalPatterns {
		if strings.Contains(lowerName, pattern) {
			return DocTypeLegal
		}
	}

	// General document patterns (including CIM which is marketing/overview)
	generalPatterns := []string{
		"cim", "confidential_information", "confidential-information",
		"teaser", "pitch", "deck", "overview", "summary",
		"presentation", "marketing", "brochure", "profile",
		"company_profile", "company-profile", "executive_summary", "executive-summary",
		"management_presentation", "management-presentation",
	}

	for _, pattern := range generalPatterns {
		if strings.Contains(lowerName, pattern) {
			return DocTypeGeneral
		}
	}

	// Enhanced Financial document patterns (excluding CIM and marketing docs)
	financialPatterns := []string{
		"financial", "financials", "finance", "finances",
		"p&l", "pnl", "pandl", "profit_loss", "profit_and_loss", "profit-loss",
		"balance_sheet", "balance-sheet", "balance", "sheet",
		"income_statement", "income-statement", "income",
		"cash_flow", "cash-flow", "cashflow", "cash",
		"revenue", "revenues", "ebitda", "earnings", "earning",
		"budget", "budgets", "forecast", "forecasts", "projection", "projections",
		"tax", "taxes", "audit", "audits", "accounting", "accounts",
		"bank_statement", "bank-statement", "bank", "statement",
		"valuation", "valuations", "financial_model", "financial-model", "model", "models",
		"metrics", "metric", "kpi", "kpis", "performance", "performances",
		"annual_report", "annual-report", "quarterly", "monthly",
		"investor", "investment",
		"due_diligence", "due-diligence", "dd", "diligence",
	}

	for _, pattern := range financialPatterns {
		if strings.Contains(lowerName, pattern) {
			return DocTypeFinancial
		}
	}

	// Spreadsheets are often financial
	if ext == ".xls" || ext == ".xlsx" || ext == ".csv" {
		return DocTypeFinancial
	}

	// Default to general
	return DocTypeGeneral
}

// IsSupportedFile checks if a file type is supported
func (dp *DocumentProcessor) IsSupportedFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return dp.supportedExts[ext]
}

// GetSupportedExtensions returns list of supported file extensions
func (dp *DocumentProcessor) GetSupportedExtensions() []string {
	exts := make([]string, 0, len(dp.supportedExts))
	for ext := range dp.supportedExts {
		exts = append(exts, ext)
	}
	return exts
}

// isImageFile checks if file is an image
func (dp *DocumentProcessor) isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".tiff": true,
		".bmp":  true,
		".gif":  true,
	}
	return imageExts[ext]
}

// ProcessBatch processes multiple documents
func (dp *DocumentProcessor) ProcessBatch(filePaths []string) ([]*DocumentInfo, error) {
	results := make([]*DocumentInfo, 0, len(filePaths))

	for _, path := range filePaths {
		docInfo, err := dp.ProcessDocument(path)
		if err != nil {
			// Add error info but continue processing
			docInfo = &DocumentInfo{
				Path:         path,
				Name:         filepath.Base(path),
				Type:         DocTypeUnknown,
				ErrorMessage: err.Error(),
			}
		}
		results = append(results, docInfo)
	}

	return results, nil
}

// ExtractText extracts text content from a document
func (dp *DocumentProcessor) ExtractText(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".txt", ".md":
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read text file: %w", err)
		}
		return string(content), nil

	case ".pdf", ".doc", ".docx":
		// Use OCR if available for these document types
		if dp.ocrService != nil && dp.ocrService.IsEnabled() {
			return dp.ExtractTextWithOCR(filePath)
		}
		return "", fmt.Errorf("text extraction for %s requires OCR service", ext)

	default:
		// For images, always use OCR
		if dp.isImageFile(ext) {
			return dp.ExtractTextWithOCR(filePath)
		}
		return "", fmt.Errorf("text extraction not supported for %s files", ext)
	}
}

// ExtractTextWithOCR uses OCR to extract text from images or scanned documents
func (dp *DocumentProcessor) ExtractTextWithOCR(filePath string) (string, error) {
	if dp.ocrService == nil || !dp.ocrService.IsEnabled() {
		return "", fmt.Errorf("OCR service not available")
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle PDFs separately
	if ext == ".pdf" {
		result, err := dp.ocrService.ProcessPDF(filePath)
		if err != nil {
			return "", fmt.Errorf("OCR failed for PDF: %w", err)
		}
		return result.Text, nil
	}

	// Handle images
	if dp.isImageFile(ext) {
		// Preprocess image if needed
		processedPath, err := dp.ocrService.PreprocessImage(filePath)
		if err != nil {
			// Continue with original if preprocessing fails
			processedPath = filePath
		}

		result, err := dp.ocrService.ProcessImage(processedPath)
		if err != nil {
			return "", fmt.Errorf("OCR failed for image: %w", err)
		}
		return result.Text, nil
	}

	return "", fmt.Errorf("file type %s not supported for OCR", ext)
}

// GetDocumentMetadata extracts metadata from a document
func (dp *DocumentProcessor) GetDocumentMetadata(filePath string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Basic file metadata
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	metadata["size"] = info.Size()
	metadata["modified"] = info.ModTime()
	metadata["name"] = info.Name()

	// Add more metadata based on file type
	// This is a placeholder for actual implementation

	return metadata, nil
}
