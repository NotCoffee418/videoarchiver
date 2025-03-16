package ytdlp

type PlaylistInfo struct {
	Title        string
	ThumbnailURL string
	CleanUrl     string
	Entries      []Entry
}

type Entry struct {
	Title string
	URL   string
}
