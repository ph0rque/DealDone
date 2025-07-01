package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dealdone-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config directory for testing
	originalGetConfigDir := getConfigDir
	getConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		getConfigDir = originalGetConfigDir
	}()

	t.Run("NewConfigService creates default config on first run", func(t *testing.T) {
		cs, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}

		if !cs.IsFirstRun() {
			t.Error("Expected first run to be true")
		}

		if cs.GetDealDoneRoot() == "" {
			t.Error("Expected DealDoneRoot to be set")
		}
	})

	t.Run("SetDealDoneRoot updates and persists", func(t *testing.T) {
		cs, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}

		newRoot := "/custom/path/DealDone"
		err = cs.SetDealDoneRoot(newRoot)
		if err != nil {
			t.Fatalf("Failed to set DealDone root: %v", err)
		}

		if cs.GetDealDoneRoot() != newRoot {
			t.Errorf("Expected root to be %s, got %s", newRoot, cs.GetDealDoneRoot())
		}

		// Create new service to test persistence
		cs2, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create second config service: %v", err)
		}

		if cs2.GetDealDoneRoot() != newRoot {
			t.Errorf("Expected persisted root to be %s, got %s", newRoot, cs2.GetDealDoneRoot())
		}
	})

	t.Run("GetTemplatesPath returns correct path", func(t *testing.T) {
		cs, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}

		root := "/test/DealDone"
		cs.SetDealDoneRoot(root)

		expected := filepath.Join(root, "Templates")
		if cs.GetTemplatesPath() != expected {
			t.Errorf("Expected templates path to be %s, got %s", expected, cs.GetTemplatesPath())
		}
	})

	t.Run("GetDealsPath returns correct path", func(t *testing.T) {
		cs, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}

		root := "/test/DealDone"
		cs.SetDealDoneRoot(root)

		expected := filepath.Join(root, "Deals")
		if cs.GetDealsPath() != expected {
			t.Errorf("Expected deals path to be %s, got %s", expected, cs.GetDealsPath())
		}
	})

	t.Run("SetFirstRun updates and persists", func(t *testing.T) {
		cs, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}

		err = cs.SetFirstRun(false)
		if err != nil {
			t.Fatalf("Failed to set first run: %v", err)
		}

		if cs.IsFirstRun() {
			t.Error("Expected first run to be false")
		}

		// Create new service to test persistence
		cs2, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create second config service: %v", err)
		}

		if cs2.IsFirstRun() {
			t.Error("Expected persisted first run to be false")
		}
	})
}
