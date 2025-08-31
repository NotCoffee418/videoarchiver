package download

import (
	"database/sql"
	"path/filepath"
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
		statuses = append(statuses, StSuccess, StSuccessPlaylistRemoved, StSuccessDuplicate)
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

// CheckDuplicateByMD5 checks if a file with the same MD5 exists in the database
// It looks for files with similar names (base name with or without "-1", "-2", etc.)
func (d *DownloadDB) CheckDuplicateByMD5(md5 string, baseFilename string, playlistId int) (*Download, error) {
	if md5 == "" {
		return nil, nil
	}

	// Extract base name without extension for pattern matching
	baseName := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	ext := filepath.Ext(baseFilename)
	
	// Query for existing downloads with same MD5 and similar filename pattern
	query := `SELECT * FROM downloads 
	         WHERE md5 = ? AND playlist_id = ? AND status IN (?, ?) 
	         AND (output_filename = ? OR output_filename LIKE ? OR output_filename LIKE ?)`
	
	// Pattern matching for "-1", "-2", etc. variations
	pattern1 := baseName + "-%" + ext  // matches "filename-1.ext", "filename-10.ext" etc
	pattern2 := baseName + "-%"        // matches "filename-1", "filename-10" etc (if extension is missing)
	
	rows, err := d.db.Query(query, md5, playlistId, StSuccess, StSuccessDuplicate, baseFilename, pattern1, pattern2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	downloads, err := d.scanRows(rows)
	if err != nil {
		return nil, err
	}
	
	if len(downloads) > 0 {
		return &downloads[0], nil
	}
	return nil, nil
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

func (d *Download) SetSuccessDuplicate(dlDB *DownloadDB, outputFilename string, md5 string) error {
	d.Status = StSuccessDuplicate
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

func (d *DownloadDB) SetManualRetry(downloadId int) error {
	_, err := d.db.Exec(
		"UPDATE downloads SET status = ?, last_attempt = ? WHERE id = ?",
		StFailedManualRetry,
		time.Now().Unix(),
		downloadId,
	)
	return err
}

func (d *DownloadDB) RegisterAllFailedForRetryManual() error {
	_, err := d.db.Exec(
		"UPDATE downloads SET status = ? WHERE status = ?",
		StFailedManualRetry,
		StFailedGiveUp,
	)
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
