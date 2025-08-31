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

// LogServiceInterface defines the logging interface to avoid circular imports
type LogServiceInterface interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
}

type DownloadService struct {
	ctx                 context.Context
	settingsService     *settings.SettingsService
	downloadDB          *DownloadDB
	daemonSignalService *daemonsignal.DaemonSignalService
	logService          LogServiceInterface
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
	logService LogServiceInterface,
) *DownloadService {
	return &DownloadService{
		ctx:                 ctx,
		settingsService:     settingsService,
		downloadDB:          downloadDB,
		daemonSignalService: daemonSignalService,
		logService:          logService,
	}
}

// DownloadResult contains information about a download
type DownloadResult struct {
	FilePath    string
	IsDuplicate bool
	DuplicateOf string // filename of the original file if it's a duplicate
}

// Download a file via Ytdlp
func (d *DownloadService) DownloadFile(url, directory, format string) (string, error) {
	result, err := d.DownloadFileWithDuplicateCheck(url, directory, format)
	if err != nil {
		return "", err
	}
	return result.FilePath, nil
}

// DownloadFileWithDuplicateCheck downloads a file and checks for duplicates
func (d *DownloadService) DownloadFileWithDuplicateCheck(url, directory, format string) (*DownloadResult, error) {
	d.logService.Debug(fmt.Sprintf("Starting download: %s (format: %s, directory: %s)", url, format, directory))
	
	// Set temp path for the file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-download-%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tmpFile)

	// Download to temp path
	outputString, err := ytdlp.DownloadFile(d.settingsService, url, tmpFile, format, d.logService)
	if err != nil {
		return nil, fmt.Errorf("%s%w", ErrDownloadErrorBase, err)
	}

	// Extract video title from ytdlp output
	videoTitle, err := ytdlp.GetString(outputString, "fulltitle")
	if err != nil {
		return nil, fmt.Errorf("download service: failed to get title: %w", err)
	}

	baseFilename := filepath.Base(videoTitle + "." + strings.ToLower(format))
	
	// Calculate MD5 of the downloaded file
	fileMD5, err := CalculateMD5(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("download service: failed to calculate MD5: %w", err)
	}

	// Check for duplicate in target directory
	duplicateFilename, err := d.CheckForDuplicateInDirectory(fileMD5, directory, baseFilename)
	if err != nil {
		return nil, fmt.Errorf("download service: failed to check for duplicates: %w", err)
	}

	if duplicateFilename != "" {
		// Duplicate found - don't move the file, just return the duplicate info
		duplicatePath := filepath.Join(directory, duplicateFilename)
		d.logService.Info(fmt.Sprintf("Download skipped: duplicate found - %s", duplicateFilename))
		return &DownloadResult{
			FilePath:    duplicatePath,
			IsDuplicate: true,
			DuplicateOf: duplicateFilename,
		}, nil
	}

	// No duplicate found - proceed with normal file placement and suffix logic
	savePath := filepath.Join(directory, baseFilename)
	fileNum := 0
	for d.fileExists(savePath) {
		fileNum++
		baseName := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
		ext := filepath.Ext(baseFilename)
		savePath = filepath.Join(directory, baseName+"-"+strconv.Itoa(fileNum)+ext)
	}

	// Move file to directory
	err = os.Rename(tmpFile, savePath)
	if err != nil {
		return nil, fmt.Errorf("download service: failed to move file: %w", err)
	}

	d.logService.Info(fmt.Sprintf("Download completed successfully: %s", filepath.Base(savePath)))
	return &DownloadResult{
		FilePath:    savePath,
		IsDuplicate: false,
		DuplicateOf: "",
	}, nil
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

// CheckForDuplicateInDirectory checks if any existing file in the directory has the same MD5
// It looks for files with the base name and numbered suffixes (filename.ext, filename-1.ext, etc.)
func (d *DownloadService) CheckForDuplicateInDirectory(fileMD5, targetDir, baseFilename string) (string, error) {
	// Extract base name and extension
	baseName := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	ext := filepath.Ext(baseFilename)
	
	// Check base filename and numbered variations
	for i := 0; i < 100; i++ { // reasonable limit to avoid infinite loops
		var checkFilename string
		if i == 0 {
			checkFilename = baseFilename
		} else {
			checkFilename = baseName + "-" + strconv.Itoa(i) + ext
		}
		
		checkPath := filepath.Join(targetDir, checkFilename)
		
		// Check if file exists
		if d.fileExists(checkPath) {
			// Calculate MD5 of existing file
			existingMD5, err := CalculateMD5(checkPath)
			if err != nil {
				// Skip files we can't read
				continue
			}
			
			// If MD5 matches, we found a duplicate
			if existingMD5 == fileMD5 {
				return checkFilename, nil
			}
		}
	}
	
	return "", nil // No duplicate found
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

func (d *DownloadService) RegisterAllFailedForRetryManual() error {
	if err := d.downloadDB.RegisterAllFailedForRetryManual(); err != nil {
		return fmt.Errorf("failed to register all failed downloads for retry: %w", err)
	}
	return d.daemonSignalService.TriggerChange()
}
