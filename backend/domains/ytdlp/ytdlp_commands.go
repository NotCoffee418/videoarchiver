package ytdlp

import (
	"encoding/json"
	"errors"
	"fmt"
	"videoarchiver/backend/domains/settings"
)

// Get minimal playlist info
func GetPlaylistInfoFlat(url string) (*YtdlpPlaylistInfo, error) {
	raw, err := runCommand("--no-warnings", "--flat-playlist", "--yes-playlist", "-J", url)
	if err != nil {
		return nil, err
	}

	// Prepare result object
	result := &YtdlpPlaylistInfo{
		Title:        "",
		ThumbnailURL: "",
		Entries:      make([]YtdlpEntry, 0),
	}

	// Parse json
	var data map[string]interface{}
	err = json.Unmarshal([]byte(raw), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Confirm that the data is a playlist
	// Get URL
	urlType, ok := data["_type"].(string)
	if !ok || urlType != "playlist" {
		return nil, errors.New("invalid or private playlist url")
	}

	// Get playlist name (type assertion with check)
	playlistName, ok := data["title"].(string)
	if !ok {
		return nil, errors.New("missing or invalid playlist title")
	}
	result.Title = playlistName

	// Get clean URL
	cleanUrl, ok := data["webpage_url"].(string)
	if !ok {
		return nil, errors.New("missing or invalid webpage URL")
	}
	result.CleanUrl = cleanUrl

	// Get thumbnail URL (check existence + type)
	thumbnails, ok := data["thumbnails"].([]interface{})
	if ok && len(thumbnails) > 0 {
		// Safely extract the last thumbnail
		lastThumb, ok := thumbnails[len(thumbnails)-1].(map[string]interface{})
		if ok {
			thumbnailURL, ok := lastThumb["url"].(string)
			if ok {
				result.ThumbnailURL = thumbnailURL
			}
		}
	}

	// Get playlist entries
	entries, ok := data["entries"].([]interface{})
	if !ok {
		// No entries found
		return result, nil
	}

	// Iterate over entries and add to result
	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		// Get title
		title, ok := entryMap["title"].(string)
		if !ok {
			continue
		}

		// Get URL
		url, ok := entryMap["url"].(string)
		if !ok {
			continue
		}

		// Add entry to result
		result.Entries = append(result.Entries, YtdlpEntry{
			Title: title,
			URL:   url,
		})
	}

	return result, nil
}

func DownloadFile(
	settingsService *settings.SettingsService,
	url,
	outputPath,
	format string,
) (string, error) {
	if format != "mp3" && format != "mp4" {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	ffmpegDir, err := getFfmpegDir()
	if err != nil {
		return "", fmt.Errorf("failed to get ffmpeg dir: %w", err)
	}

	baseArgs := []string{
		"--ffmpeg-location", ffmpegDir,
		"--prefer-ffmpeg",
		"--add-metadata",
		"--embed-thumbnail",
		"--embed-metadata",
		"--print-json",
		"--metadata-from-title", "%(artist)s - %(title)s",
		"--no-warnings",
	}

	var outputString string
	var outputError error

	if format == "mp3" {
		args := append([]string{"-x", "--audio-format", "mp3"}, baseArgs...)

		// Sponsorblock audio (stored as comma seperated string for multiselect settings)
		sponsorblockAudio, err := settingsService.GetSettingString("sponsorblock_audio")
		if err != nil {
			return "", fmt.Errorf("failed to get sponsorblock audio setting: %w", err)
		}
		if sponsorblockAudio != "" {
			args = append(args, "--sponsorblock-remove", sponsorblockAudio)
		}

		// Download
		outputString, outputError = runCommand(append(args, "-o", outputPath, url)...)
	} else { // mp4
		args := append([]string{
			"-f",
			"bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
			"--embed-chapters"},
			baseArgs...)

		// Sponsorblock video (stored as comma seperated string for multiselect settings)
		sponsorblockVideo, err := settingsService.GetSettingString("sponsorblock_video")
		if err != nil {
			return "", fmt.Errorf("failed to get sponsorblock video setting: %w", err)
		}
		if sponsorblockVideo != "" {
			args = append(args, "--sponsorblock-remove", sponsorblockVideo)
		}

		// Download
		outputString, outputError = runCommand(append(args, "-o", outputPath, url)...)
	}

	return outputString, outputError
}
