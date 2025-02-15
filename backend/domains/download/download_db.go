package download

import (
	"database/sql"
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
			&download.ID, &download.PlaylistID, &download.VideoID,
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
			&download.ID, &download.PlaylistID, &download.VideoID,
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

func (d *DownloadDB) UpdateDownloadStatus(id int, newStatus int, failMessage *string) error {
	_, err := d.db.Exec(
		`UPDATE downloads
		 SET status = ?,
		 fail_message = COALESCE(?, fail_message),
		 last_attempt = CURRENT_TIMESTAMP,
		 attempt_count = attempt_count + 1
		 WHERE id = ?`,
		newStatus, failMessage, id,
	)
	return err
}
