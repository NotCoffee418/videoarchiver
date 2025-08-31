package config

import (
	"path/filepath"
	"testing"
)

func TestConfigServiceBasic(t *testing.T) {
	// Test creating a new config service
	configSvc, err := NewConfigService()
	if err != nil {
		t.Fatalf("Failed to create config service: %v", err)
	}

	// Test that config was created with default database path
	config := configSvc.GetConfig()
	if config.DatabasePath == "" {
		t.Error("Expected default database path to be set")
	}

	// Test getting database path
	dbPath, err := configSvc.GetDatabasePath()
	if err != nil {
		t.Fatalf("Failed to get database path: %v", err)
	}
	if dbPath == "" {
		t.Error("Expected database path to be returned")
	}

	// Test that the path ends with db.sqlite (default behavior)
	if filepath.Base(dbPath) != "db.sqlite" {
		t.Errorf("Expected database path to end with 'db.sqlite', got: %s", filepath.Base(dbPath))
	}
}