package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// createFileSystemItem creates a FileSystemItem from a file path
func createFileSystemItem(path string) (*FileSystemItem, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	name := getBaseName(path)
	isDir := info.IsDir()

	item := &FileSystemItem{
		ID:          createFileSystemID(path),
		Name:        name,
		Path:        path,
		IsDirectory: isDir,
		Size:        info.Size(),
		ModifiedAt:  getModTime(path),
		CreatedAt:   getBirthTime(path),
		Permissions: getFilePermissions(path),
	}

	if !isDir {
		item.Extension = getFileExtension(name)
		item.MimeType = getMimeType(name)
	}

	return item, nil
}

// Sub-task 2.2: ListDirectory reads directory contents and returns file listing
func (a *App) ListDirectory(request DirectoryListRequest) ([]FileSystemItem, error) {
	if !isValidPath(request.Path) {
		return nil, fmt.Errorf("invalid path: %s", request.Path)
	}

	if !isDirectory(request.Path) {
		return nil, fmt.Errorf("path is not a directory: %s", request.Path)
	}

	entries, err := os.ReadDir(request.Path)
	if err != nil {
		return nil, fmt.Errorf(formatError("read directory", request.Path, err))
	}

	// Initialize as empty slice to ensure JSON serialization returns [] not null
	items := make([]FileSystemItem, 0)

	for _, entry := range entries {
		// Skip hidden files if not requested
		if !request.ShowHidden && isHiddenFile(entry.Name()) {
			continue
		}

		fullPath := joinPath(request.Path, entry.Name())
		item, err := createFileSystemItem(fullPath)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Could not process %s: %v\n", fullPath, err)
			continue
		}

		items = append(items, *item)
	}

	// Sort items: directories first, then files, alphabetically
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDirectory && !items[j].IsDirectory {
			return true
		}
		if !items[i].IsDirectory && items[j].IsDirectory {
			return false
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}

// Sub-task 2.3: CreateFile creates a new file
func (a *App) CreateFile(request CreateFileRequest) FileOperationResult {
	if err := validateFileName(request.Name); err != nil {
		return FileOperationResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	fullPath := joinPath(request.Path, request.Name)

	if fileExists(fullPath) {
		return FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("file or directory already exists: %s", request.Name),
		}
	}

	var err error
	if request.IsDirectory {
		err = os.MkdirAll(fullPath, 0755)
		if err != nil {
			return FileOperationResult{
				Success: false,
				Error:   formatError("create directory", fullPath, err),
			}
		}
	} else {
		file, err := os.Create(fullPath)
		if err != nil {
			return FileOperationResult{
				Success: false,
				Error:   formatError("create file", fullPath, err),
			}
		}
		file.Close()
	}

	itemType := "file"
	if request.IsDirectory {
		itemType = "directory"
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created %s: %s", itemType, request.Name),
	}
}

// Sub-task 2.4: CopyFile copies a file or directory
func (a *App) CopyFile(source, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return copyDirectory(source, destination)
	}
	return copyFile(source, destination)
}

// copyFile copies a single file
func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	destDir := getParentDir(destination)
	if err := ensureDirectory(destDir); err != nil {
		return err
	}

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(destination, sourceInfo.Mode())
}

// copyDirectory recursively copies a directory
func copyDirectory(source, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(destination, sourceInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := joinPath(source, entry.Name())
		destPath := joinPath(destination, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyItems copies multiple files/directories
func (a *App) CopyItems(request CopyMoveRequest) FileOperationResult {
	if !isDirectory(request.TargetPath) {
		return FileOperationResult{
			Success: false,
			Error:   "target path must be a directory",
		}
	}

	var errors []string
	successCount := 0

	for _, sourcePath := range request.SourcePaths {
		if !fileExists(sourcePath) {
			errors = append(errors, fmt.Sprintf("source does not exist: %s", sourcePath))
			continue
		}

		sourceName := getBaseName(sourcePath)
		destPath := joinPath(request.TargetPath, sourceName)

		if err := a.CopyFile(sourcePath, destPath); err != nil {
			errors = append(errors, formatError("copy", sourcePath, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return FileOperationResult{
			Success: successCount > 0,
			Error:   strings.Join(errors, "; "),
			Message: fmt.Sprintf("Copied %d items, %d errors", successCount, len(errors)),
		}
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully copied %d items", successCount),
	}
}

// MoveItems moves multiple files/directories
func (a *App) MoveItems(request CopyMoveRequest) FileOperationResult {
	if !isDirectory(request.TargetPath) {
		return FileOperationResult{
			Success: false,
			Error:   "target path must be a directory",
		}
	}

	var errors []string
	successCount := 0

	for _, sourcePath := range request.SourcePaths {
		if !fileExists(sourcePath) {
			errors = append(errors, fmt.Sprintf("source does not exist: %s", sourcePath))
			continue
		}

		sourceName := getBaseName(sourcePath)
		destPath := joinPath(request.TargetPath, sourceName)

		// Try rename first (fastest for same filesystem)
		if err := os.Rename(sourcePath, destPath); err != nil {
			// If rename fails, try copy then delete
			if copyErr := a.CopyFile(sourcePath, destPath); copyErr != nil {
				errors = append(errors, formatError("move", sourcePath, copyErr))
				continue
			}
			if deleteErr := os.RemoveAll(sourcePath); deleteErr != nil {
				errors = append(errors, formatError("delete after copy", sourcePath, deleteErr))
				continue
			}
		}
		successCount++
	}

	if len(errors) > 0 {
		return FileOperationResult{
			Success: successCount > 0,
			Error:   strings.Join(errors, "; "),
			Message: fmt.Sprintf("Moved %d items, %d errors", successCount, len(errors)),
		}
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully moved %d items", successCount),
	}
}

// Sub-task 2.5: DeleteItems deletes files/directories with error handling
func (a *App) DeleteItems(request DeleteRequest) FileOperationResult {
	var errors []string
	successCount := 0

	for _, path := range request.Paths {
		if !fileExists(path) {
			errors = append(errors, fmt.Sprintf("path does not exist: %s", path))
			continue
		}

		if err := os.RemoveAll(path); err != nil {
			errors = append(errors, formatError("delete", path, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return FileOperationResult{
			Success: successCount > 0,
			Error:   strings.Join(errors, "; "),
			Message: fmt.Sprintf("Deleted %d items, %d errors", successCount, len(errors)),
		}
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted %d items", successCount),
	}
}

// Sub-task 2.6: RenameItem renames a file or directory
func (a *App) RenameItem(request RenameRequest) FileOperationResult {
	if err := validateFileName(request.NewName); err != nil {
		return FileOperationResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	if !fileExists(request.Path) {
		return FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("path does not exist: %s", request.Path),
		}
	}

	parentDir := getParentDir(request.Path)
	newPath := joinPath(parentDir, request.NewName)

	if fileExists(newPath) {
		return FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("file or directory already exists: %s", request.NewName),
		}
	}

	if err := os.Rename(request.Path, newPath); err != nil {
		return FileOperationResult{
			Success: false,
			Error:   formatError("rename", request.Path, err),
		}
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully renamed to: %s", request.NewName),
	}
}

// Sub-task 2.7: SearchFiles searches for files and folders by name
func (a *App) SearchFiles(request SearchRequest) (SearchResult, error) {
	if !isValidPath(request.Path) {
		return SearchResult{}, fmt.Errorf("invalid search path: %s", request.Path)
	}

	query := strings.ToLower(strings.TrimSpace(request.Query))
	if query == "" {
		return SearchResult{
			Items:      []FileSystemItem{},
			Query:      request.Query,
			TotalCount: 0,
		}, nil
	}

	// Initialize as empty slice to ensure JSON serialization returns [] not null
	results := make([]FileSystemItem, 0)

	err := filepath.Walk(request.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't read
			return nil
		}

		// Skip hidden files
		if isHiddenFile(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if filename contains search query
		if strings.Contains(strings.ToLower(info.Name()), query) {
			item, err := createFileSystemItem(path)
			if err == nil {
				results = append(results, *item)
			}
		}

		return nil
	})

	if err != nil {
		return SearchResult{}, fmt.Errorf(formatError("search", request.Path, err))
	}

	// Sort results: directories first, then files, alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDirectory && !results[j].IsDirectory {
			return true
		}
		if !results[i].IsDirectory && results[j].IsDirectory {
			return false
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return SearchResult{
		Items:      results,
		Query:      request.Query,
		TotalCount: len(results),
	}, nil
}

// Sub-task 2.8: OpenFile opens a file with system default application
func (a *App) OpenFile(request OpenFileRequest) FileOperationResult {
	if !fileExists(request.Path) {
		return FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("file does not exist: %s", request.Path),
		}
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", request.Path)
	case "linux":
		cmd = exec.Command("xdg-open", request.Path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", request.Path)
	default:
		return FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported operating system: %s", runtime.GOOS),
		}
	}

	if err := cmd.Start(); err != nil {
		return FileOperationResult{
			Success: false,
			Error:   formatError("open file", request.Path, err),
		}
	}

	return FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully opened: %s", getBaseName(request.Path)),
	}
}
