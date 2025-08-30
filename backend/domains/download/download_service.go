package download

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/ytdlp"
)

type DownloadService struct {
	ctx                 context.Context
	settingsService     *settings.SettingsService
	downloadDB          *DownloadDB
	daemonSignalService *daemonsignal.DaemonSignalService
}

const (
	// Used to wrap errors from download service
	ErrDownloadErrorBase = "download service: failed to download file: "
)

func NewDownloadService(
	ctx context.Context,
	settingsService *settings.SettingsService,
	downloadDB *DownloadDB,
	daemonSignalService *daemonsignal.DaemonSignalService,
) *DownloadService {
	return &DownloadService{
		ctx:                 ctx,
		settingsService:     settingsService,
		downloadDB:          downloadDB,
		daemonSignalService: daemonSignalService,
	}
}

// Download a file via Ytdlp
func (d *DownloadService) DownloadFile(url, directory, format string) (string, error) {
	// Set temp path for the file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-download-%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tmpFile)

	// Download to temp path
	outputString, err := ytdlp.DownloadFile(d.settingsService, url, tmpFile, format)
	if err != nil {
		return "", fmt.Errorf("%s%w", ErrDownloadErrorBase, err)
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
	err = os.Rename(tmpFile, savePath)
	if err != nil {
		return "", fmt.Errorf("download service: failed to move file: %w", err)
	}

	return savePath, nil
}

func CalculateMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (d *DownloadService) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (d *DownloadService) SetManualRetry(downloadId int) error {
	if err := d.downloadDB.SetManualRetry(downloadId); err != nil {
		return fmt.Errorf("failed to set manual retry: %w", err)
	}
	return d.daemonSignalService.TriggerChange()
}
