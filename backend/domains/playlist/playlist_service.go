package playlist

import (
	"errors"
	"time"
)

type PlaylistService struct {
	db *PlaylistDB
}

func NewPlaylistService(db *PlaylistDB) *PlaylistService {
	return &PlaylistService{db: db}
}

func (p *PlaylistService) TryAddNewPlaylist(url, directory, format string) error {
	//todo: implement
	// validate and get full data with ytdlp
	// use playlist DB service
	time.Sleep(2 * time.Second)
	return errors.New("not implemented")
}
