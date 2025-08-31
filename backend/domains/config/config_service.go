package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"videoarchiver/backend/domains/pathing"
)

// Config represents the application configuration
type Config struct {
	DatabasePath string `json:"database_path"`
}

// ConfigService handles loading and saving configuration
type ConfigService struct {
	config     *Config
	configPath string
}

// NewConfigService creates a new configuration service
func NewConfigService() (*ConfigService, error) {
	configPath, err := pathing.GetWorkingFile("config.json")
	if err != nil {
		return nil, err
	}

	service := &ConfigService{
		configPath: configPath,
	}

	err = service.loadConfig()
	if err != nil {
		return nil, err
	}

	return service, nil
}

// GetConfig returns the current configuration
func (c *ConfigService) GetConfig() *Config {
	return c.config
}

// GetDatabasePath returns the configured database path
func (c *ConfigService) GetDatabasePath() (string, error) {
	if c.config.DatabasePath == "" {
		// Return default path if not configured
		return getDefaultDatabasePath()
	}
	
	// If the configured path is relative, make it absolute using GetWorkingFile
	if !filepath.IsAbs(c.config.DatabasePath) {
		return pathing.GetWorkingFile(c.config.DatabasePath)
	}
	
	return c.config.DatabasePath, nil
}

// loadConfig loads configuration from file or creates default if it doesn't exist
func (c *ConfigService) loadConfig() error {
	// Check if config file exists
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		// Create default configuration
		defaultPath, err := getDefaultDatabasePath()
		if err != nil {
			return err
		}
		
		c.config = &Config{
			DatabasePath: defaultPath,
		}
		
		// Save the default configuration
		return c.saveConfig()
	}

	// Read existing configuration
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	c.config = &Config{}
	err = json.Unmarshal(data, c.config)
	if err != nil {
		return err
	}

	// If database path is empty, set to default
	if c.config.DatabasePath == "" {
		defaultPath, err := getDefaultDatabasePath()
		if err != nil {
			return err
		}
		c.config.DatabasePath = defaultPath
		// Save the updated configuration
		return c.saveConfig()
	}

	return nil
}

// saveConfig saves the current configuration to file
func (c *ConfigService) saveConfig() error {
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(c.configPath, data, 0644)
}

// getDefaultDatabasePath returns the default database path (current behavior)
func getDefaultDatabasePath() (string, error) {
	return pathing.GetWorkingFile("db.sqlite")
}