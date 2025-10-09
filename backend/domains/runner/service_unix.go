//go:build !windows

package runner

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// configureProcessAttributes sets up process isolation on Unix systems.
// This helps prevent corrupted external binaries from crashing the main process.
func configureProcessAttributes(cmd *exec.Cmd) {
	// Set up process group isolation to prevent signal propagation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// configureProcessAttributesWithFlags sets up process isolation with flags on Unix systems.
func configureProcessAttributesWithFlags(cmd *exec.Cmd, creationFlags uint32) {
	// On Unix, we ignore creationFlags and use standard isolation
	configureProcessAttributes(cmd)
}

// runWithTimeoutCheck executes a command with a timeout to detect corrupted binaries.
// Returns an error if the binary crashes or times out.
func runWithTimeoutCheck(timeout time.Duration, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	configureProcessAttributes(cmd)

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("command timed out after %v (possible corrupted binary): %s", timeout, name)
	}
	if err != nil {
		// Convert exit errors to more descriptive messages for corruption detection
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ProcessState != nil && exitError.ProcessState.Sys() != nil {
				if ws, ok := exitError.ProcessState.Sys().(syscall.WaitStatus); ok {
					if ws.Signaled() {
						sig := ws.Signal()
						return fmt.Errorf("binary crashed with signal %v (likely corrupted): %s", sig, name)
					}
				}
			}
		}
		return fmt.Errorf("binary execution failed (possible corruption): %v", err)
	}
	return nil
}
