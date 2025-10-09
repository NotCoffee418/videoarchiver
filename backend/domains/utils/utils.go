package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"videoarchiver/backend/domains/runner"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Utils struct {
	ctx context.Context
}

// ✅ Fix: Assign context correctly in constructor
func NewUtils(ctx context.Context) *Utils {
	return &Utils{ctx: ctx}
}

// ✅ Fix: Add startup method if binding directly
func (u *Utils) startup(ctx context.Context) {
	u.ctx = ctx
}

func (u *Utils) SelectDirectory() (string, error) {
	if u.ctx == nil {
		return "", fmt.Errorf("context not initialized")
	}
	return wailsRuntime.OpenDirectoryDialog(u.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Folder",
	})
}

func (u *Utils) OpenDirectory(path string) error {
	switch runtime.GOOS {
	case "windows":
		return runner.StartDetached("explorer", path)
	case "darwin":
		return runner.StartDetached("open", path) // macOS Finder
	case "linux":
		return runner.StartDetached("xdg-open", path) // Linux file manager
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func (u *Utils) GetClipboard() (string, error) {
	return wailsRuntime.ClipboardGetText(u.ctx)
}

func (u *Utils) GetDownloadsDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	var downloadDir string
	switch runtime.GOOS {
	case "windows":
		downloadDir = filepath.Join(home, "Downloads")
	case "darwin":
		downloadDir = filepath.Join(home, "Downloads")
	case "linux":
		downloadDir = filepath.Join(home, "Downloads")
	default:
		downloadDir = home
	}

	return downloadDir, nil
}
