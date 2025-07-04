package deals

import (
	"time"
)

// Deal represents a deal with its folder structure
type Deal struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	RootPath   string       `json:"rootPath"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	Status     DealStatus   `json:"status"`
	Metadata   DealMetadata `json:"metadata"`
	FolderTree *FolderNode  `json:"folderTree"`
}

// DealStatus represents the current status of a deal
type DealStatus string

const (
	DealStatusActive   DealStatus = "active"
	DealStatusPending  DealStatus = "pending"
	DealStatusClosed   DealStatus = "closed"
	DealStatusArchived DealStatus = "archived"
)

// DealMetadata contains additional information about a deal
type DealMetadata struct {
	Company         string                 `json:"company"`
	DealType        string                 `json:"dealType"`
	Value           float64                `json:"value"`
	Currency        string                 `json:"currency"`
	TargetCloseDate *time.Time             `json:"targetCloseDate,omitempty"`
	Tags            []string               `json:"tags"`
	CustomFields    map[string]interface{} `json:"customFields"`
}

// FolderNode represents a node in the folder hierarchy
type FolderNode struct {
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Type        FolderType    `json:"type"`
	IsDirectory bool          `json:"isDirectory"`
	Children    []*FolderNode `json:"children,omitempty"`
	FileCount   int           `json:"fileCount"`
	Size        int64         `json:"size"`
	ModifiedAt  time.Time     `json:"modifiedAt"`
}

// FolderType represents the type of folder in the deal structure
type FolderType string

const (
	FolderTypeRoot      FolderType = "root"
	FolderTypeLegal     FolderType = "legal"
	FolderTypeFinancial FolderType = "financial"
	FolderTypeGeneral   FolderType = "general"
	FolderTypeAnalysis  FolderType = "analysis"
	FolderTypeCustom    FolderType = "custom"
)

// FolderTemplate represents a template for folder structure
type FolderTemplate struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Structure   map[string]FolderTemplate `json:"structure"`
}

// DealStatistics represents statistics about a deal
type DealStatistics struct {
	TotalFiles         int            `json:"totalFiles"`
	TotalSize          int64          `json:"totalSize"`
	FilesByType        map[string]int `json:"filesByType"`
	FilesByFolder      map[string]int `json:"filesByFolder"`
	ProcessingProgress float64        `json:"processingProgress"`
	LastActivityAt     time.Time      `json:"lastActivityAt"`
	DocumentCategories map[string]int `json:"documentCategories"`
}

// FolderPermissions represents permissions for a folder
type FolderPermissions struct {
	Owner       string   `json:"owner"`
	ReadUsers   []string `json:"readUsers"`
	WriteUsers  []string `json:"writeUsers"`
	AdminUsers  []string `json:"adminUsers"`
	IsPublic    bool     `json:"isPublic"`
	InheritFrom string   `json:"inheritFrom,omitempty"`
}

// FolderOperationResult represents the result of a folder operation
type FolderOperationResult struct {
	Success       bool          `json:"success"`
	Operation     string        `json:"operation"`
	Path          string        `json:"path"`
	Message       string        `json:"message,omitempty"`
	Error         string        `json:"error,omitempty"`
	AffectedFiles int           `json:"affectedFiles"`
	Duration      time.Duration `json:"duration"`
}

// DealSearchCriteria represents search criteria for deals
type DealSearchCriteria struct {
	Name           string     `json:"name,omitempty"`
	Company        string     `json:"company,omitempty"`
	Status         DealStatus `json:"status,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	MinValue       float64    `json:"minValue,omitempty"`
	MaxValue       float64    `json:"maxValue,omitempty"`
	CreatedAfter   *time.Time `json:"createdAfter,omitempty"`
	CreatedBefore  *time.Time `json:"createdBefore,omitempty"`
	ModifiedAfter  *time.Time `json:"modifiedAfter,omitempty"`
	ModifiedBefore *time.Time `json:"modifiedBefore,omitempty"`
}

// DealActivity represents an activity on a deal
type DealActivity struct {
	ID          string                 `json:"id"`
	DealID      string                 `json:"dealId"`
	Type        ActivityType           `json:"type"`
	Description string                 `json:"description"`
	User        string                 `json:"user"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeFileAdded      ActivityType = "file_added"
	ActivityTypeFileRemoved    ActivityType = "file_removed"
	ActivityTypeFileModified   ActivityType = "file_modified"
	ActivityTypeFolderCreated  ActivityType = "folder_created"
	ActivityTypeFolderDeleted  ActivityType = "folder_deleted"
	ActivityTypeStatusChanged  ActivityType = "status_changed"
	ActivityTypeMetadataUpdate ActivityType = "metadata_update"
	ActivityTypeComment        ActivityType = "comment"
)
