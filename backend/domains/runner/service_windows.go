//go:build windows

package runner

import (
	"os/exec"
	"syscall"
	"time"
)

// configureProcessAttributes sets Windows-specific process attributes to hide console windows.
func configureProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

// configureProcessAttributesWithFlags sets Windows-specific process attributes with custom creation flags.
func configureProcessAttributesWithFlags(cmd *exec.Cmd, creationFlags uint32) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: creationFlags,
	}
}

// runWithTimeoutCheck executes a command normally on Windows since Windows handles corruption properly.
func runWithTimeoutCheck(timeout time.Duration, name string, args ...string) error {
	// Windows already handles corrupted binaries gracefully, so just run normally
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)
	return cmd.Run()
}