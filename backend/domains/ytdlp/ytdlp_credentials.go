package ytdlp

import (
	"fmt"
	"os"
	"videoarchiver/backend/domains/pathing"
)

// ExportBrowserCredentials exports browser credentials to a file using yt-dlp
// Returns the path to the credentials file and an error if any
// browserName should be one of: chrome, firefox, edge, opera, brave, safari, or "none"
func ExportBrowserCredentials(browserName string, logService LogServiceInterface) (string, error) {
	if browserName == "" || browserName == "none" {
		return "", nil // No credentials to export
	}

	// Get the credentials file path
	credPath, err := GetCredentialsFilePath()
	if err != nil {
		return "", fmt.Errorf("failed to get credentials file path: %w", err)
	}

	// Remove existing credentials file if it exists
	if fileExists(credPath) {
		os.Remove(credPath)
	}

	if logService != nil {
		logService.Debug(fmt.Sprintf("Exporting browser credentials from %s to %s", browserName, credPath))
	}

	// Export credentials using yt-dlp's --cookies-from-browser option
	// We need to provide a dummy URL to make yt-dlp actually export the cookies
	_, err = runCommand("--cookies-from-browser", browserName, "--cookies", credPath, "--print", "cookies", "https://www.youtube.com/")
	if err != nil {
		return "", fmt.Errorf("failed to export browser credentials from %s: %w", browserName, err)
	}

	if logService != nil {
		logService.Debug(fmt.Sprintf("Successfully exported browser credentials from %s", browserName))
	}

	return credPath, nil
}

// CleanupCredentialsFile removes the credentials file if it exists
func CleanupCredentialsFile(logService LogServiceInterface) error {
	credPath, err := GetCredentialsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get credentials file path: %w", err)
	}

	if fileExists(credPath) {
		if logService != nil {
			logService.Debug(fmt.Sprintf("Cleaning up credentials file: %s", credPath))
		}
		err := os.Remove(credPath)
		if err != nil {
			return fmt.Errorf("failed to remove credentials file: %w", err)
		}
	}

	return nil
}

// GetCredentialsFilePath returns the path where credentials file will be stored
func GetCredentialsFilePath() (string, error) {
	return pathing.GetWorkingFile("cookies.txt")
}

// GetCredentialsFilePathForDownload returns the path to use in download commands
// Returns empty string if no credentials file exists
func GetCredentialsFilePathForDownload() string {
	credPath, err := GetCredentialsFilePath()
	if err != nil {
		return ""
	}

	if fileExists(credPath) {
		return credPath
	}

	return ""
}
