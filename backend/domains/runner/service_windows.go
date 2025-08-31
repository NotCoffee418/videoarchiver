//go:build windows

package runner

import (
	"os/exec"
	"syscall"
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
