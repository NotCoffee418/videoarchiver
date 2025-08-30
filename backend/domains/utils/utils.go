package utils

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"syscall"

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
		cmd := exec.Command("explorer", path)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		return cmd.Start()

	case "darwin":
		cmd := exec.Command("open", path) // macOS Finder
		return cmd.Start()
	case "linux":
		cmd := exec.Command("xdg-open", path) // Linux file manager
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func (u *Utils) GetClipboard() (string, error) {
	return wailsRuntime.ClipboardGetText(u.ctx)
}
