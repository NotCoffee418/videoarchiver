package runner

import (
	"bytes"
	"os/exec"
	"time"
)

// StartDetached starts a command and immediately returns without waiting.
// Used for operations like opening directories or starting daemon processes.
// OS-specific implementations handle console window hiding on Windows.
func StartDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)
	return cmd.Start()
}

// StartDetachedWithFlags starts a command with custom Windows creation flags.
// Used specifically for daemon processes that need special Windows process creation.
// On non-Windows systems, this behaves the same as StartDetached.
func StartDetachedWithFlags(name string, creationFlags uint32, args ...string) error {
	cmd := exec.Command(name, args...)
	configureProcessAttributesWithFlags(cmd, creationFlags)
	return cmd.Start()
}

// RunAndWait executes a command and waits for it to complete.
// Returns an error if the command fails.
// OS-specific implementations handle console window hiding on Windows.
func RunAndWait(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)
	return cmd.Run()
}

// RunWithOutput executes a command and returns stdout and stderr separately.
// Used for commands where we need to process the output, like ytdlp operations.
// OS-specific implementations handle console window hiding on Windows.
func RunWithOutput(name string, args ...string) (stdout string, stderr string, err error) {
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// RunCombinedOutput executes a command and returns the combined stdout and stderr.
// Used for simple version checks and similar operations.
// OS-specific implementations handle console window hiding on Windows.
func RunCombinedOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)
	return cmd.CombinedOutput()
}

// RunCombinedOutputWithTimeout executes a command with a timeout and returns the combined output.
// This is specifically used for corruption checks to prevent hanging on corrupted binaries.
func RunCombinedOutputWithTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	// Use the Unix-specific timeout function on Unix systems, regular execution on Windows
	if err := runWithTimeoutCheck(timeout, name, args...); err != nil {
		return nil, err
	}

	// If no crash/timeout, get the actual output
	cmd := exec.Command(name, args...)
	configureProcessAttributes(cmd)
	return cmd.CombinedOutput()
}
