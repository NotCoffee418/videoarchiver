package fileregistry

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/fileutils"
	"videoarchiver/backend/domains/logging"
)

type FileRegistryService struct {
	db *sql.DB
}

func NewFileRegistryService(dbService *db.DatabaseService) *FileRegistryService {
	return &FileRegistryService{db: dbService.GetDB()}
}

// GetByMD5 returns the first registered file with the given MD5 hash
func (f *FileRegistryService) GetByMD5(md5Hash string) (*RegisteredFile, error) {
	row := f.db.QueryRow(
		"SELECT id, filename, file_path, md5, registered_at FROM file_registry WHERE md5 = ? LIMIT 1",
		md5Hash,
	)
	
	var file RegisteredFile
	err := row.Scan(&file.ID, &file.Filename, &file.FilePath, &file.MD5, &file.RegisteredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No duplicate found
		}
		return nil, err
	}
	
	return &file, nil
}

// RegisterFile adds a new file to the registry
func (f *FileRegistryService) RegisterFile(filename, filePath, md5Hash string) error {
	_, err := f.db.Exec(
		"INSERT INTO file_registry (filename, file_path, md5, registered_at) VALUES (?, ?, ?, ?)",
		filename, filePath, md5Hash, time.Now().Unix(),
	)
	return err
}

// GetAllPaginated returns a paginated list of registered files
func (f *FileRegistryService) GetAllPaginated(offset, limit int) ([]RegisteredFile, error) {
	rows, err := f.db.Query(
		"SELECT id, filename, file_path, md5, registered_at FROM file_registry ORDER BY registered_at DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []RegisteredFile
	for rows.Next() {
		var file RegisteredFile
		err := rows.Scan(&file.ID, &file.Filename, &file.FilePath, &file.MD5, &file.RegisteredAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	
	return files, nil
}

// ClearAll removes all registered files from the database
func (f *FileRegistryService) ClearAll() error {
	_, err := f.db.Exec("DELETE FROM file_registry")
	return err
}

// ProgressCallback defines the signature for progress reporting callbacks
type ProgressCallback func(percent int, message string)

// RegisterDirectoryWithProgress registers all files in a directory with progress reporting
func (f *FileRegistryService) RegisterDirectoryWithProgress(directoryPath string, logService *logging.LogService, progressCallback ProgressCallback) error {
	// Step 1: Initialize and validate input
	if progressCallback != nil {
		progressCallback(0, "Initializing directory registration...")
	}
	
	// Trim and validate directory path
	directoryPath = strings.TrimSpace(directoryPath)
	logService.Info(fmt.Sprintf("Starting directory registration for: '%s' (length: %d)", directoryPath, len(directoryPath)))
	
	if directoryPath == "" {
		return fmt.Errorf("directory path is empty")
	}
	
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directoryPath)
	}
	
	// Step 2: Scan directory to count files
	if progressCallback != nil {
		progressCallback(10, "Scanning directory structure...")
	}
	
	var allFiles []string
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logService.Warn(fmt.Sprintf("Error accessing path %s: %v", path, err))
			return nil // Continue walking, don't fail on individual file errors
		}
		
		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}
	
	totalFiles := len(allFiles)
	logService.Info(fmt.Sprintf("Found %d files to register", totalFiles))
	
	if totalFiles == 0 {
		if progressCallback != nil {
			progressCallback(100, "No files found in directory")
		}
		return nil
	}
	
	// Step 3: Process files with progress updates
	registeredCount := 0
	errorCount := 0
	
	for i, filePath := range allFiles {
		// Calculate progress (20% for setup, 80% for file processing)
		progressPercent := 20 + int(float64(i)/float64(totalFiles)*80)
		
		if progressCallback != nil {
			progressCallback(progressPercent, fmt.Sprintf("Processing file %d of %d: %s", i+1, totalFiles, filepath.Base(filePath)))
		}
		
		// Calculate MD5 hash
		md5Hash, err := fileutils.CalculateMD5(filePath)
		if err != nil {
			logService.Warn(fmt.Sprintf("Failed to calculate MD5 for %s: %v", filePath, err))
			errorCount++
			continue
		}
		
		// Register the file
		filename := filepath.Base(filePath)
		err = f.RegisterFile(filename, filePath, md5Hash)
		if err != nil {
			logService.Warn(fmt.Sprintf("Failed to register file %s: %v", filePath, err))
			errorCount++
			continue
		}
		
		registeredCount++
		logService.Debug(fmt.Sprintf("Registered file: %s (MD5: %s)", filePath, md5Hash))
	}
	
	// Final progress update
	if progressCallback != nil {
		var message string
		if errorCount > 0 {
			message = fmt.Sprintf("Registration completed: %d files registered, %d errors", registeredCount, errorCount)
		} else {
			message = fmt.Sprintf("Registration completed successfully: %d files registered", registeredCount)
		}
		progressCallback(100, message)
	}
	
	logService.Info(fmt.Sprintf("Directory registration completed: %d files registered, %d errors", registeredCount, errorCount))
	return nil
}