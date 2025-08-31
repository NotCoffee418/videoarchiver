package ytdlp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"videoarchiver/backend/domains/logging"
	"videoarchiver/backend/domains/pathing"
	"videoarchiver/backend/domains/runner"

	"github.com/mholt/archives"
)

const baseYtdlpDownloadUrl = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"

// memoized - call getExecutableFileName() once and store the result
var (
	ytdlpExecutableFileName   string = ""
	ffmpegExecutableFullPath  string = ""
	ffprobeExecutableFullPath string = ""
)

// SettingsChecker interface to check settings without importing the settings package
type SettingsChecker interface {
	GetSettingString(key string) (string, error)
}

func InstallOrUpdate(forceReinstall bool, settingsChecker SettingsChecker, logger ...*logging.LogService) error {
	var log *logging.LogService
	if len(logger) > 0 {
		log = logger[0]
	}
	
	// Install or update ytdlp
	err := installUpdateYtdlp(settingsChecker, log)
	if err != nil {
		return err
	}

	// Install or update ffmpeg
	err = installUpdateFfmpeg(false, log)
	if err != nil {
		return err
	}

	return nil
}

// Runs a ytdlp command and returns the stdout and stderr
func runCommand(args ...string) (string, error) {
	// Note: This function doesn't use logger to avoid changing all call sites
	// The command execution details are not critical for logging
	ytdlpPath, err := getYtdlpPath()
	if err != nil {
		return "", err
	}

	stdout, stderr, err := runner.RunWithOutput(ytdlpPath, args...)
	if err != nil {
		return stdout, fmt.Errorf("%s: %s", err, stderr)
	}

	stdoutStr := strings.TrimSpace(stdout)
	stderrStr := strings.TrimSpace(stderr)

	if stderrStr != "" {
		return stdoutStr, fmt.Errorf("ytdlp command failed: %s", stderrStr)
	}

	return stdoutStr, nil
}

// Install or update ytdlp
func installUpdateYtdlp(settingsChecker SettingsChecker, logger *logging.LogService) error {
	ytdlpPath, err := getYtdlpPath()
	if err != nil {
		return err
	}

	// Check for corruption
	if fileExists(ytdlpPath) {
		err = ytdlpCorruptionCheck(ytdlpPath)
		if err != nil {
			if logger != nil {
				logger.Info("ytdlp corruption check failed, reinstalling")
			}

			// Delete old version
			err = os.Remove(ytdlpPath)
			if err != nil {
				return fmt.Errorf("ytdlp instancer: failed to delete old ytdlp: %w", err)
			}
		}
	}

	// Download ytdlp if it doesn't exist
	if !fileExists(ytdlpPath) {
		if logger != nil {
			logger.Info("Downloading ytdlp...")
		}
		
		// Download
		downloadUrl := baseYtdlpDownloadUrl + getYtdlpExecutableFileName()
		err := downloadFileHttp(downloadUrl, ytdlpPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to download ytdlp: %w", err)
		}

		if logger != nil {
			logger.Info("ytdlp downloaded successfully")
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

	// Update ytdlp (check autoupdate setting first)
	shouldUpdate := true
	if settingsChecker != nil {
		autoupdate, err := settingsChecker.GetSettingString("autoupdate_ytdlp")
		if err == nil && autoupdate == "false" {
			shouldUpdate = false
			if logger != nil {
				logger.Debug("Skipping ytdlp update - autoupdate_ytdlp is disabled")
			}
		}
	}

	if shouldUpdate {
		if logger != nil {
			logger.Debug("Checking for ytdlp updates...")
		}
		_, err = runCommand("-U")
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to update ytdlp: %w", err)
		}

		if logger != nil {
			logger.Debug("ytdlp update check completed")
		}
	}

	return nil
}

func installUpdateFfmpeg(forceReinstall bool, logger *logging.LogService) error {
	if logger != nil {
		logger.Info("Installing or updating ffmpeg")
	}

	// Get ffmpeg path
	ffmpegPath, err := getFfmpegPath()
	if err != nil {
		return err
	}

	// Get ffprobe path
	ffprobePath, err := getFfprobePath()
	if err != nil {
		return err
	}

	// Delete existing ffmpeg if any, if forceUpdate is true
	if forceReinstall || ffmpegCorruptionCheck(ffmpegPath) != nil {
		if fileExists(ffmpegPath) {
			err := os.Remove(ffmpegPath)
			if err != nil {
				return fmt.Errorf("ytdlp instancer: failed to delete old ffmpeg: %w", err)
			}
		}
		forceReinstall = true
	}

	// Delete existing ffprobe if any, if forceUpdate is true
	if forceReinstall || ffprobeCorruptionCheck(ffprobePath) != nil {
		if fileExists(ffprobePath) {
			err := os.Remove(ffprobePath)
			if err != nil {
				return fmt.Errorf("ytdlp instancer: failed to delete old ffprobe: %w", err)
			}
		}
		forceReinstall = true
	}

	// Already exists, no update needed
	if !forceReinstall && fileExists(ffmpegPath) && fileExists(ffprobePath) {
		if logger != nil {
			logger.Debug("ffmpeg and ffprobe already exist, no update needed")
		}
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		if logger != nil {
			logger.Debug("Downloading ffmpeg for Windows...")
		}
		
		// Download ffmpeg to temp file
		downloadUrl := "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.7z"
		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("videoarchiver-ffmpeg-%d.7z", time.Now().UnixNano()))
		defer os.Remove(tmpFile)

		// Download archive to temp file
		err = downloadFileHttp(downloadUrl, tmpFile)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to download ffmpeg: %w", err)
		}

		if logger != nil {
			logger.Debug("Extracting ffmpeg and ffprobe...")
		}

		// Extract ffmpeg.exe
		err = extractFile(tmpFile, "bin/ffmpeg.exe", ffmpegPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffmpeg: %w", err)
		}

		// Extract ffprobe.exe
		err = extractFile(tmpFile, "bin/ffprobe.exe", ffprobePath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffprobe: %w", err)
		}
	case "linux":
		if logger != nil {
			logger.Debug("Downloading ffmpeg for Linux...")
		}
		
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

		if logger != nil {
			logger.Debug("Extracting ffmpeg and ffprobe...")
		}

		// Extract ffmpeg
		err = extractFile(tmpFile, "bin/ffmpeg", ffmpegPath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffmpeg: %w", err)
		}

		// Extract ffprobe
		err = extractFile(tmpFile, "bin/ffprobe", ffprobePath)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to extract ffprobe: %w", err)
		}

		// Make executable on unix
		err = os.Chmod(ffmpegPath, 0755)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to make ffmpeg executable: %w", err)
		}

		// Make executable on unix
		err = os.Chmod(ffprobePath, 0755)
		if err != nil {
			return fmt.Errorf("ytdlp instancer: failed to make ffprobe executable: %w", err)
		}
	default:
		return fmt.Errorf("ytdlp instancer: unsupported OS: %s", runtime.GOOS)
	}

	if logger != nil {
		logger.Info("ffmpeg and ffprobe installation completed successfully")
	}
	return nil
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

func getFfprobePath() (string, error) {
	if ffprobeExecutableFullPath == "" {
		p, err := pathing.GetWorkingFile(getFfprobeExecutableFileName())
		if err != nil {
			return "", err
		}
		ffprobeExecutableFullPath = p
	}

	return ffprobeExecutableFullPath, nil
}

func getFfmpegDir() (string, error) {
	ffmpegPath, err := getFfmpegPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ffmpegPath), nil
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

func getFfprobeExecutableFileName() string {
	switch runtime.GOOS {
	case "windows":
		return "ffprobe.exe"
	case "linux":
		return "ffprobe"
	}
	return "ffprobe"
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

func ffmpegCorruptionCheck(ffmpegPath string) error {
	// Use timeout-based check to detect crashes on Unix systems
	_, err := runner.RunCombinedOutputWithTimeout(5*time.Second, ffmpegPath, "-version")
	if err != nil {
		return fmt.Errorf("ffmpeg corruption check failed: %v", err)
	}
	return nil
}

func ffprobeCorruptionCheck(ffprobePath string) error {
	// Use timeout-based check to detect crashes on Unix systems
	_, err := runner.RunCombinedOutputWithTimeout(5*time.Second, ffprobePath, "-version")
	if err != nil {
		return fmt.Errorf("ffprobe corruption check failed: %v", err)
	}
	return nil
}

func ytdlpCorruptionCheck(ytdlpPath string) error {
	_, err := runCommand("--version")
	if err != nil {
		return fmt.Errorf("ytdlp corruption check failed, reinstalling: %v", err)
	}
	return nil
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
			if strings.HasSuffix(f.NameInArchive, fileToExtract) {
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
