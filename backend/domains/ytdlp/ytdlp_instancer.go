package ytdlp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
	"videoarchiver/backend/domains/pathing"

	"github.com/mholt/archives"
)

const baseYtdlpDownloadUrl = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"

// memoized - call getExecutableFileName() once and store the result
var (
	ytdlpExecutableFileName  string = ""
	ffmpegExecutableFullPath string = ""
)

func InstallOrUpdate(forceReinstall bool) error {
	// Install or update ytdlp
	err := installUpdateYtdlp()
	if err != nil {
		return err
	}

	// Install or update ffmpeg
	err = installUpdateFfmpeg(false)
	if err != nil {
		return err
	}

	return nil
}

// Install or update ytdlp
func installUpdateYtdlp() error {
	ytdlpPath, err := getYtdlpPath()
	if err != nil {
		return err
	}

	// Download ytdlp if it doesn't exist
	if _, err := os.Stat(ytdlpPath); os.IsNotExist(err) {
		// Download
		downloadUrl := baseYtdlpDownloadUrl + getYtdlpExecutableFileName()
		err := downloadFileHttp(downloadUrl, ytdlpPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to download ytdlp: %w", err)
		}

		// Make executable on unix
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			err := os.Chmod(ytdlpPath, 0755)
			if err != nil {
				return fmt.Errorf("ytdlp instancer: failed to make executable: %w", err)
			}
		}

		// No need to check for update if we just installed it
		return nil
	}

	// Update ytdlp (even on fresh install)
	_, err = runCommand("-U")
	if err != nil {
		return fmt.Errorf("ytdlp instancer: failed to update ytdlp: %w", err)
	}

	return nil
}

func installUpdateFfmpeg(forceReinstall bool) error {
	// Get ffmpeg path
	ffmpegPath, err := getFfmpegPath()
	if err != nil {
		return err
	}

	// Delete existing ffmpeg if any, if forceUpdate is true
	if forceReinstall || fileExists(ffmpegPath) {
		err := os.Remove(ffmpegPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to delete old ffmpeg: %w", err)
		}
	}

	// Already exists, no update needed
	if fileExists(ffmpegPath) {
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		// Download ffmpeg to temp file
		downloadUrl := "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.7z"
		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-ffmpeg-%d.7z", time.Now().UnixNano()))
		defer os.Remove(tmpFile)

		// Download archive to temp file
		err = downloadFileHttp(downloadUrl, tmpFile)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to download ffmpeg: %w", err)
		}

		// Extract ffmpeg.exe
		err = extractFile(tmpFile, "bin/ffmpeg.exe", ffmpegPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffmpeg: %w", err)
		}
	case "linux":
		// Identify download for architecture
		var archStr string
		switch runtime.GOARCH {
		case "amd64":
			archStr = "linux64"
		case "arm64":
			archStr = "linuxarm64"
		default:
			return fmt.Errorf("ytdlp instancer: unsupported architecture: %s", runtime.GOARCH)
		}

		tarName := fmt.Sprintf("ffmpeg-master-latest-%s-lgpl", archStr)
		downloadUrl := fmt.Sprintf("https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/%s.tar.xz", tarName)

		// Download ffmpeg to temp file
		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-ffmpeg-%d.tar.xz", time.Now().UnixNano()))
		defer os.Remove(tmpFile)

		// Download ffmpeg to temp file
		err = downloadFileHttp(downloadUrl, tmpFile)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to download ffmpeg: %w", err)
		}

		// Extract ffmpeg
		err = extractFile(tmpFile, fmt.Sprintf("%s/bin/ffmpeg", tarName), ffmpegPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffmpeg: %w", err)
		}

		// Make executable on unix
		err = os.Chmod(ffmpegPath, 0755)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to make ffmpeg executable: %w", err)
		}
	default:
		return fmt.Errorf("ytdlp instancer: unsupported OS: %s", runtime.GOOS)
	}

	return nil
}

func runCommand(args ...string) (string, error) {
	ytdlpPath, err := getYtdlpPath()
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

// Get full path to the ytdlp executable
func getYtdlpPath() (string, error) {
	return pathing.GetWorkingFile(getYtdlpExecutableFileName())
}

// Get full path to the ffmpeg executable
func getFfmpegPath() (string, error) {
	if ffmpegExecutableFullPath == "" {
		p, err := pathing.GetWorkingFile(getFfmpegExecutableFileName())
		if err != nil {
			return "", err
		}
		ffmpegExecutableFullPath = p
	}

	return ffmpegExecutableFullPath, nil
}

// Get the name of the ytdlp executable on the current OS
func getYtdlpExecutableFileName() string {
	if ytdlpExecutableFileName != "" {
		return ytdlpExecutableFileName
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

// Get the name of the ffmpeg executable on the current OS
func getFfmpegExecutableFileName() string {
	switch runtime.GOOS {
	case "windows":
		return "ffmpeg.exe"
	case "linux":
		return "ffmpeg"
	}
	return "ffmpeg"
}

func downloadFileHttp(url string, filePath string) error {
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Extract a file from an archive
func extractFile(archivePath, fileToExtract, outputFile string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	format, _, err := archives.Identify(context.Background(), archivePath, file)
	if err != nil {
		return err
	}

	if ex, ok := format.(archives.Extractor); ok {
		found := false

		file.Seek(0, 0)
		err := ex.Extract(context.Background(), file, func(ctx context.Context, f archives.FileInfo) error {
			if f.NameInArchive == fileToExtract {
				found = true
				outFile, err := os.Create(outputFile)
				if err != nil {
					return err
				}
				defer outFile.Close()

				reader, err := f.Open()
				if err != nil {
					return err
				}
				defer reader.Close()

				_, err = io.Copy(outFile, reader)
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("file %s not found in archive", fileToExtract)
		}
		return nil
	}

	return fmt.Errorf("unsupported archive format")
}
