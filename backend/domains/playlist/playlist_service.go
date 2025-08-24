package playlist

import (
	"fmt"
	"os"
	"videoarchiver/backend/domains/ytdlp"
	"videoarchiver/backend/imaging"
)

type PlaylistService struct {
	db *PlaylistDB
}

func NewPlaylistService(db *PlaylistDB) *PlaylistService {
	return &PlaylistService{db: db}
}

func (p *PlaylistService) TryAddNewPlaylist(url, directory, format string) error {
	// Check if directory exists
	if _, err := os.Stat(directory); directory == "" || os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}

	// Check if directory is writable
	if _, err := os.Stat(directory); os.IsPermission(err) {
		return fmt.Errorf("no permission to write to directory: %s", directory)
	}

	// Get playlist info
	plInfo, err := ytdlp.GetPlaylistInfoFlat(url)
	if err != nil {
		return err
	}

	// Get thumbnail
	thumbnailBase64, err := imaging.GetBase64Thumb(plInfo.ThumbnailURL)
	if err != nil {
		return err
	}

	// Duplicate check
	isDuplicate, err := p.db.IsDuplicatePlaylistConfig(
		plInfo.PlaylistGUID,
		directory,
		format,
	)
	if err != nil {
		return err
	}
	if isDuplicate {
		return fmt.Errorf("playlist is already listed with this configuration")
	}

	// Add playlist to database
	return p.db.AddPlaylist(
		plInfo.Title,
		plInfo.PlaylistGUID,
		directory,
		format,
		thumbnailBase64,
	)
}
