package download

import (
	"database/sql"
)

const (
	// Amount of retries before status is changed to StGiveUp
	MaxRetryCount = 5
)

type Download struct {
	ID               int            `json:"id" db:"id"`
	PlaylistID       int            `json:"playlist_id" db:"playlist_id"`
	Url              string         `json:"url" db:"url"`
	Status           Status         `json:"status" db:"status"`
	FormatDownloaded string         `json:"format_downloaded" db:"format_downloaded"`
	MD5              sql.NullString `json:"md5,omitempty" db:"md5"`
	OutputFilename   sql.NullString `json:"output_filename,omitempty" db:"output_filename"`
	LastAttempt      int64          `json:"last_attempt" db:"last_attempt"`
	FailMessage      sql.NullString `json:"fail_message,omitempty" db:"fail_message"`
	AttemptCount     int            `json:"attempt_count" db:"attempt_count"`
}

// Creates new instance of Download without an ID or attempt info
func NewDownload(
	playlistId int,
	cleanUrl string,
	formatDownloaded string,
) *Download {
	return &Download{
		PlaylistID:       playlistId,
		Url:              cleanUrl,
		Status:           StUndownloaded,
		FormatDownloaded: formatDownloaded,
		MD5:              sql.NullString{String: "", Valid: false},
		OutputFilename:   sql.NullString{String: "", Valid: false},
		LastAttempt:      0,
		FailMessage:      sql.NullString{String: "", Valid: false},
		AttemptCount:     0,
	}
}

type Status int

const (
	StUndownloaded      = 0 // 0 Should not be in DB
	StSuccess           = 1
	StFailedAutoRetry   = 2
	StFailedManualRetry = 3
	StFailedGiveUp      = 4
)
