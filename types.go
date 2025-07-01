package main

import (
	"time"
)

// FileSystemItem represents a file or directory in the file system
type FileSystemItem struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	IsDirectory bool              `json:"isDirectory"`
	Size        int64             `json:"size"`
	ModifiedAt  time.Time         `json:"modifiedAt"`
	CreatedAt   time.Time         `json:"createdAt"`
	Permissions FilePermissions   `json:"permissions"`
	Extension   string            `json:"extension,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
	Children    []FileSystemItem  `json:"children,omitempty"`
	IsExpanded  bool              `json:"isExpanded,omitempty"`
	IsLoading   bool              `json:"isLoading,omitempty"`
}

// FilePermissions represents file access permissions
type FilePermissions struct {
	Readable   bool `json:"readable"`
	Writable   bool `json:"writable"`
	Executable bool `json:"executable"`
}

// FileOperationResult represents the result of a file operation
type FileOperationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Items      []FileSystemItem `json:"items"`
	Query      string           `json:"query"`
	TotalCount int              `json:"totalCount"`
}

// FileOperation represents different types of file operations
type FileOperation string

const (
	OperationCreate FileOperation = "create"
	OperationCopy   FileOperation = "copy"
	OperationMove   FileOperation = "move"
	OperationDelete FileOperation = "delete"
	OperationRename FileOperation = "rename"
	OperationOpen   FileOperation = "open"
)

// CreateFileRequest represents a request to create a file
type CreateFileRequest struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	IsDirectory bool   `json:"isDirectory"`
}

// CopyMoveRequest represents a request to copy or move files
type CopyMoveRequest struct {
	SourcePaths []string `json:"sourcePaths"`
	TargetPath  string   `json:"targetPath"`
	Operation   string   `json:"operation"` // "copy" or "move"
}

// RenameRequest represents a request to rename a file/folder
type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"newName"`
}

// DeleteRequest represents a request to delete files/folders
type DeleteRequest struct {
	Paths []string `json:"paths"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}

// OpenFileRequest represents a request to open a file
type OpenFileRequest struct {
	Path string `json:"path"`
}

// DirectoryListRequest represents a request to list directory contents
type DirectoryListRequest struct {
	Path      string `json:"path"`
	ShowHidden bool  `json:"showHidden"`
} 