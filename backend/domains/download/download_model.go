package download

import "time"

type Download struct {
	ID               int       `json:"id" db:"id"`
	PlaylistID       int       `json:"playlist_id" db:"playlist_id"`
	VideoID          string    `json:"video_id" db:"video_id"`
	Status           int       `json:"status" db:"status"`
	FormatDownloaded string    `json:"format_downloaded" db:"format_downloaded"`
	MD5              *string   `json:"md5,omitempty" db:"md5"`
	LastAttempt      time.Time `json:"last_attempt" db:"last_attempt"`
	FailMessage      *string   `json:"fail_message,omitempty" db:"fail_message"`
	AttemptCount     int       `json:"attempt_count" db:"attempt_count"`
}
