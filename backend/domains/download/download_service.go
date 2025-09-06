package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/fileregistry"
	"videoarchiver/backend/domains/fileutils"
	"videoarchiver/backend/domains/playlist"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/ytdlp"
)

// LogServiceInterface defines the logging interface to avoid circular imports
type LogServiceInterface interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
	Fatal(message string)
}

type DownloadService struct {
	ctx                 context.Context
	settingsService     *settings.SettingsService
	downloadDB          *DownloadDB
	fileRegistryService *fileregistry.FileRegistryService
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
	fileRegistryService *fileregistry.FileRegistryService,
	daemonSignalService *daemonsignal.DaemonSignalService,
	logService LogServiceInterface,
) *DownloadService {
	return &DownloadService{
		ctx:                 ctx,
		settingsService:     settingsService,
		downloadDB:          downloadDB,
		fileRegistryService: fileRegistryService,
		daemonSignalService: daemonSignalService,
		logService:          logService,
	}
}

// DownloadResult contains information about a download
type DownloadResult struct {
	TempFilePath   string
	FinalDirectory string
	FinalFileName  string
	FinalFullPath  string
	VideoTitle     string
	Format         string
	MD5            string
}

// ArchiveDownloadFile used by daemon and automated operations. Handles errors and logging.
// Handles duplicates, downloads table, error logging.
func (d *DownloadService) ArchiveDownloadFile(dl *Download, pl *playlist.Playlist) {
	// Download file
	d.logService.Info(fmt.Sprintf("Downloading new item: %s", dl.Url))
	dlR, err := d.DownloadFile(dl.Url, pl.SaveDirectory, pl.OutputFormat)
	if err != nil {
		d.logService.Error(fmt.Sprintf("Failed to download item %s: %v", dl.Url, err))
		dl.SetFail(d.downloadDB, err.Error())
		return
	}

	// Check settings if we need to filter duplicates
	allowDuplicates, err := d.settingsService.GetSettingBool("allow_duplicates")
	if err != nil {
		// should never happen
		d.logService.Fatal(fmt.Sprintf("Failed to get allow_duplicates setting for %s: %v", dl.Url, err))
		return
	}

	if !allowDuplicates {
		// Handle duplicate in downloads table
		isDup, existingId, err := d.HasDownloadsDuplicate(dlR.MD5, dl.ID)
		if err != nil {
			d.logService.Error(fmt.Sprintf("Failed to check for duplicate in downloads table for %s: %v", dl.Url, err))
			dl.SetFail(d.downloadDB, fmt.Sprintf("failed to check for duplicate in downloads table: %v", err))
			return
		}
		if isDup {
			d.logService.Info(fmt.Sprintf("Duplicate download detected in downloads table for %s (MD5: %s), skipping download. Existing ID: %d", dl.Url, dlR.MD5, existingId))
			if err := dl.SetSuccessDuplicate(d.downloadDB, dlR.FinalFileName, dlR.MD5); err != nil {
				d.logService.Error(fmt.Sprintf("Failed to mark download as duplicate for %s: %v", dl.Url, err))
			}
			return
		}

		// Handle duplicate in file registry
		isDup, err = d.HasFileRegistryDuplicate(dlR.MD5)
		if err != nil {
			d.logService.Error(fmt.Sprintf("Failed to check for duplicate in file registry for %s: %v", dl.Url, err))
			dl.SetFail(d.downloadDB, fmt.Sprintf("failed to check for duplicate in file registry: %v", err))
			return
		}
		if isDup {
			d.logService.Info(fmt.Sprintf("Duplicate download detected in file registry for %s (MD5: %s), skipping download.", dl.Url, dlR.MD5))
			if err := dl.SetSuccessDuplicate(d.downloadDB, dlR.FinalFileName, dlR.MD5); err != nil {
				d.logService.Error(fmt.Sprintf("Failed to mark download as duplicate for %s: %v", dl.Url, err))
			}
			return
		}
	}

	// Move to final location
	err = dlR.MoveToFinalLocation(pl.SaveDirectory)
	if err != nil {
		d.logService.Error(fmt.Sprintf("Failed to move downloaded file to final location for %s: %v", dl.Url, err))
		dl.SetFail(d.downloadDB, fmt.Sprintf("failed to move file to final location: %v", err))
		return
	}

	// Mark download as success
	if err := dl.SetSuccess(d.downloadDB, dlR.FinalFileName, dlR.MD5); err != nil {
		d.logService.Error(fmt.Sprintf("Failed to mark download as success for %s: %v", dl.Url, err))
		return
	}
}

// Download file to a temporary location. No duplicate handling here.
func (d *DownloadService) DownloadFile(url, directory, format string) (*DownloadResult, error) {
	d.logService.Info(fmt.Sprintf("Starting download: %s (format: %s, directory: %s)", url, format, directory))

	// Set temp path for the file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-download-%d.%s", time.Now().UnixNano(), format))

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

	// Calculate MD5 of the downloaded temp file
	fileMD5, err := CalculateMD5(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("download service: failed to calculate MD5: %w", err)
	}

	// Decide available filename, handling duplicate filenames.
	baseFilename := filepath.Base(videoTitle + "." + strings.ToLower(format))
	finalPath := filepath.Join(directory, baseFilename)
	fileNum := 0
	for fileExists(finalPath) {
		fileNum++
		baseName := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
		ext := filepath.Ext(baseFilename)
		finalPath = filepath.Join(directory, baseName+"-"+strconv.Itoa(fileNum)+ext)
	}

	return &DownloadResult{
		TempFilePath:   tmpFile,
		FinalDirectory: directory,
		FinalFileName:  filepath.Base(finalPath),
		FinalFullPath:  finalPath,
		VideoTitle:     videoTitle,
		Format:         format,
		MD5:            fileMD5,
	}, nil
}

// MoveToFinalLocation moves the downloaded file to its final location.
// Returns final path and error (if any)
func (dlR *DownloadResult) MoveToFinalLocation(finalDir string) error {
	// Move file to directory
	err := os.Rename(dlR.TempFilePath, dlR.FinalFullPath)
	if err != nil {
		return fmt.Errorf("download service: failed to move file: %w", err)
	}
	// Clean up temp file if it still exists
	os.Remove(dlR.TempFilePath)
	return nil
}

// CalculateMD5 calculates the MD5 hash of a file (now delegated to fileutils)
func CalculateMD5(path string) (string, error) {
	return fileutils.CalculateMD5(path)
}

func (d *DownloadService) HasFileRegistryDuplicate(fileMD5 string) (bool, error) {
	return d.fileRegistryService.CheckForDuplicateInFileRegistry(fileMD5)
}

func (d *DownloadService) HasDownloadsDuplicate(fileMD5 string, ignoredOwnId int) (bool, int, error) {
	return d.downloadDB.CheckForDuplicateInDownloads(fileMD5, ignoredOwnId)
}

func fileExists(path string) bool {
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
