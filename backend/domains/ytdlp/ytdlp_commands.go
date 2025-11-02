package ytdlp

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/imaging"
)

// LogServiceInterface defines the logging interface to avoid circular imports
type LogServiceInterface interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
	Fatal(message string)
}

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
	logService LogServiceInterface,
	withCredentials bool,
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
		"--add-metadata",
		"--embed-thumbnail",
		"--embed-metadata",
		"--print-json",
		"--metadata-from-title", "%(artist)s - %(title)s",
		"--no-warnings",
		"--no-playlist",
	}

	// Add credentials if requested
	if withCredentials {
		credPath := GetCredentialsFilePathForDownload()
		if credPath != "" {
			baseArgs = append(baseArgs, "--cookies", credPath)
			if logService != nil {
				logService.Debug("Using credentials file for download")
			}
		}
	}

	var outputString string
	var outputError error

	if format == "mp3" {
		args := append([]string{"-x", "--audio-format", "mp3", "--audio-quality", "0"}, baseArgs...)

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
			"bestvideo+bestaudio/best",
			"--merge-output-format", "mp4",
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

	// Check if download failed due to private/age-restricted content and retry with credentials if not already used
	if outputError != nil && !withCredentials {
		errorMsg := outputError.Error()
		needsAuth := strings.Contains(errorMsg, "Private video") ||
			strings.Contains(errorMsg, "members-only") ||
			strings.Contains(errorMsg, "This video is private") ||
			strings.Contains(errorMsg, "Sign in to confirm your age") ||
			strings.Contains(errorMsg, "age-restricted") ||
			strings.Contains(errorMsg, "This video requires payment") ||
			strings.Contains(errorMsg, "Join this channel")

		if needsAuth {
			// Try again with browser credentials if configured
			browserSource, err := settingsService.GetSettingString("browser_credentials_source")
			if err == nil && browserSource != "" && browserSource != "none" {
				if logService != nil {
					logService.Info(fmt.Sprintf("Video requires authentication, retrying with browser credentials from %s", browserSource))
				}

				// Export credentials
				credPath, err := ExportBrowserCredentials(browserSource, logService)
				if err != nil {
					if logService != nil {
						logService.Warn(fmt.Sprintf("Failed to export credentials for retry: %v", err))
					}
				} else if credPath != "" {
					// Retry with credentials
					outputString, outputError = DownloadFile(settingsService, url, outputPath, format, logService, true)

					if logService != nil {
						if outputError == nil {
							logService.Info("Successfully downloaded video with browser credentials")
						} else {
							logService.Warn(fmt.Sprintf("Retry with credentials also failed: %v", outputError))
						}
					}
				}
			}
		}
	}

	// Log verbose output for debugging
	if logService != nil {
		if outputString != "" {
			logService.Debug("yt-dlp output: " + outputString)
		}
		if outputError != nil {
			logService.Debug("yt-dlp error: " + outputError.Error())
		}
	}

	return outputString, outputError
}

// GetThumbnailBase64 fetches a thumbnail from a URL and returns it as base64
func GetThumbnailBase64(url string) (string, error) {
	return imaging.GetBase64Thumb(url)
}
