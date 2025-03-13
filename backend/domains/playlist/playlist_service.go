package playlist

type PlaylistService struct {
	db *PlaylistDB
}

func NewPlaylistService(db *PlaylistDB) *PlaylistService {
	return &PlaylistService{db: db}
}

func (p *PlaylistService) AddNewPlaylist(url, directory, format string) error {
	//todo: implement
	return nil
}
