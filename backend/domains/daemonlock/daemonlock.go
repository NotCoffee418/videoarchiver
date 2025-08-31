package daemonlock

import (
	"fmt"
	"os"
	"time"
	"videoarchiver/backend/domains/pathing"
)

const (
	lockFileName    = "daemon.lock"
	lockTimeout     = 10 * time.Minute // Timeout for lock files
)

// DaemonLockService handles daemon startup locking
type DaemonLockService struct{}

// NewDaemonLockService creates a new daemon lock service
func NewDaemonLockService() *DaemonLockService {
	return &DaemonLockService{}
}

// getLockFilePath returns the full path to the daemon lock file
func (d *DaemonLockService) getLockFilePath() (string, error) {
	return pathing.GetWorkingFile(lockFileName)
}

// CreateLock creates a daemon startup lock file
func (d *DaemonLockService) CreateLock() error {
	lockPath, err := d.getLockFilePath()
	if err != nil {
		return fmt.Errorf("failed to get lock file path: %v", err)
	}

	// Create lock file with current timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	err = os.WriteFile(lockPath, []byte(timestamp), 0644)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %v", err)
	}

	return nil
}

// RemoveLock removes the daemon startup lock file
func (d *DaemonLockService) RemoveLock() error {
	lockPath, err := d.getLockFilePath()
	if err != nil {
		return fmt.Errorf("failed to get lock file path: %v", err)
	}

	// Remove lock file if it exists
	if _, err := os.Stat(lockPath); err == nil {
		err = os.Remove(lockPath)
		if err != nil {
			return fmt.Errorf("failed to remove lock file: %v", err)
		}
	}

	return nil
}

// IsLocked checks if daemon is currently locked (booting)
func (d *DaemonLockService) IsLocked() bool {
	lockPath, err := d.getLockFilePath()
	if err != nil {
		return false
	}

	// Check if lock file exists
	info, err := os.Stat(lockPath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}

	// Check if lock file is not too old (timeout handling)
	if time.Since(info.ModTime()) > lockTimeout {
		// Lock has timed out, remove it
		d.RemoveLock()
		return false
	}

	// Additional check: read timestamp from file and validate
	data, err := os.ReadFile(lockPath)
	if err != nil {
		// If we can't read the file, assume it's corrupted and remove it
		d.RemoveLock()
		return false
	}

	// Validate timestamp format and timeout
	var timestamp int64
	if _, err := fmt.Sscanf(string(data), "%d", &timestamp); err != nil {
		// Invalid timestamp format, remove lock
		d.RemoveLock()
		return false
	}

	lockTime := time.Unix(timestamp, 0)
	if time.Since(lockTime) > lockTimeout {
		// Lock has timed out, remove it
		d.RemoveLock()
		return false
	}

	return true
}

// GetLockAge returns how long the lock has been active
func (d *DaemonLockService) GetLockAge() (time.Duration, error) {
	lockPath, err := d.getLockFilePath()
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return 0, fmt.Errorf("lock file not found or unreadable")
	}

	var timestamp int64
	if _, err := fmt.Sscanf(string(data), "%d", &timestamp); err != nil {
		return 0, fmt.Errorf("invalid timestamp in lock file")
	}

	lockTime := time.Unix(timestamp, 0)
	return time.Since(lockTime), nil
}