package lockfile

import (
	"fmt"
	"os"
	"time"
	"videoarchiver/backend/domains/pathing"
)

const lockFileName = ".lock"

// CreateLock creates a lock file in the working directory
func CreateLock() error {
	lockPath, err := pathing.GetWorkingFile(lockFileName)
	if err != nil {
		return fmt.Errorf("failed to get lock file path: %w", err)
	}

	// Create the lock file with current timestamp
	file, err := os.Create(lockPath)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	defer file.Close()

	// Write timestamp to the file
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	if _, err := file.WriteString(timestamp); err != nil {
		return fmt.Errorf("failed to write timestamp to lock file: %w", err)
	}

	return nil
}

// RemoveLock removes the lock file
func RemoveLock() error {
	lockPath, err := pathing.GetWorkingFile(lockFileName)
	if err != nil {
		return fmt.Errorf("failed to get lock file path: %w", err)
	}

	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}

// IsLocked checks if a lock file exists and returns true if it exists and is recent
func IsLocked() (bool, error) {
	lockPath, err := pathing.GetWorkingFile(lockFileName)
	if err != nil {
		return false, fmt.Errorf("failed to get lock file path: %w", err)
	}

	stat, err := os.Stat(lockPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat lock file: %w", err)
	}

	// Check if the lock file is older than 5 minutes
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	if stat.ModTime().Before(fiveMinutesAgo) {
		// Lock is stale, remove it
		if err := RemoveLock(); err != nil {
			return false, fmt.Errorf("failed to remove stale lock file: %w", err)
		}
		return false, nil
	}

	return true, nil
}

// WaitForLockRelease waits for the lock file to be removed
// Returns true if lock was released, false if timeout reached
func WaitForLockRelease(timeout time.Duration) (bool, error) {
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			locked, err := IsLocked()
			if err != nil {
				return false, err
			}
			if !locked {
				return true, nil
			}
			if time.Since(start) >= timeout {
				return false, nil
			}
		}
	}
}
