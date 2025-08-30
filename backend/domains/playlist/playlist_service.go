package playlist

import (
	"fmt"
	"os"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/ytdlp"
	"videoarchiver/backend/imaging"

	"github.com/pkg/errors"
)

type PlaylistService struct {
	db              *PlaylistDB
	daemonSignalSvc *daemonsignal.DaemonSignalService
}

func NewPlaylistService(
	db *PlaylistDB,
	daemonSignalSvc *daemonsignal.DaemonSignalService,
) *PlaylistService {
	return &PlaylistService{
		db:              db,
		daemonSignalSvc: daemonSignalSvc,
	}
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
		plInfo.CleanUrl,
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
	err = p.db.AddPlaylist(
		plInfo.Title,
		plInfo.CleanUrl,
		directory,
		format,
		thumbnailBase64,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add playlist to database")
	}

	// Notify daemon of change
	err = p.daemonSignalSvc.TriggerChange()
	if err != nil {
		return err
	}

	return nil
}

func (p *PlaylistService) TryDeletePlaylist(id int) error {
	// Delete playlist from database (soft delete)
	err := p.db.DeletePlaylist(id)
	if err != nil {
		return errors.Wrap(err, "failed to delete playlist from database")
	}

	// Notify daemon of change
	err = p.daemonSignalSvc.TriggerChange()
	if err != nil {
		return err
	}

	return nil
}

func (p *PlaylistService) TryUpdatePlaylistDirectory(id int, newDirectory string) error {
	// Check if directory exists
	if _, err := os.Stat(newDirectory); newDirectory == "" || os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", newDirectory)
	}

	// Check if directory is writable
	if _, err := os.Stat(newDirectory); os.IsPermission(err) {
		return fmt.Errorf("no permission to write to directory: %s", newDirectory)
	}

	// Update playlist directory in database
	err := p.db.UpdatePlaylistDirectory(id, newDirectory)
	if err != nil {
		return errors.Wrap(err, "failed to update playlist directory in database")
	}

	// Notify daemon of change
	err = p.daemonSignalSvc.TriggerChange()
	if err != nil {
		return err
	}

	return nil
}
