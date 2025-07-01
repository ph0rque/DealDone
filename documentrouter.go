package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentRouter handles routing documents to appropriate folders based on classification
type DocumentRouter struct {
	folderManager     *FolderManager
	documentProcessor *DocumentProcessor
}

// NewDocumentRouter creates a new document router
func NewDocumentRouter(folderManager *FolderManager, documentProcessor *DocumentProcessor) *DocumentRouter {
	return &DocumentRouter{
		folderManager:     folderManager,
		documentProcessor: documentProcessor,
	}
}

// RoutingResult represents the result of routing a document
type RoutingResult struct {
	SourcePath      string       `json:"sourcePath"`
	DestinationPath string       `json:"destinationPath"`
	DocumentType    DocumentType `json:"documentType"`
	Success         bool         `json:"success"`
	Error           string       `json:"error,omitempty"`
	ProcessingTime  int64        `json:"processingTimeMs"`
}

// RouteDocument processes and routes a single document to the appropriate folder
func (dr *DocumentRouter) RouteDocument(filePath string, dealName string) (*RoutingResult, error) {
	startTime := time.Now()

	result := &RoutingResult{
		SourcePath: filePath,
	}

	// Ensure deal folder exists
	if !dr.folderManager.DealExists(dealName) {
		_, err := dr.folderManager.CreateDealFolder(dealName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to create deal folder: %v", err)
			result.ProcessingTime = time.Since(startTime).Milliseconds()
			return result, err
		}
	}

	// Process document to determine type
	docInfo, err := dr.documentProcessor.ProcessDocument(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to process document: %v", err)
		result.ProcessingTime = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Check for unsupported files
	if docInfo.ErrorMessage != "" {
		result.Error = docInfo.ErrorMessage
		result.DocumentType = docInfo.Type
		result.ProcessingTime = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf(docInfo.ErrorMessage)
	}

	// Determine destination folder based on document type
	subfolder := dr.getSubfolderForType(docInfo.Type)
	destFolder := dr.folderManager.GetDealSubfolderPath(dealName, subfolder)

	// Ensure destination folder exists
	if err := dr.folderManager.EnsureFolderExists(destFolder); err != nil {
		result.Error = fmt.Sprintf("failed to create destination folder: %v", err)
		result.ProcessingTime = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Copy file to destination
	destPath := filepath.Join(destFolder, filepath.Base(filePath))
	if err := dr.copyFile(filePath, destPath); err != nil {
		result.Error = fmt.Sprintf("failed to copy file: %v", err)
		result.ProcessingTime = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Update result
	result.DestinationPath = destPath
	result.DocumentType = docInfo.Type
	result.Success = true
	result.ProcessingTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// RouteDocuments processes and routes multiple documents
func (dr *DocumentRouter) RouteDocuments(filePaths []string, dealName string) ([]*RoutingResult, error) {
	results := make([]*RoutingResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		result, err := dr.RouteDocument(filePath, dealName)
		if err != nil {
			// Log error but continue processing other files
			if result == nil {
				result = &RoutingResult{
					SourcePath: filePath,
					Error:      err.Error(),
				}
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// RouteFolder processes all documents in a folder
func (dr *DocumentRouter) RouteFolder(folderPath string, dealName string) ([]*RoutingResult, error) {
	// Get all files in folder
	files, err := dr.getFilesInFolder(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in folder: %w", err)
	}

	return dr.RouteDocuments(files, dealName)
}

// getSubfolderForType returns the appropriate subfolder name for a document type
func (dr *DocumentRouter) getSubfolderForType(docType DocumentType) string {
	switch docType {
	case DocTypeLegal:
		return "legal"
	case DocTypeFinancial:
		return "financial"
	case DocTypeGeneral:
		return "general"
	default:
		return "general"
	}
}

// copyFile copies a file from source to destination
func (dr *DocumentRouter) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Check if destination already exists
	if _, err := os.Stat(dst); err == nil {
		// Add timestamp to filename to avoid overwriting
		ext := filepath.Ext(dst)
		base := dst[:len(dst)-len(ext)]
		timestamp := time.Now().Format("20060102_150405")
		dst = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// getFilesInFolder recursively gets all files in a folder
func (dr *DocumentRouter) getFilesInFolder(folderPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// MoveDocument moves a document instead of copying it
func (dr *DocumentRouter) MoveDocument(filePath string, dealName string) (*RoutingResult, error) {
	// First route (copy) the document
	result, err := dr.RouteDocument(filePath, dealName)
	if err != nil {
		return result, err
	}

	// If successful, delete the original
	if result.Success {
		if err := os.Remove(filePath); err != nil {
			// Log error but don't fail the operation
			result.Error = fmt.Sprintf("file copied but failed to delete original: %v", err)
		}
	}

	return result, nil
}

// GetRoutingSummary returns a summary of routing results
func (dr *DocumentRouter) GetRoutingSummary(results []*RoutingResult) map[string]interface{} {
	summary := map[string]interface{}{
		"total":      len(results),
		"successful": 0,
		"failed":     0,
		"byType": map[string]int{
			"legal":     0,
			"financial": 0,
			"general":   0,
			"unknown":   0,
		},
		"avgProcessingTimeMs": int64(0),
	}

	totalTime := int64(0)
	byType := summary["byType"].(map[string]int)

	for _, result := range results {
		if result.Success {
			summary["successful"] = summary["successful"].(int) + 1
		} else {
			summary["failed"] = summary["failed"].(int) + 1
		}

		byType[string(result.DocumentType)]++
		totalTime += result.ProcessingTime
	}

	if len(results) > 0 {
		summary["avgProcessingTimeMs"] = totalTime / int64(len(results))
	}

	return summary
}
