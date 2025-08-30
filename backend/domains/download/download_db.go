package download

import (
	"database/sql"
	"regexp"
	"strings"
	"time"
	"videoarchiver/backend/domains/db"
)

type DownloadDB struct {
	db *sql.DB
}

func NewDownloadDB(dbService *db.DatabaseService) *DownloadDB {
	return &DownloadDB{db: dbService.GetDB()}
}

func (d *DownloadDB) GetAllDownloads(limit int) ([]Download, error) {
	rows, err := d.db.Query("SELECT * FROM downloads ORDER BY last_attempt DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanRows(rows)
}

func (d *DownloadDB) GetDownloadsForPlaylist(playlistId int) ([]Download, error) {
	rows, err := d.db.Query("SELECT * FROM downloads WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanRows(rows)
}

func (d *DownloadDB) GetDownloadHistoryPage(offset, limit int, showSuccess, showFailed bool) ([]Download, error) {
	var statuses []int
	if showSuccess {
		statuses = append(statuses, StSuccess, StSuccessPlaylistRemoved)
	}
	if showFailed {
		statuses = append(statuses, StFailedAutoRetry, StFailedManualRetry, StFailedGiveUp, StFailedPlaylistRemoved)
	}

	// Return empty if no filters selected
	if len(statuses) == 0 {
		return []Download{}, nil
	}

	query := "SELECT * FROM downloads WHERE status IN (?" + strings.Repeat(",?", len(statuses)-1) + ") ORDER BY last_attempt DESC LIMIT ? OFFSET ?"

	// Convert statuses to interface{} for query args
	args := make([]interface{}, len(statuses)+2)
	for i, status := range statuses {
		args[i] = status
	}
	args[len(statuses)] = limit
	args[len(statuses)+1] = offset

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanRows(rows)
}

func (d *DownloadDB) scanRows(rows *sql.Rows) ([]Download, error) {
	var downloads []Download
	for rows.Next() {
		var download Download
		err := rows.Scan(
			&download.ID, &download.PlaylistID, &download.Url,
			&download.Status, &download.FormatDownloaded, &download.MD5, &download.OutputFilename,
			&download.LastAttempt, &download.FailMessage, &download.AttemptCount,
		)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, download)
	}
	return downloads, nil
}

func (d *Download) SetSuccess(dlDB *DownloadDB, outputFilename string, md5 string) error {
	d.Status = StSuccess
	d.MD5 = sql.NullString{String: md5, Valid: true}
	d.OutputFilename = sql.NullString{String: outputFilename, Valid: true}
	d.FailMessage = sql.NullString{String: "", Valid: false}
	d.AttemptCount += 1
	d.LastAttempt = time.Now().Unix()

	var err error
	if d.ID == 0 {
		err = d.insertDownload(dlDB)
	} else {
		err = d.updateDownload(dlDB)
	}
	return err
}

func (d *Download) SetFail(dlDB *DownloadDB, failMessage string) error {
	d.AttemptCount += 1
	if d.AttemptCount > MaxRetryCount {
		d.Status = StFailedGiveUp
	} else {
		d.Status = StFailedAutoRetry
	}

	// Clean fail message
	failMessage = cleanDownloadFailMessage(failMessage)
	d.FailMessage = sql.NullString{String: failMessage, Valid: true}

	d.LastAttempt = time.Now().Unix()

	var err error
	if d.ID == 0 {
		err = d.insertDownload(dlDB)
	} else {
		err = d.updateDownload(dlDB)
	}
	return err
}

func (d *Download) insertDownload(dlDB *DownloadDB) error {
	_, err := dlDB.db.Exec(
		`INSERT INTO downloads (playlist_id, url, status, format_downloaded, md5, output_filename, last_attempt, fail_message, attempt_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.OutputFilename, d.LastAttempt, d.FailMessage, d.AttemptCount,
	)
	return err
}

func (d *Download) updateDownload(dlDB *DownloadDB) error {
	_, err := dlDB.db.Exec(
		`UPDATE downloads SET playlist_id = ?, url = ?, status = ?, format_downloaded = ?, md5 = ?, last_attempt = ?, fail_message = ?, attempt_count = ? WHERE id = ?`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.LastAttempt, d.FailMessage, d.AttemptCount, d.ID)
	return err
}

// Removes excessive error message parts
func cleanDownloadFailMessage(msg string) string {
	rx := regexp.MustCompile(`(` + ErrDownloadErrorBase + `)?(exit status \d+?: )?(ERROR: )?`)
	return rx.ReplaceAllString(msg, "")
}
