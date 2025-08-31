//go:build !windows

package runner

import "os/exec"

// configureProcessAttributes does nothing on non-Windows systems.
func configureProcessAttributes(cmd *exec.Cmd) {
	// No special configuration needed on non-Windows systems
}

// configureProcessAttributesWithFlags does nothing on non-Windows systems.
func configureProcessAttributesWithFlags(cmd *exec.Cmd, creationFlags uint32) {
	// No special configuration needed on non-Windows systems
}
