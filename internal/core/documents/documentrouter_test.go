package documents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentRouter(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "docrouter-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test infrastructure
	cs := &ConfigService{
		config: &Config{
			DealDoneRoot: filepath.Join(tempDir, "DealDone"),
		},
	}

	fm := NewFolderManager(cs)
	fm.InitializeFolderStructure()

	dp := NewDocumentProcessor(nil) // No AI service for tests
	dr := NewDocumentRouter(fm, dp)

	t.Run("RouteDocument creates deal folder if needed", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "test_contract.pdf")
		os.WriteFile(testFile, []byte("test content"), 0644)

		dealName := "NewDeal"
		result, err := dr.RouteDocument(testFile, dealName)

		if err != nil {
			t.Fatalf("Failed to route document: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected routing to succeed, got error: %s", result.Error)
		}

		// Check deal folder was created
		if !fm.DealExists(dealName) {
			t.Error("Deal folder was not created")
		}

		// Check file was copied to correct location
		if _, err := os.Stat(result.DestinationPath); os.IsNotExist(err) {
			t.Error("File was not copied to destination")
		}

		// Should be in legal folder based on filename
		if result.DocumentType != DocTypeLegal {
			t.Errorf("Expected document type 'legal', got %s", result.DocumentType)
		}
	})

	t.Run("RouteDocument handles different document types", func(t *testing.T) {
		dealName := "TestDeal"
		fm.CreateDealFolder(dealName)

		testCases := []struct {
			filename     string
			expectedType DocumentType
			subfolder    string
		}{
			{"NDA_2024.pdf", DocTypeLegal, "legal"},
			{"Financial_Statement.xlsx", DocTypeFinancial, "financial"},
			{"Company_Overview.docx", DocTypeGeneral, "general"},
		}

		for _, tc := range testCases {
			testFile := filepath.Join(tempDir, tc.filename)
			os.WriteFile(testFile, []byte("test"), 0644)

			result, err := dr.RouteDocument(testFile, dealName)
			if err != nil {
				t.Errorf("Failed to route %s: %v", tc.filename, err)
				continue
			}

			if result.DocumentType != tc.expectedType {
				t.Errorf("File %s: expected type %s, got %s", tc.filename, tc.expectedType, result.DocumentType)
			}

			// Check file is in correct subfolder
			expectedPath := filepath.Join(cs.GetDealsPath(), dealName, tc.subfolder, tc.filename)
			if result.DestinationPath != expectedPath {
				t.Errorf("File %s: expected destination %s, got %s", tc.filename, expectedPath, result.DestinationPath)
			}
		}
	})

	t.Run("RouteDocument handles unsupported files", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "video.mp4")
		os.WriteFile(testFile, []byte("fake video"), 0644)

		result, err := dr.RouteDocument(testFile, "TestDeal")

		if err == nil {
			t.Error("Expected error for unsupported file type")
		}

		if result.Error == "" {
			t.Error("Expected error message in result")
		}

		if result.Success {
			t.Error("Expected success to be false for unsupported file")
		}
	})

	t.Run("RouteDocuments processes multiple files", func(t *testing.T) {
		dealName := "BatchDeal"
		files := []string{
			filepath.Join(tempDir, "doc1.pdf"),
			filepath.Join(tempDir, "doc2.xlsx"),
			filepath.Join(tempDir, "doc3.txt"),
		}

		// Create test files
		for _, file := range files {
			os.WriteFile(file, []byte("test"), 0644)
		}

		results, err := dr.RouteDocuments(files, dealName)
		if err != nil {
			t.Fatalf("Failed to route documents: %v", err)
		}

		if len(results) != len(files) {
			t.Errorf("Expected %d results, got %d", len(files), len(results))
		}

		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
			}
		}

		if successCount != len(files) {
			t.Errorf("Expected all %d files to succeed, got %d", len(files), successCount)
		}
	})

	t.Run("RouteFolder processes all files in folder", func(t *testing.T) {
		// Create test folder with files
		testFolder := filepath.Join(tempDir, "test_folder")
		os.MkdirAll(testFolder, 0755)

		files := []string{"file1.pdf", "file2.docx", "file3.xlsx"}
		for _, file := range files {
			os.WriteFile(filepath.Join(testFolder, file), []byte("test"), 0644)
		}

		// Create a subdirectory with a file
		subDir := filepath.Join(testFolder, "subdir")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "file4.txt"), []byte("test"), 0644)

		dealName := "FolderDeal"
		results, err := dr.RouteFolder(testFolder, dealName)

		if err != nil {
			t.Fatalf("Failed to route folder: %v", err)
		}

		// Should find all 4 files (including in subdirectory)
		if len(results) != 4 {
			t.Errorf("Expected 4 results, got %d", len(results))
		}
	})

	t.Run("MoveDocument deletes original after routing", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "move_test.pdf")
		os.WriteFile(testFile, []byte("test"), 0644)

		dealName := "MoveDeal"
		result, err := dr.MoveDocument(testFile, dealName)

		if err != nil {
			t.Fatalf("Failed to move document: %v", err)
		}

		if !result.Success {
			t.Error("Expected move to succeed")
		}

		// Original file should be deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("Original file was not deleted")
		}

		// Destination should exist
		if _, err := os.Stat(result.DestinationPath); os.IsNotExist(err) {
			t.Error("Destination file does not exist")
		}
	})

	t.Run("getSubfolderForType returns correct folders", func(t *testing.T) {
		testCases := []struct {
			docType  DocumentType
			expected string
		}{
			{DocTypeLegal, "legal"},
			{DocTypeFinancial, "financial"},
			{DocTypeGeneral, "general"},
			{DocTypeUnknown, "general"},
		}

		for _, tc := range testCases {
			result := dr.getSubfolderForType(tc.docType)
			if result != tc.expected {
				t.Errorf("For type %s, expected %s, got %s", tc.docType, tc.expected, result)
			}
		}
	})

	t.Run("copyFile handles existing files", func(t *testing.T) {
		src := filepath.Join(tempDir, "source.txt")
		dst := filepath.Join(tempDir, "dest.txt")

		os.WriteFile(src, []byte("content"), 0644)
		os.WriteFile(dst, []byte("existing"), 0644)

		err := dr.copyFile(src, dst)
		if err != nil {
			t.Fatalf("Failed to copy file: %v", err)
		}

		// Original destination should still exist
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Error("Original destination file was overwritten")
		}

		// Should have created a new file with timestamp
		files, _ := filepath.Glob(filepath.Join(tempDir, "dest_*"))
		if len(files) == 0 {
			t.Error("No timestamped file was created")
		}
	})

	t.Run("GetRoutingSummary provides correct statistics", func(t *testing.T) {
		results := []*RoutingResult{
			{Success: true, DocumentType: DocTypeLegal, ProcessingTime: 100},
			{Success: true, DocumentType: DocTypeFinancial, ProcessingTime: 150},
			{Success: false, DocumentType: DocTypeGeneral, ProcessingTime: 50},
			{Success: true, DocumentType: DocTypeLegal, ProcessingTime: 120},
		}

		summary := dr.GetRoutingSummary(results)

		if summary["total"] != 4 {
			t.Errorf("Expected total 4, got %v", summary["total"])
		}

		if summary["successful"] != 3 {
			t.Errorf("Expected 3 successful, got %v", summary["successful"])
		}

		if summary["failed"] != 1 {
			t.Errorf("Expected 1 failed, got %v", summary["failed"])
		}

		byType := summary["byType"].(map[string]int)
		if byType["legal"] != 2 {
			t.Errorf("Expected 2 legal documents, got %d", byType["legal"])
		}

		avgTime := summary["avgProcessingTimeMs"].(int64)
		expectedAvg := int64(105) // (100 + 150 + 50 + 120) / 4
		if avgTime != expectedAvg {
			t.Errorf("Expected average time %d, got %d", expectedAvg, avgTime)
		}
	})
}
