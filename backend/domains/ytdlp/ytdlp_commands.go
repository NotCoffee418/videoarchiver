package ytdlp

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Get minimal playlist info
func GetPlaylistInfoFlat(url string) (*PlaylistInfo, error) {
	raw, err := runCommand("--no-warnings", "--flat-playlist", "--yes-playlist", "-J", url)
	if err != nil {
		return nil, err
	}

	// Prepare result object
	result := &PlaylistInfo{
		Title:        "",
		ThumbnailURL: "",
		Entries:      make([]Entry, 0),
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
		result.Entries = append(result.Entries, Entry{
			Title: title,
			URL:   url,
		})
	}

	return result, nil
}
