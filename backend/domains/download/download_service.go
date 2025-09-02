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
	"videoarchiver/backend/domains/fileregistry"
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
	ctx                    context.Context
	settingsService        *settings.SettingsService
	downloadDB             *DownloadDB
	fileRegistryService    *fileregistry.FileRegistryService
	daemonSignalService    *daemonsignal.DaemonSignalService
	logService             LogServiceInterface
}

const (
	// Used to wrap errors from download service
	ErrDownloadErrorBase = "download service: failed to download file: "
)

func NewDownloadService(
	ctx context.Context,
	settingsService *settings.SettingsService,
	downloadDB *DownloadDB,
	fileRegistryService *fileregistry.FileRegistryService,
	daemonSignalService *daemonsignal.DaemonSignalService,
	logService LogServiceInterface,
) *DownloadService {
	return &DownloadService{
		ctx:                    ctx,
		settingsService:        settingsService,
		downloadDB:             downloadDB,
		fileRegistryService:    fileRegistryService,
		daemonSignalService:    daemonSignalService,
		logService:             logService,
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
	d.logService.Info(fmt.Sprintf("Starting download: %s (format: %s, directory: %s)", url, format, directory))

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

	// Check for duplicate across all sources (file_registry, downloads table, and directory)
	duplicateInfo, err := d.CheckForDuplicateAcrossAllSources(fileMD5, directory, baseFilename)
	if err != nil {
		return nil, fmt.Errorf("download service: failed to check for duplicates: %w", err)
	}

	if duplicateInfo != nil {
		// Duplicate found - don't move the file, just return the duplicate info
		d.logService.Info(fmt.Sprintf("Download skipped: duplicate found - %s (source: %s)", 
			duplicateInfo.DuplicateOf, duplicateInfo.Source))
		return &DownloadResult{
			FilePath:    duplicateInfo.FilePath,
			IsDuplicate: true,
			DuplicateOf: duplicateInfo.DuplicateOf,
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

	// Register file in file registry for future duplicate detection
	fileName := filepath.Base(savePath)
	err = d.fileRegistryService.RegisterFile(fileName, savePath, fileMD5)
	if err != nil {
		d.logService.Warn(fmt.Sprintf("Failed to register file in registry: %v", err))
		// Don't fail the download if registry registration fails
	}

	d.logService.Info(fmt.Sprintf("Download completed successfully: %s", fileName))
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

// DuplicateInfo contains information about a detected duplicate
type DuplicateInfo struct {
	FilePath    string
	DuplicateOf string // filename of the original file
	Source      string // which source detected the duplicate: "file_registry", "downloads", or "directory"
}

// CheckForDuplicateAcrossAllSources checks for duplicates across file_registry table, downloads table, and directory files
func (d *DownloadService) CheckForDuplicateAcrossAllSources(fileMD5, targetDir, baseFilename string) (*DuplicateInfo, error) {
	// First check if duplicate checking is disabled
	allowDuplicates, err := d.settingsService.GetSettingBool("allow_duplicates")
	if err != nil {
		d.logService.Warn(fmt.Sprintf("Failed to get allow_duplicates setting: %v", err))
		// Continue with duplicate checking if setting can't be read
	} else if allowDuplicates {
		d.logService.Debug("Duplicate checking disabled by allow_duplicates setting")
		return nil, nil
	}

	// 1. Check file_registry table for duplicate by MD5
	registeredFile, err := d.fileRegistryService.GetByMD5(fileMD5)
	if err != nil {
		return nil, fmt.Errorf("failed to check file_registry for duplicates: %w", err)
	}
	if registeredFile != nil {
		return &DuplicateInfo{
			FilePath:    registeredFile.FilePath,
			DuplicateOf: registeredFile.Filename,
			Source:      "file_registry",
		}, nil
	}

	// 2. Check downloads table for duplicate by MD5
	downloadDuplicate, err := d.CheckForDuplicateInDownloads(fileMD5)
	if err != nil {
		return nil, fmt.Errorf("failed to check downloads table for duplicates: %w", err)
	}
	if downloadDuplicate != nil {
		return downloadDuplicate, nil
	}

	// 3. Check target directory for duplicate by MD5 (existing logic)
	duplicateFilename, err := d.CheckForDuplicateInDirectory(fileMD5, targetDir, baseFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to check directory for duplicates: %w", err)
	}
	if duplicateFilename != "" {
		duplicatePath := filepath.Join(targetDir, duplicateFilename)
		return &DuplicateInfo{
			FilePath:    duplicatePath,
			DuplicateOf: duplicateFilename,
			Source:      "directory",
		}, nil
	}

	return nil, nil // No duplicate found
}

// CheckForDuplicateInDownloads checks if any existing download has the same MD5
func (d *DownloadService) CheckForDuplicateInDownloads(fileMD5 string) (*DuplicateInfo, error) {
	// Query downloads table for matching MD5 with valid output filename
	rows, err := d.downloadDB.db.Query(
		"SELECT output_filename FROM downloads WHERE md5 = ? AND output_filename IS NOT NULL AND output_filename != '' LIMIT 1",
		fileMD5,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var outputFilename string
		err := rows.Scan(&outputFilename)
		if err != nil {
			return nil, err
		}
		
		return &DuplicateInfo{
			FilePath:    outputFilename, // This might be just filename, not full path
			DuplicateOf: filepath.Base(outputFilename),
			Source:      "downloads",
		}, nil
	}
	
	return nil, nil // No duplicate found
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
