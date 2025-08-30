package download

import (
	"database/sql"
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

	var downloads []Download
	for rows.Next() {
		var download Download
		err := rows.Scan(
			&download.ID, &download.PlaylistID, &download.Url,
			&download.Status, &download.FormatDownloaded, &download.MD5,
			&download.LastAttempt, &download.FailMessage, &download.AttemptCount,
		)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, download)
	}
	return downloads, nil
}

func (d *DownloadDB) GetDownloadsForPlaylist(playlistId int) ([]Download, error) {
	rows, err := d.db.Query("SELECT * FROM downloads WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var download Download
		err := rows.Scan(
			&download.ID, &download.PlaylistID, &download.Url,
			&download.Status, &download.FormatDownloaded, &download.MD5,
			&download.LastAttempt, &download.FailMessage, &download.AttemptCount,
		)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, download)
	}
	return downloads, nil
}

func (d *Download) SetSuccess(dlDB *DownloadDB, md5 string) error {
	d.Status = StSuccess
	d.MD5 = sql.NullString{String: md5, Valid: true}
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
		d.Status = StFailedRetry
	}

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
		`INSERT INTO downloads (playlist_id, video_id, status, format_downloaded, md5, last_attempt, fail_message, attempt_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.LastAttempt, d.FailMessage, d.AttemptCount,
	)
	return err
}

func (d *Download) updateDownload(dlDB *DownloadDB) error {
	_, err := dlDB.db.Exec(
		`UPDATE downloads SET playlist_id = ?, video_id = ?, status = ?, format_downloaded = ?, md5 = ?, last_attempt = ?, fail_message = ?, attempt_count = ? WHERE id = ?`,
		d.PlaylistID, d.Url, d.Status, d.FormatDownloaded, d.MD5, d.LastAttempt, d.FailMessage, d.AttemptCount, d.ID)
	return err
}
