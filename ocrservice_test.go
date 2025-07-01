package main

import (
	"testing"
)

func TestOCRService(t *testing.T) {
	t.Run("NewOCRService enables service with provider", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		if !ocr.enabled {
			t.Error("Expected OCR to be enabled with provider")
		}

		if ocr.provider != "tesseract" {
			t.Errorf("Expected provider 'tesseract', got '%s'", ocr.provider)
		}
	})

	t.Run("NewOCRService disables service without provider", func(t *testing.T) {
		ocr := NewOCRService("")

		if ocr.enabled {
			t.Error("Expected OCR to be disabled without provider")
		}
	})

	t.Run("ProcessImage requires enabled service", func(t *testing.T) {
		ocr := NewOCRService("")

		_, err := ocr.ProcessImage("test.jpg")
		if err == nil {
			t.Error("Expected error when OCR is disabled")
		}
	})

	t.Run("ProcessImage validates image format", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		// Unsupported format
		_, err := ocr.ProcessImage("document.txt")
		if err == nil {
			t.Error("Expected error for unsupported format")
		}

		// Supported formats should work (placeholder implementation)
		supportedFormats := []string{
			"scan.jpg",
			"image.png",
			"document.tiff",
			"page.bmp",
		}

		for _, file := range supportedFormats {
			result, err := ocr.ProcessImage(file)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", file, err)
			}

			if result.Text == "" {
				t.Errorf("Expected text result for %s", file)
			}

			if result.Confidence <= 0 || result.Confidence > 1 {
				t.Errorf("Expected confidence between 0 and 1 for %s, got %f", file, result.Confidence)
			}
		}
	})

	t.Run("ProcessPDF handles PDF files", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		result, err := ocr.ProcessPDF("scanned.pdf")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Text == "" {
			t.Error("Expected text from PDF")
		}

		if result.PageCount <= 0 {
			t.Error("Expected positive page count")
		}
	})

	t.Run("isSupportedImageFormat checks formats", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		supportedFormats := []string{
			"test.jpg", "test.jpeg", "test.png",
			"test.tiff", "test.tif", "test.bmp", "test.pdf",
		}

		for _, file := range supportedFormats {
			if !ocr.isSupportedImageFormat(file) {
				t.Errorf("Expected %s to be supported", file)
			}
		}

		unsupportedFormats := []string{
			"test.txt", "test.doc", "test.xlsx", "test.mp4",
		}

		for _, file := range unsupportedFormats {
			if ocr.isSupportedImageFormat(file) {
				t.Errorf("Expected %s to be unsupported", file)
			}
		}
	})

	t.Run("GetSupportedLanguages returns languages", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		languages := ocr.GetSupportedLanguages()
		if len(languages) == 0 {
			t.Error("Expected at least one supported language")
		}

		// Check for common languages
		hasEnglish := false
		for _, lang := range languages {
			if lang == "en" {
				hasEnglish = true
				break
			}
		}

		if !hasEnglish {
			t.Error("Expected English to be supported")
		}
	})

	t.Run("ConfigureLanguage validates language", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		// Valid language
		err := ocr.ConfigureLanguage("en")
		if err != nil {
			t.Errorf("Unexpected error for valid language: %v", err)
		}

		// Invalid language
		err = ocr.ConfigureLanguage("xx")
		if err == nil {
			t.Error("Expected error for invalid language")
		}
	})

	t.Run("BatchProcess handles multiple images", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		images := []string{"scan1.jpg", "scan2.png", "scan3.tiff"}
		results, err := ocr.BatchProcess(images)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(results) != len(images) {
			t.Errorf("Expected %d results, got %d", len(images), len(results))
		}

		for i, result := range results {
			if result == nil {
				t.Errorf("Result %d is nil", i)
			}
		}
	})

	t.Run("ExtractTables returns table data", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		tables, err := ocr.ExtractTables("table.jpg")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(tables) == 0 {
			t.Error("Expected at least one table")
		}

		if len(tables[0]) == 0 {
			t.Error("Expected table to have columns")
		}
	})

	t.Run("PreprocessImage returns processed path", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		processedPath, err := ocr.PreprocessImage("original.jpg")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if processedPath == "" {
			t.Error("Expected processed image path")
		}
	})

	t.Run("DetectTextOrientation returns rotation", func(t *testing.T) {
		ocr := NewOCRService("tesseract")

		rotation, err := ocr.DetectTextOrientation("rotated.jpg")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		validRotations := map[int]bool{0: true, 90: true, 180: true, 270: true}
		if !validRotations[rotation] {
			t.Errorf("Invalid rotation value: %d", rotation)
		}
	})
}
