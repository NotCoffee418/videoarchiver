package download

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	tmpFile, err := os.CreateTemp("", "videoarchiver-*")
	if err != nil {
		return fmt.Errorf("download service: failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Download to temp path
	_, err = ytdlp.DownloadFile(url, tmpFile.Name(), format)
	if err != nil {
		return fmt.Errorf("download service: failed to download file: %w", err)
	}

	// Move file to directory

	// Add to database

	return errors.New("not implemented")
	return nil
}
