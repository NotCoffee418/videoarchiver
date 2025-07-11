package download

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"videoarchiver/backend/domains/ytdlp"
)

type DownloadService struct {
	ctx context.Context
}

func NewDownloadService(ctx context.Context) *DownloadService {
	return &DownloadService{ctx: ctx}
}

// Download a file via Ytdlp
func (d *DownloadService) DownloadFile(url, directory, format string) error {
	// Set temp path for the file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-download-%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tmpFile)

	// Download to temp path
	_, err := ytdlp.DownloadFile(url, tmpFile, format)
	if err != nil {
		return fmt.Errorf("download service: failed to download file: %w", err)
	}

	// Move file to directory
	// this is borked ------------------------------------- TODO!!!!!!
	os.Rename(tmpFile, filepath.Join(directory, filepath.Base(tmpFile)))

	// Add to database

	return errors.New("not implemented")
	return nil
}
