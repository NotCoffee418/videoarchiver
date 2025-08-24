package playlist

import (
	"database/sql"
	"fmt"
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
	_, err := p.db.Exec("UPDATE playlists SET save_directory = ? WHERE id = ?", newDirectory, id)
	return err
}

func (p *PlaylistDB) DeletePlaylist(id int) error {
	fmt.Println("Deleting playlist", id)
	//_, err := p.db.Exec("DELETE FROM playlists WHERE id = ?", id)
	//return err
	return nil
}

func (p *PlaylistDB) UpdatePlaylistThumbnail(id int, thumbnailBase64 string) error {
	_, err := p.db.Exec("UPDATE playlists SET thumbnail_base64 = ? WHERE id = ?", thumbnailBase64, id)
	return err
}

func (p *PlaylistDB) GetPlaylists() ([]Playlist, error) {
	rows, err := p.db.Query("SELECT * FROM playlists ORDER BY added_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := make([]Playlist, 0)
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(
			&playlist.ID, &playlist.Name, &playlist.PlaylistGUID,
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
	playlistGUID string,
	directory string,
	format string,
) (bool, error) {

	// Check if playlist already exists
	var count int
	err := p.db.QueryRow(
		"SELECT COUNT(*) FROM playlists WHERE playlist_guid = ? AND save_directory = ? AND output_format = ?",
		playlistGUID, directory, format).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (p *PlaylistDB) AddPlaylist(
	name,
	playlistGUID,
	directory,
	format,
	thumbnail string,
) error {
	// Add new playlist
	_, err := p.db.Exec(
		`INSERT INTO playlists (name, playlist_guid, output_format, save_directory, thumbnail_base64, is_enabled)
		VALUES (?, ?, ?, ?, ?, 1)`,
		name, playlistGUID, format, directory, thumbnail,
	)
	return err
}
