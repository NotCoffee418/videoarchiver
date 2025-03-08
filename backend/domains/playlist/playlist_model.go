package playlist

import (
	"database/sql"
	"time"
)

type Playlist struct {
	ID              int            `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	URL             string         `json:"url" db:"url"`
	OutputFormat    string         `json:"output_format" db:"output_format"`
	SaveDirectory   string         `json:"save_directory" db:"save_directory"`
	ThumbnailBase64 sql.NullString `json:"thumbnail_base64,omitempty" db:"thumbnail_base64"`
	IsEnabled       bool           `json:"is_enabled" db:"is_enabled"`
	AddedAt         time.Time      `json:"added_at" db:"added_at"`
}
