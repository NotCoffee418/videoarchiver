package ytdlp

type YtdlpPlaylistInfo struct {
	Title        string
	ThumbnailURL string
	CleanUrl     string
	Entries      []YtdlpEntry
}

type YtdlpEntry struct {
	Title string
	URL   string
}
