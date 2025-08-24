package download

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/ytdlp"
)

type DownloadService struct {
	ctx             context.Context
	settingsService *settings.SettingsService
}

func NewDownloadService(ctx context.Context, settingsService *settings.SettingsService) *DownloadService {
	return &DownloadService{ctx: ctx, settingsService: settingsService}
}

// Download a file via Ytdlp
func (d *DownloadService) DownloadFile(url, directory, format string, saveToDB bool) (string, error) {
	// Set temp path for the file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-download-%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tmpFile)

	// Download to temp path
	outputString, err := ytdlp.DownloadFile(d.settingsService, url, tmpFile, format)
	if err != nil {
		return "", fmt.Errorf("download service: failed to download file: %w", err)
	}

	// Extract video title from ytdlp outpuit
	videoTitle, err := ytdlp.GetString(outputString, "fulltitle")
	if err != nil {
		return "", fmt.Errorf("download service: failed to get title: %w", err)
	}

	savePath := filepath.Join(directory, filepath.Base(videoTitle+"."+strings.ToLower(format)))
	fileNum := 0
	for d.fileExists(savePath) {
		fileNum++
		savePath = filepath.Join(directory, filepath.Base(videoTitle+"-"+strconv.Itoa(fileNum)+"."+strings.ToLower(format)))
	}

	// Move file to directory
	os.Rename(tmpFile, savePath)

	// Add to database

	return "", errors.New("not implemented")
}

func (d *DownloadService) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
