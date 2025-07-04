package documents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentProcessor(t *testing.T) {
	// Create test AI service (nil for rule-based testing)
	dp := NewDocumentProcessor(nil)

	t.Run("IsSupportedFile detects supported files", func(t *testing.T) {
		supportedFiles := []string{
			"document.pdf",
			"contract.doc",
			"financial.xlsx",
			"presentation.pptx",
			"notes.txt",
			"data.csv",
			"scan.jpg",
			"image.png",
		}

		for _, file := range supportedFiles {
			if !dp.IsSupportedFile(file) {
				t.Errorf("Expected %s to be supported", file)
			}
		}

		unsupportedFiles := []string{
			"video.mp4",
			"audio.mp3",
			"archive.zip",
			"executable.exe",
			"noextension",
		}

		for _, file := range unsupportedFiles {
			if dp.IsSupportedFile(file) {
				t.Errorf("Expected %s to be unsupported", file)
			}
		}
	})

	t.Run("detectDocumentTypeByRules identifies legal documents", func(t *testing.T) {
		legalFiles := []string{
			"NDA_Company_2024.pdf",
			"Letter_of_Intent.docx",
			"Purchase_Agreement_Final.doc",
			"contract-services.pdf",
			"terms_and_conditions.txt",
			"patent_application.pdf",
			"litigation_summary.docx",
		}

		for _, file := range legalFiles {
			docType := dp.detectDocumentTypeByRules(file, filepath.Ext(file))
			if docType != DocTypeLegal {
				t.Errorf("Expected %s to be classified as legal, got %s", file, docType)
			}
		}
	})

	t.Run("detectDocumentTypeByRules identifies financial documents", func(t *testing.T) {
		financialFiles := []string{
			"P&L_Statement_2024.xlsx",
			"Balance_Sheet_Q4.xls",
			"cash_flow_projection.xlsx",
			"financial_model_v2.xlsx",
			"Revenue_Report.csv",
			"EBITDA_Analysis.xlsx",
			"tax_returns_2023.pdf",
		}

		for _, file := range financialFiles {
			docType := dp.detectDocumentTypeByRules(file, filepath.Ext(file))
			if docType != DocTypeFinancial {
				t.Errorf("Expected %s to be classified as financial, got %s", file, docType)
			}
		}
	})

	t.Run("detectDocumentTypeByRules defaults to general", func(t *testing.T) {
		generalFiles := []string{
			"company_overview.pdf",
			"meeting_notes.docx",
			"presentation.pptx",
			"misc_document.txt",
		}

		for _, file := range generalFiles {
			docType := dp.detectDocumentTypeByRules(file, filepath.Ext(file))
			if docType != DocTypeGeneral {
				t.Errorf("Expected %s to be classified as general, got %s", file, docType)
			}
		}
	})

	t.Run("ProcessDocument handles files correctly", func(t *testing.T) {
		// Create a temporary test file
		tempDir, err := os.MkdirTemp("", "docprocessor-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "test_financial_report.txt")
		testContent := []byte("This is a test financial report with revenue data")
		if err := os.WriteFile(testFile, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		docInfo, err := dp.ProcessDocument(testFile)
		if err != nil {
			t.Fatalf("Failed to process document: %v", err)
		}

		// Check basic info
		if docInfo.Name != "test_financial_report.txt" {
			t.Errorf("Expected name 'test_financial_report.txt', got %s", docInfo.Name)
		}

		if docInfo.Extension != ".txt" {
			t.Errorf("Expected extension '.txt', got %s", docInfo.Extension)
		}

		if docInfo.Type != DocTypeFinancial {
			t.Errorf("Expected type 'financial', got %s", docInfo.Type)
		}

		if docInfo.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), docInfo.Size)
		}

		if docInfo.IsScanned {
			t.Error("Text file should not be marked as scanned")
		}

		// Rule-based detection should have 0.7 confidence
		if docInfo.Confidence != 0.7 {
			t.Errorf("Expected confidence 0.7, got %f", docInfo.Confidence)
		}
	})

	t.Run("ProcessDocument handles unsupported files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docprocessor-test-unsupported-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "video.mp4")
		if err := os.WriteFile(testFile, []byte("fake video"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		docInfo, err := dp.ProcessDocument(testFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if docInfo.ErrorMessage != "Unsupported file type" {
			t.Errorf("Expected error message about unsupported type, got: %s", docInfo.ErrorMessage)
		}

		if docInfo.Type != DocTypeUnknown {
			t.Errorf("Expected type unknown for unsupported file, got %s", docInfo.Type)
		}
	})

	t.Run("ProcessDocument handles non-existent files", func(t *testing.T) {
		_, err := dp.ProcessDocument("/non/existent/file.pdf")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("ProcessDocument rejects directories", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docprocessor-test-dir-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		_, err = dp.ProcessDocument(tempDir)
		if err == nil {
			t.Error("Expected error for directory path")
		}
	})

	t.Run("isImageFile detects image files", func(t *testing.T) {
		imageExts := []string{".jpg", ".jpeg", ".png", ".tiff", ".bmp", ".gif"}
		for _, ext := range imageExts {
			if !dp.isImageFile(ext) {
				t.Errorf("Expected %s to be identified as image", ext)
			}
		}

		nonImageExts := []string{".pdf", ".doc", ".txt", ".xlsx"}
		for _, ext := range nonImageExts {
			if dp.isImageFile(ext) {
				t.Errorf("Expected %s NOT to be identified as image", ext)
			}
		}
	})

	t.Run("ProcessBatch handles multiple files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docprocessor-test-batch-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create test files
		files := []string{
			"contract.pdf",
			"financial_report.xlsx",
			"overview.txt",
		}

		filePaths := make([]string, len(files))
		for i, file := range files {
			path := filepath.Join(tempDir, file)
			os.WriteFile(path, []byte("test content"), 0644)
			filePaths[i] = path
		}

		// Add a non-existent file to test error handling
		filePaths = append(filePaths, filepath.Join(tempDir, "nonexistent.pdf"))

		results, err := dp.ProcessBatch(filePaths)
		if err != nil {
			t.Fatalf("ProcessBatch failed: %v", err)
		}

		if len(results) != len(filePaths) {
			t.Errorf("Expected %d results, got %d", len(filePaths), len(results))
		}

		// Check that the last one has an error
		lastResult := results[len(results)-1]
		if lastResult.ErrorMessage == "" {
			t.Error("Expected error message for non-existent file")
		}
	})

	t.Run("ExtractText works for text files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docprocessor-test-extract-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "This is test content"
		os.WriteFile(testFile, []byte(testContent), 0644)

		extracted, err := dp.ExtractText(testFile)
		if err != nil {
			t.Fatalf("Failed to extract text: %v", err)
		}

		if extracted != testContent {
			t.Errorf("Expected '%s', got '%s'", testContent, extracted)
		}
	})

	t.Run("GetDocumentMetadata returns basic metadata", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docprocessor-test-metadata-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		metadata, err := dp.GetDocumentMetadata(testFile)
		if err != nil {
			t.Fatalf("Failed to get metadata: %v", err)
		}

		if metadata["name"] != "test.txt" {
			t.Errorf("Expected name 'test.txt', got %v", metadata["name"])
		}

		if metadata["size"].(int64) != 4 {
			t.Errorf("Expected size 4, got %v", metadata["size"])
		}

		if metadata["modified"] == nil {
			t.Error("Expected modified time to be set")
		}
	})
}
