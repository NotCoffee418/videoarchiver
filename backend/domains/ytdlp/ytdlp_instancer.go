package ytdlp

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
	"videoarchiver/backend/domains/pathing"
)

const baseYtdlpDownloadUrl = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"

// memoized - call getExecutableFileName() once and store the result
var executableFileName string = ""

func InstallOrUpdate() error {
	ytdlpPath, err := getLocalPath()
	if err != nil {
		return err
	}

	// Download ytdlp if it doesn't exist
	if _, err := os.Stat(ytdlpPath); os.IsNotExist(err) {
		// Download
		downloadUrl := baseYtdlpDownloadUrl + getExecutableFileName()
		err := downloadFile(downloadUrl, ytdlpPath)
		if err != nil {
			return err
		}

		// Make executable on unix
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			err := os.Chmod(ytdlpPath, 0755)
			if err != nil {
				return err
			}
		}

		// No need to check for update if we just installed it
		return nil
	}

	// Update
	_, err = runCommand("-U")
	if err != nil {
		return err
	}

	return nil
}

func runCommand(args ...string) (string, error) {
	ytdlpPath, err := getLocalPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ytdlpPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func getLocalPath() (string, error) {
	return pathing.GetWorkingFile(getExecutableFileName())
}

func getExecutableFileName() string {
	if executableFileName != "" {
		return executableFileName
	}

	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "386" {
			return "yt-dlp_x86.exe"
		}
		return "yt-dlp.exe"
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "yt-dlp_linux"
		case "arm64":
			return "yt-dlp_linux_aarch64"
		case "arm":
			return "yt-dlp_linux_armv7l"
		}
	case "darwin":
		return "yt-dlp_macos"
	}
	return "yt-dlp"
}

func downloadFile(url string, filePath string) error {
	// !!! TODO DELETE ME
	time.Sleep(2 * time.Second)

	// Open HTTP connection to the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Create the destination file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Stream response body directly to file (no memory buffering)
	_, err = out.ReadFrom(resp.Body)
	return err
}
