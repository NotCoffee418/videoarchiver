package playlist

import (
	"database/sql"
	"videoarchiver/backend/domains/db"
)

type PlaylistDB struct {
	db *sql.DB
}

func NewPlaylistDB(dbService *db.DatabaseService) *PlaylistDB {
	return &PlaylistDB{db: dbService.GetDB()}
}

func (p *PlaylistDB) UpdatePlaylistName(id int, newName string) error {
	_, err := p.db.Exec("UPDATE playlists SET name = ? WHERE id = ?", newName, id)
	return err
}

func (p *PlaylistDB) UpdatePlaylistDirectory(id int, newDirectory string) error {
	_, err := p.db.Exec("UPDATE playlists SET save_directory = ? WHERE id = ? AND is_enabled = 1", newDirectory, id)
	return err
}

func (p *PlaylistDB) DeletePlaylist(id int) error {
	_, err := p.db.Exec("UPDATE playlists SET is_enabled = 0 WHERE id = ?", id)
	return err
}

func (p *PlaylistDB) UpdatePlaylistThumbnail(id int, thumbnailBase64 string) error {
	_, err := p.db.Exec("UPDATE playlists SET thumbnail_base64 = ? WHERE id = ?", thumbnailBase64, id)
	return err
}

func (p *PlaylistDB) GetActivePlaylists() ([]Playlist, error) {
	rows, err := p.db.Query("SELECT * FROM playlists WHERE is_enabled = 1 ORDER BY added_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := make([]Playlist, 0)
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(
			&playlist.ID, &playlist.Name, &playlist.URL,
			&playlist.OutputFormat, &playlist.SaveDirectory, &playlist.ThumbnailBase64,
			&playlist.IsEnabled, &playlist.AddedAt,
		)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (p *PlaylistDB) IsDuplicatePlaylistConfig(
	webpageUrl string,
	directory string,
	format string,
) (bool, error) {

	// Check if playlist already exists
	var count int
	err := p.db.QueryRow(
		"SELECT COUNT(*) FROM playlists WHERE url = ? AND save_directory = ? AND output_format = ? AND is_enabled = 1",
		webpageUrl, directory, format).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (p *PlaylistDB) AddPlaylist(
	name,
	webpageUrl,
	directory,
	format,
	thumbnail string,
) error {
	// Add new playlist
	_, err := p.db.Exec(
		`INSERT INTO playlists (name, url, output_format, save_directory, thumbnail_base64, is_enabled)
		VALUES (?, ?, ?, ?, ?, 1)`,
		name, webpageUrl, format, directory, thumbnail,
	)
	return err
}
