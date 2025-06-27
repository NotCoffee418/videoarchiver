package download

import (
	"context"
	"errors"
)

type DownloadService struct {
	ctx context.Context
}

func NewDownloadService(ctx context.Context) *DownloadService {
	return &DownloadService{ctx: ctx}
}

func (d *DownloadService) DownloadFile(url, directory, format string) error {
	return errors.New("not implemented")
}
