package fileregistry

import "database/sql"

// RegisteredFile represents a file that has been registered for duplicate detection
type RegisteredFile struct {
	ID           int            `json:"id" db:"id"`
	Filename     string         `json:"filename" db:"filename"`
	FilePath     string         `json:"file_path" db:"file_path"`
	MD5          string         `json:"md5" db:"md5"`
	RegisteredAt int64          `json:"registered_at" db:"registered_at"`
	KnownUrl     sql.NullString `json:"known_url" db:"known_url"`
}