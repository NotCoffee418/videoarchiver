package download

import (
	"database/sql"
)

const (
	MaxRetryCount = 5
)

type Download struct {
	ID               int            `json:"id" db:"id"`
	PlaylistID       int            `json:"playlist_id" db:"playlist_id"`
	VideoID          string         `json:"video_id" db:"video_id"`
	Status           Status         `json:"status" db:"status"`
	FormatDownloaded string         `json:"format_downloaded" db:"format_downloaded"`
	MD5              sql.NullString `json:"md5,omitempty" db:"md5"`
	LastAttempt      int64          `json:"last_attempt" db:"last_attempt"`
	FailMessage      sql.NullString `json:"fail_message,omitempty" db:"fail_message"`
	AttemptCount     int            `json:"attempt_count" db:"attempt_count"`
}

// Creates new instance of Download without an ID or attempt info
func NewDownload(
	playlistId int,
	videoId string,
	formatDownloaded string,
) *Download {
	return &Download{
		PlaylistID:       playlistId,
		VideoID:          videoId,
		Status:           StUndownloaded,
		FormatDownloaded: formatDownloaded,
		MD5:              sql.NullString{String: "", Valid: false},
		LastAttempt:      0,
		FailMessage:      sql.NullString{String: "", Valid: false},
		AttemptCount:     0,
	}
}

type Status int

const (
	StUndownloaded = 0 // 0 Should not be in DB
	StSuccess      = 1
	StFailedRetry  = 2
	StFailedGiveUp = 3
)
