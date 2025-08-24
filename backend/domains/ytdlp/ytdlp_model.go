package ytdlp

type YtdlpPlaylistInfo struct {
	Title        string
	ThumbnailURL string
	Entries      []YtdlpEntry
	PlaylistGUID string
}

type YtdlpEntry struct {
	Title string
	URL   string
}
