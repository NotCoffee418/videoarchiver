package settings

import (
	"database/sql"
	"errors"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/logging"
)

// Settings must be created in a migration.
type SettingsService struct {
	db       *sql.DB
	handlers map[string]SettingHandler
	logger   *logging.LogService
}

func NewSettingsService(db *db.DatabaseService, logger *logging.LogService) *SettingsService {
	return &SettingsService{
		db:       db.GetDB(),
		handlers: GetSettingHandlers(),
		logger:   logger,
	}
}

// GetSettingString gets the raw value from the database
func (s *SettingsService) GetSettingString(key string) (string, error) {
	row := s.db.QueryRow("SELECT setting_value FROM settings WHERE setting_key = ?", key)
	var value string
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return "", errors.New("setting not found")
	}
	return value, err
}

// GetSettingBool gets a boolean setting from the database
func (s *SettingsService) GetSettingBool(key string) (bool, error) {
	value, err := s.GetSettingString(key)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// Set inserts or updates the setting value and triggers handlers
func (s *SettingsService) SetPreparsed(key string, value string) error {
	// Get old value for handler
	oldValue, _ := s.GetSettingString(key) // Ignore error - may not exist
	
	_, err := s.db.Exec(`
		UPDATE settings SET setting_value = ? WHERE setting_key = ?
	`, value, key)
	if err != nil {
		return err
	}

	// Call handler if one exists
	if handler, exists := s.handlers[key]; exists {
		if handlerErr := handler.HandleSettingChange(key, oldValue, value, s.logger); handlerErr != nil {
			// Log the error but don't fail the setting update
			if s.logger != nil {
				s.logger.Error("Setting handler failed for " + key + ": " + handlerErr.Error())
			}
		}
	}

	return nil
}
