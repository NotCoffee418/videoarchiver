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
	rows, err := d.db.Query(`SELECT 
		id, playlist_id, url, status, format_downloaded, md5, output_filename, 
		last_attempt, fail_message, attempt_count, NULL as save_directory, thumbnail_base64
		FROM downloads ORDER BY last_attempt DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanRows(rows)
}

func (d *DownloadDB) GetDownloadsForPlaylist(playlistId int) ([]Download, error) {
	rows, err := d.db.Query(`SELECT 
		id, playlist_id, url, status, format_downloaded, md5, output_filename, 
		last_attempt, fail_message, attempt_count, NULL as save_directory, thumbnail_base64
		FROM downloads WHERE playlist_id = ?`, playlistId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanRows(rows)
}

func (d *DownloadDB) GetDownloadHistoryPage(offset, limit int, showSuccess, showFailed, showDuplicate bool) ([]Download, error) {
	var statuses []int
	if showSuccess {
		statuses = append(statuses, StSuccess, StSuccessPlaylistRemoved)
	}
	if showFailed {
		statuses = append(statuses, StFailedAutoRetry, StFailedManualRetry, StFailedGiveUp, StFailedPlaylistRemoved)
	}
	if showDuplicate {
		statuses = append(statuses, StSuccessDuplicate)
	}

	// Return empty if no filters selected
	if len(statuses) == 0 {
		return []Download{}, nil
	}

	query := `SELECT 
		d.id, d.playlist_id, d.url, d.status, d.format_downloaded, d.md5, d.output_filename, 
		d.last_attempt, d.fail_message, d.attempt_count, p.save_directory, d.thumbnail_base64
		FROM downloads d 
		LEFT JOIN playlists p ON d.playlist_id = p.id 
		WHERE d.status IN (` + strings.Repeat("?,", len(statuses)-1) + `?) 
		ORDER BY d.last_attempt DESC 
		LIMIT ? OFFSET ?`

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
			&download.LastAttempt, &download.FailMessage, &download.AttemptCount, &download.SaveDirectory,
			&download.ThumbnailBase64,
		)
		if err != nil {
			return nil, err
		}

		// Compute FullPath using proper path joining
		if download.OutputFilename.Valid && download.OutputFilename.String != "" &&
			download.SaveDirectory.Valid && download.SaveDirectory.String != "" {
			fullPath := filepath.Join(download.SaveDirectory.String, download.OutputFilename.String)
			download.FullPath = sql.NullString{String: fullPath, Valid: true}
		} else {
			download.FullPath = sql.NullString{String: "", Valid: false}
		}

		downloads = append(downloads, download)
	}
	return downloads, nil
}

func (d *Download) SetSuccess(
	dlDB *DownloadDB,
	outputFilename string,
	md5 string,
	thumbnailBase64 string,
) error {
	d.Status = StSuccess
	d.MD5 = sql.NullString{String: md5, Valid: true}
	d.OutputFilename = sql.NullString{String: outputFilename, Valid: true}
	d.FailMessage = sql.NullString{String: "", Valid: false}
	d.ThumbnailBase64 = sql.NullString{String: thumbnailBase64, Valid: thumbnailBase64 != ""}
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

func (d *Download) SetSuccessDuplicate(
	dlDB *DownloadDB, outputFilename string, md5 string, thumbnailBase64 string,
) error {
	d.Status = StSuccessDuplicate
	d.MD5 = sql.NullString{String: md5, Valid: true}
	d.OutputFilename = sql.NullString{String: outputFilename, Valid: true}
	d.FailMessage = sql.NullString{String: "", Valid: false}
	d.ThumbnailBase64 = sql.NullString{String: thumbnailBase64, Valid: thumbnailBase64 != ""}
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

func (d *Download) SetFail(
	dlDB *DownloadDB,
	failMessage string,
) error {
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
		`INSERT INTO downloads (playlist_id, url, status, format_downloaded, md5, output_filename, last_attempt, fail_message, attempt_count, thumbnail_base64)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.OutputFilename, d.LastAttempt, d.FailMessage, d.AttemptCount, d.ThumbnailBase64,
	)
	return err
}

func (d *Download) updateDownload(dlDB *DownloadDB) error {
	_, err := dlDB.db.Exec(
		`UPDATE downloads SET playlist_id = ?, url = ?, status = ?, format_downloaded = ?, md5 = ?, output_filename = ?, last_attempt = ?, fail_message = ?, attempt_count = ?, thumbnail_base64 = ? WHERE id = ?`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.OutputFilename, d.LastAttempt, d.FailMessage, d.AttemptCount, d.ThumbnailBase64, d.ID)
	return err
}

// CheckForDuplicateInDownloads checks if any existing download has the same MD5
// Returns: exists (bool), id (int), error
func (d *DownloadDB) CheckForDuplicateInDownloads(fileMD5 string, ignoredOwnId int) (bool, int, error) {
	var id int

	// Query downloads table for matching MD5
	err := d.db.QueryRow(
		"SELECT id FROM downloads WHERE md5 = ? AND id != ? LIMIT 1",
		fileMD5, ignoredOwnId,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			// No duplicate found
			return false, 0, nil
		}
		// Actual error occurred
		return false, 0, err
	}

	// Duplicate found
	return true, id, nil
}

// Removes excessive error message parts
func cleanDownloadFailMessage(msg string) string {
	rx := regexp.MustCompile(`(` + ErrDownloadErrorBase + `)?(exit status \d+?: )?(ERROR: )?`)
	return rx.ReplaceAllString(msg, "")
}
