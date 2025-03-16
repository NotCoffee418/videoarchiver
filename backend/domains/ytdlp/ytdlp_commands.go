package ytdlp

func getPlaylistInfo(url string) (string, error) {
	return runCommand("-J", url)
}
