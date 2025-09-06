package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

// GetConfigString gets a config field value as a string
func (c *ConfigService) GetConfigString(key string) (string, error) {
	// Use reflection to get field value from config struct
	configValue := reflect.ValueOf(c.config).Elem()
	configType := reflect.TypeOf(c.config).Elem()

	// Find field by JSON tag
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == key {
			fieldValue := configValue.Field(i)
			if !fieldValue.IsValid() {
				return "", fmt.Errorf("field %s not found", key)
			}
			return fmt.Sprintf("%v", fieldValue.Interface()), nil
		}
	}

	return "", fmt.Errorf("config field %s not found", key)
}

// SetConfigString sets a config field value from a string
func (c *ConfigService) SetConfigString(key string, value string) error {
	// Use reflection to set field value in config struct
	configValue := reflect.ValueOf(c.config).Elem()
	configType := reflect.TypeOf(c.config).Elem()

	// Find field by JSON tag
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == key {
			fieldValue := configValue.Field(i)
			if !fieldValue.IsValid() || !fieldValue.CanSet() {
				return fmt.Errorf("field %s cannot be set", key)
			}
			
			// For now, assume all fields are strings (as per requirements)
			if fieldValue.Kind() == reflect.String {
				fieldValue.SetString(value)
				// Save the updated configuration
				return c.saveConfig()
			} else {
				return fmt.Errorf("field %s is not a string type", key)
			}
		}
	}

	return fmt.Errorf("config field %s not found", key)
}