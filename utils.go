package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Numeric type constraint for min/max functions
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// min returns the smaller of two comparable values
func min[T Numeric](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two comparable values
func max[T Numeric](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// createFileSystemID generates a unique ID for a file system item based on its path
func createFileSystemID(path string) string {
	hash := md5.Sum([]byte(path))
	return hex.EncodeToString(hash[:])
}

// getFileExtension extracts the file extension from a filename
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return strings.ToLower(ext[1:]) // Remove the dot and convert to lowercase
	}
	return ""
}

// getMimeType determines the MIME type of a file based on its extension
func getMimeType(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return mime.TypeByExtension(ext)
	}
	return ""
}

// getFilePermissions checks file permissions for read, write, and execute
func getFilePermissions(path string) FilePermissions {
	perms := FilePermissions{
		Readable:   false,
		Writable:   false,
		Executable: false,
	}

	info, err := os.Stat(path)
	if err != nil {
		return perms
	}

	mode := info.Mode()

	// Check readable
	if mode&0400 != 0 { // Owner read permission
		perms.Readable = true
	}

	// Check writable
	if mode&0200 != 0 { // Owner write permission
		perms.Writable = true
	}

	// Check executable
	if mode&0100 != 0 { // Owner execute permission
		perms.Executable = true
	}

	return perms
}

// isHiddenFile checks if a file is hidden (starts with dot on Unix systems)
func isHiddenFile(name string) bool {
	if runtime.GOOS == "windows" {
		// On Windows, check file attributes if needed
		return false
	}
	return strings.HasPrefix(name, ".")
}

// isValidPath checks if a path is valid and accessible
func isValidPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ensureDirectory creates a directory if it doesn't exist
func ensureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// joinPath safely joins path components
func joinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// cleanPath cleans and normalizes a file path
func cleanPath(path string) string {
	return filepath.Clean(path)
}

// getParentDir returns the parent directory of a path
func getParentDir(path string) string {
	return filepath.Dir(path)
}

// getBaseName returns the base name of a file path
func getBaseName(path string) string {
	return filepath.Base(path)
}

// fileExists checks if a file or directory exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// isDirectory checks if a path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isFile checks if a path is a regular file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// getFileSize returns the size of a file in bytes
func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// getModTime returns the modification time of a file
func getModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// getBirthTime returns the creation time of a file (best effort)
func getBirthTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	// On most systems, we use ModTime as creation time
	// This could be enhanced with platform-specific code
	return info.ModTime()
}

// formatError creates a standardized error message
func formatError(operation string, path string, err error) string {
	return fmt.Sprintf("Failed to %s '%s': %v", operation, path, err)
}

// validateFileName checks if a filename is valid
func validateFileName(name string) error {
	if name == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	// Check for reserved names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	for _, reserved := range reservedNames {
		if upperName == reserved || strings.HasPrefix(upperName, reserved+".") {
			return fmt.Errorf("filename uses reserved name: %s", name)
		}
	}

	return nil
}
