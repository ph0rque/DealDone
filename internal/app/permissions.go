package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PermissionChecker handles file system permission validation
type PermissionChecker struct{}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

// FolderPermissions represents the permissions for a folder
type FolderPermissions struct {
	Path       string `json:"path"`
	CanRead    bool   `json:"canRead"`
	CanWrite   bool   `json:"canWrite"`
	CanExecute bool   `json:"canExecute"`
	IsDir      bool   `json:"isDir"`
	Exists     bool   `json:"exists"`
}

// CheckFolderPermissions checks all permissions for a given folder
func (pc *PermissionChecker) CheckFolderPermissions(path string) (*FolderPermissions, error) {
	perms := &FolderPermissions{
		Path: path,
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		perms.Exists = false
		return perms, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	perms.Exists = true
	perms.IsDir = info.IsDir()

	// Check read permission by trying to list directory
	if perms.IsDir {
		_, err = os.ReadDir(path)
		perms.CanRead = err == nil
	} else {
		// For files, try to open for reading
		file, err := os.Open(path)
		if err == nil {
			file.Close()
			perms.CanRead = true
		}
	}

	// Check write permission
	perms.CanWrite = pc.checkWritePermission(path)

	// Check execute permission (mainly for directories)
	perms.CanExecute = pc.checkExecutePermission(path, info)

	return perms, nil
}

// checkWritePermission tests if we can write to a path
func (pc *PermissionChecker) checkWritePermission(path string) bool {
	// Generate unique test filename
	testFile := filepath.Join(path, fmt.Sprintf(".dealdone_perm_test_%d", time.Now().UnixNano()))

	// Try to create a test file
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()

	// Clean up
	os.Remove(testFile)
	return true
}

// checkExecutePermission checks if a path is executable/traversable
func (pc *PermissionChecker) checkExecutePermission(path string, info os.FileInfo) bool {
	// On Windows, directories are always traversable if readable
	if runtime.GOOS == "windows" {
		return info.IsDir()
	}

	// On Unix-like systems, check the execute bit
	mode := info.Mode()
	return mode&0111 != 0
}

// ValidateDealDonePermissions validates all required permissions for DealDone
func (pc *PermissionChecker) ValidateDealDonePermissions(rootPath string) error {
	// Check root folder
	rootPerms, err := pc.CheckFolderPermissions(rootPath)
	if err != nil {
		return fmt.Errorf("failed to check root permissions: %w", err)
	}

	if !rootPerms.Exists {
		// Try to check parent directory permissions
		parentPath := filepath.Dir(rootPath)
		parentPerms, err := pc.CheckFolderPermissions(parentPath)
		if err != nil {
			return fmt.Errorf("failed to check parent directory permissions: %w", err)
		}

		if !parentPerms.CanWrite {
			return fmt.Errorf("cannot create DealDone folder: no write permission in parent directory %s", parentPath)
		}
		// Parent is writable, so we can create the folder
		return nil
	}

	// Root exists, check permissions
	if !rootPerms.IsDir {
		return fmt.Errorf("path exists but is not a directory: %s", rootPath)
	}

	if !rootPerms.CanRead {
		return fmt.Errorf("cannot read from DealDone folder: %s", rootPath)
	}

	if !rootPerms.CanWrite {
		return fmt.Errorf("cannot write to DealDone folder: %s", rootPath)
	}

	return nil
}

// EnsurePermissions attempts to fix permission issues if possible
func (pc *PermissionChecker) EnsurePermissions(path string) error {
	// On Windows, permissions are usually not an issue
	if runtime.GOOS == "windows" {
		return nil
	}

	// Try to set reasonable permissions (rwxr-xr-x for directories)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		err = os.Chmod(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to set directory permissions: %w", err)
		}
	} else {
		// For files, set rw-r--r--
		err = os.Chmod(path, 0644)
		if err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil
}

// GetRecommendedPath returns OS-specific recommended paths for DealDone
func (pc *PermissionChecker) GetRecommendedPaths() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(home, "Desktop", "DealDone"),
			filepath.Join(home, "Documents", "DealDone"),
			filepath.Join(os.Getenv("USERPROFILE"), "DealDone"),
		}
	case "darwin": // macOS
		return []string{
			filepath.Join(home, "Desktop", "DealDone"),
			filepath.Join(home, "Documents", "DealDone"),
			filepath.Join(home, "DealDone"),
		}
	default: // Linux and others
		return []string{
			filepath.Join(home, "Desktop", "DealDone"),
			filepath.Join(home, "Documents", "DealDone"),
			filepath.Join(home, "DealDone"),
		}
	}
}

// IsPathSafe checks if a path is safe to use (not system critical)
func (pc *PermissionChecker) IsPathSafe(path string) error {
	// Normalize the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check for system directories
	dangerousPaths := pc.getDangerousPaths()

	for _, dangerous := range dangerousPaths {
		if absPath == dangerous || strings.HasPrefix(absPath, dangerous+string(filepath.Separator)) {
			return fmt.Errorf("cannot use system directory: %s", path)
		}
	}

	return nil
}

// getDangerousPaths returns a list of paths that should not be used
func (pc *PermissionChecker) getDangerousPaths() []string {
	paths := []string{
		"/",
		"/bin",
		"/sbin",
		"/usr",
		"/etc",
		"/var",
		"/tmp",
		"/dev",
		"/proc",
		"/sys",
	}

	if runtime.GOOS == "windows" {
		winPaths := []string{
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			os.Getenv("WINDIR"),
			os.Getenv("PROGRAMFILES"),
			os.Getenv("PROGRAMFILES(X86)"),
		}
		for _, p := range winPaths {
			if p != "" {
				paths = append(paths, p)
			}
		}
	}

	return paths
}
