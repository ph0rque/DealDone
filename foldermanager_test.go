package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFolderManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dealdone-folder-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config service with test directory
	cs := &ConfigService{
		config: &Config{
			DealDoneRoot: filepath.Join(tempDir, "DealDone"),
			FirstRun:     true,
		},
	}

	fm := NewFolderManager(cs)

	t.Run("InitializeFolderStructure creates all folders", func(t *testing.T) {
		err := fm.InitializeFolderStructure()
		if err != nil {
			t.Fatalf("Failed to initialize folder structure: %v", err)
		}

		// Check root folder
		if _, err := os.Stat(cs.GetDealDoneRoot()); os.IsNotExist(err) {
			t.Error("DealDone root folder was not created")
		}

		// Check Templates folder
		if _, err := os.Stat(cs.GetTemplatesPath()); os.IsNotExist(err) {
			t.Error("Templates folder was not created")
		}

		// Check Deals folder
		if _, err := os.Stat(cs.GetDealsPath()); os.IsNotExist(err) {
			t.Error("Deals folder was not created")
		}
	})

	t.Run("ValidateFolderStructure validates existing structure", func(t *testing.T) {
		// Initialize first
		fm.InitializeFolderStructure()

		err := fm.ValidateFolderStructure()
		if err != nil {
			t.Errorf("ValidateFolderStructure failed on valid structure: %v", err)
		}

		// Remove Templates folder and validate again
		os.RemoveAll(cs.GetTemplatesPath())
		err = fm.ValidateFolderStructure()
		if err == nil {
			t.Error("ValidateFolderStructure should fail when Templates folder is missing")
		}
	})

	t.Run("CreateDealFolder creates deal with subfolders", func(t *testing.T) {
		// Initialize structure first
		fm.InitializeFolderStructure()

		dealName := "TestDeal"
		dealPath, err := fm.CreateDealFolder(dealName)
		if err != nil {
			t.Fatalf("Failed to create deal folder: %v", err)
		}

		// Check deal folder exists
		if _, err := os.Stat(dealPath); os.IsNotExist(err) {
			t.Error("Deal folder was not created")
		}

		// Check subfolders
		subfolders := []string{"legal", "financial", "general", "analysis"}
		for _, subfolder := range subfolders {
			subPath := filepath.Join(dealPath, subfolder)
			if _, err := os.Stat(subPath); os.IsNotExist(err) {
				t.Errorf("Subfolder %s was not created", subfolder)
			}
		}
	})

	t.Run("CreateDealFolder rejects empty name", func(t *testing.T) {
		_, err := fm.CreateDealFolder("")
		if err == nil {
			t.Error("CreateDealFolder should reject empty deal name")
		}
	})

	t.Run("ListDeals returns created deals", func(t *testing.T) {
		// Create a fresh temp directory for this test
		testTempDir, err := os.MkdirTemp("", "dealdone-listdeals-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(testTempDir)

		// Create fresh config and folder manager
		testCs := &ConfigService{
			config: &Config{
				DealDoneRoot: filepath.Join(testTempDir, "DealDone"),
				FirstRun:     true,
			},
		}
		testFm := NewFolderManager(testCs)

		// Initialize and create some deals
		testFm.InitializeFolderStructure()
		testFm.CreateDealFolder("Deal1")
		testFm.CreateDealFolder("Deal2")
		testFm.CreateDealFolder("Deal3")

		deals, err := testFm.ListDeals()
		if err != nil {
			t.Fatalf("Failed to list deals: %v", err)
		}

		if len(deals) != 3 {
			t.Errorf("Expected 3 deals, got %d", len(deals))
		}

		// Check specific deals exist
		dealMap := make(map[string]bool)
		for _, deal := range deals {
			dealMap[deal] = true
		}

		expectedDeals := []string{"Deal1", "Deal2", "Deal3"}
		for _, expected := range expectedDeals {
			if !dealMap[expected] {
				t.Errorf("Expected deal %s not found", expected)
			}
		}
	})

	t.Run("DealExists correctly identifies existing deals", func(t *testing.T) {
		fm.InitializeFolderStructure()
		fm.CreateDealFolder("ExistingDeal")

		if !fm.DealExists("ExistingDeal") {
			t.Error("DealExists should return true for existing deal")
		}

		if fm.DealExists("NonExistentDeal") {
			t.Error("DealExists should return false for non-existent deal")
		}
	})

	t.Run("GetDealPath returns correct path", func(t *testing.T) {
		dealName := "TestDeal"
		expectedPath := filepath.Join(cs.GetDealsPath(), dealName)

		if fm.GetDealPath(dealName) != expectedPath {
			t.Errorf("GetDealPath returned incorrect path")
		}
	})

	t.Run("GetDealSubfolderPath returns correct path", func(t *testing.T) {
		dealName := "TestDeal"
		subfolder := "legal"
		expectedPath := filepath.Join(cs.GetDealsPath(), dealName, subfolder)

		if fm.GetDealSubfolderPath(dealName, subfolder) != expectedPath {
			t.Errorf("GetDealSubfolderPath returned incorrect path")
		}
	})
}
