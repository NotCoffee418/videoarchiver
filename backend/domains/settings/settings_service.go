package settings

import (
	"database/sql"
	"errors"
	"videoarchiver/backend/domains/db"
)

// Settings must be created in a migration.
type SettingsService struct {
	db *sql.DB
}

func NewSettingsService(db *db.DatabaseService) *SettingsService {
	return &SettingsService{db: db.GetDB()}
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

// Set inserts or updates the setting value
func (s *SettingsService) SetPreparsed(key string, value string) error {
	_, err := s.db.Exec(`
		UPDATE settings SET setting_value = ? WHERE setting_key = ?
	`, value, key)
	if err != nil {
		return err
	}

	return nil
}
