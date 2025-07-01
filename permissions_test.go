package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPermissionChecker(t *testing.T) {
	pc := NewPermissionChecker()

	t.Run("CheckFolderPermissions on existing directory", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "dealdone-perm-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		perms, err := pc.CheckFolderPermissions(tempDir)
		if err != nil {
			t.Fatalf("Failed to check permissions: %v", err)
		}

		if !perms.Exists {
			t.Error("Expected directory to exist")
		}

		if !perms.IsDir {
			t.Error("Expected path to be a directory")
		}

		if !perms.CanRead {
			t.Error("Expected to have read permission")
		}

		if !perms.CanWrite {
			t.Error("Expected to have write permission")
		}
	})

	t.Run("CheckFolderPermissions on non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(os.TempDir(), "this-should-not-exist-dealdone-test")
		os.RemoveAll(nonExistent) // Make sure it doesn't exist

		perms, err := pc.CheckFolderPermissions(nonExistent)
		if err != nil {
			t.Fatalf("Failed to check permissions: %v", err)
		}

		if perms.Exists {
			t.Error("Expected path to not exist")
		}
	})

	t.Run("checkWritePermission validates correctly", func(t *testing.T) {
		// Test writable directory
		tempDir, err := os.MkdirTemp("", "dealdone-write-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		if !pc.checkWritePermission(tempDir) {
			t.Error("Expected temp directory to be writable")
		}

		// Test non-writable directory (skip on Windows)
		if runtime.GOOS != "windows" {
			// Make directory read-only
			os.Chmod(tempDir, 0555)
			if pc.checkWritePermission(tempDir) {
				t.Error("Expected read-only directory to not be writable")
			}
			// Restore permissions for cleanup
			os.Chmod(tempDir, 0755)
		}
	})

	t.Run("ValidateDealDonePermissions validates correctly", func(t *testing.T) {
		// Test with valid directory
		tempDir, err := os.MkdirTemp("", "dealdone-validate-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		err = pc.ValidateDealDonePermissions(tempDir)
		if err != nil {
			t.Errorf("ValidateDealDonePermissions failed on valid directory: %v", err)
		}

		// Test with non-existent path but writable parent
		nonExistent := filepath.Join(tempDir, "new-dealdone-folder")
		err = pc.ValidateDealDonePermissions(nonExistent)
		if err != nil {
			t.Errorf("ValidateDealDonePermissions should pass for creatable path: %v", err)
		}

		// Test with file instead of directory
		testFile := filepath.Join(tempDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)
		err = pc.ValidateDealDonePermissions(testFile)
		if err == nil {
			t.Error("ValidateDealDonePermissions should fail for file path")
		}
	})

	t.Run("IsPathSafe rejects dangerous paths", func(t *testing.T) {
		dangerousPaths := []string{
			"/",
			"/etc",
			"/bin",
			"/usr/bin",
		}

		if runtime.GOOS == "windows" {
			dangerousPaths = append(dangerousPaths,
				"C:\\Windows",
				"C:\\Windows\\System32",
				"C:\\Program Files",
			)
		}

		for _, path := range dangerousPaths {
			err := pc.IsPathSafe(path)
			if err == nil {
				t.Errorf("IsPathSafe should reject dangerous path: %s", path)
			}
		}
	})

	t.Run("IsPathSafe accepts safe paths", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		safePaths := []string{
			filepath.Join(home, "Desktop", "DealDone"),
			filepath.Join(home, "Documents", "DealDone"),
			filepath.Join(home, "DealDone"),
		}

		for _, path := range safePaths {
			err := pc.IsPathSafe(path)
			if err != nil {
				t.Errorf("IsPathSafe should accept safe path %s: %v", path, err)
			}
		}
	})

	t.Run("GetRecommendedPaths returns valid paths", func(t *testing.T) {
		paths := pc.GetRecommendedPaths()

		if len(paths) == 0 {
			t.Error("GetRecommendedPaths should return at least one path")
		}

		// All recommended paths should be safe
		for _, path := range paths {
			err := pc.IsPathSafe(path)
			if err != nil {
				t.Errorf("Recommended path should be safe: %s, error: %v", path, err)
			}
		}
	})

	t.Run("EnsurePermissions sets correct permissions", func(t *testing.T) {
		// Skip on Windows where this is a no-op
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		tempDir, err := os.MkdirTemp("", "dealdone-ensure-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0600)

		// Ensure permissions on directory
		err = pc.EnsurePermissions(tempDir)
		if err != nil {
			t.Errorf("EnsurePermissions failed on directory: %v", err)
		}

		// Check directory has correct permissions
		info, _ := os.Stat(tempDir)
		mode := info.Mode().Perm()
		if mode != 0755 {
			t.Errorf("Expected directory permissions 0755, got %o", mode)
		}

		// Ensure permissions on file
		err = pc.EnsurePermissions(testFile)
		if err != nil {
			t.Errorf("EnsurePermissions failed on file: %v", err)
		}

		// Check file has correct permissions
		info, _ = os.Stat(testFile)
		mode = info.Mode().Perm()
		if mode != 0644 {
			t.Errorf("Expected file permissions 0644, got %o", mode)
		}
	})
}
