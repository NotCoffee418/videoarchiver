package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/logging"
	"videoarchiver/backend/domains/lockfile"
	"videoarchiver/backend/domains/pathing"
	"videoarchiver/backend/domains/playlist"
	"videoarchiver/backend/domains/runner"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/utils"
	"videoarchiver/backend/domains/ytdlp"

	goruntime "runtime" // renamed standard library runtime

	"github.com/NotCoffee418/dbmigrator"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/wailsapp/wails/v2/pkg/runtime" // keep wails runtime as is
)

//go:embed all:migrations
var migrationFS embed.FS

const (
	WindowsServiceName = "VideoArchiverDaemon" // Must match definition in project.nsi
	LinuxServiceName   = "video-archiver.service"
)

// App struct
type App struct {
	ctx                 context.Context
	WailsEnabled        bool
	StartupComplete     bool
	Utils               *utils.Utils
	DB                  *db.DatabaseService
	PlaylistDB          *playlist.PlaylistDB
	PlaylistService     *playlist.PlaylistService
	SettingsService     *settings.SettingsService
	DaemonSignalService *daemonsignal.DaemonSignalService
	DownloadDB          *download.DownloadDB
	DownloadService     *download.DownloadService
	LogService          *logging.LogService
	StartupProgress     string
	isDaemonRunning     bool
	mode                string
}

// NewApp creates a new App application struct
func NewApp(wailsEnabled bool, mode string) *App {
	app := &App{
		WailsEnabled: wailsEnabled,
		mode:         mode,
	}
	// Check initial daemon state
	app.isDaemonRunning = app.IsDaemonRunning()
	
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start thread for spamming startup progress early
	// We need this because desync between js/backend
	if a.WailsEnabled {
		go func() {
			for !a.StartupComplete {
				if a.StartupProgress != "" {
					runtime.EventsEmit(a.ctx, "startup-progress", a.StartupProgress)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	// Handle locking mechanism for UI mode only
	if a.WailsEnabled {
		// UI mode (slave) - wait for daemon lock if present
		if err := a.handleUILocking(); err != nil {
			a.HandleFatalError("Failed to handle UI locking: " + err.Error())
		}
	}

	// ✅ Install ytdlp in background channel
	ytdlpUpdateChan := make(chan error)
	go func() {
		defer close(ytdlpUpdateChan)
		err := ytdlp.InstallOrUpdate(false)
		if err != nil {
			ytdlpUpdateChan <- err
		}
	}()

	// Create database service ONCE
	dbService, err := db.NewDatabaseService()
	if err != nil {
		a.HandleFatalError("Failed to create database service: " + err.Error())
	}
	a.DB = dbService

	// Create SettingsService using dbService
	a.SettingsService = settings.NewSettingsService(dbService)

	// Create logging services
	a.LogService = logging.NewLogService(a.mode)

	// Create DaemonTrigger service
	a.DaemonSignalService = daemonsignal.NewDaemonSignalService(a.SettingsService)

	// Create PlaylistDB using dbService
	a.PlaylistDB = playlist.NewPlaylistDB(dbService)
	a.PlaylistService = playlist.NewPlaylistService(a.PlaylistDB, a.DaemonSignalService)

	// Create DownloadService using dbService
	a.DownloadDB = download.NewDownloadDB(dbService)
	a.DownloadService = download.NewDownloadService(
		ctx,
		a.SettingsService,
		a.DownloadDB,
		a.DaemonSignalService,
	)

	// Init utils with context
	a.Utils = utils.NewUtils(ctx)

	// Apply database migrations (AFTER setting up DB)
	a.StartupProgress = "Applying database updates..."
	db := dbService.GetDB()
	dbmigrator.SetDatabaseType(dbmigrator.SQLite)
	<-dbmigrator.MigrateUpCh(
		db,
		migrationFS,
		"migrations",
	)

	// Prepare message for ytdlp update if it needs to do a full install/update
	a.StartupProgress = "Checking dependencies..."
	ytdlpUpdateDone := false
	ytdlpUpdateStartTime := time.Now()
	go func() {
		for !ytdlpUpdateDone {
			if time.Since(ytdlpUpdateStartTime) > 3*time.Second {
				a.StartupProgress = "Updating dependencies. This may take a minute or two. Please wait..."
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Install/update ytdlp/ffmpeg
	err = <-ytdlpUpdateChan
	if err != nil {
		a.HandleFatalError("Failed to install ytdlp: " + err.Error())
	}
	ytdlpUpdateDone = true

	// Test the logging system with sample entries
	a.LogService.Info("Application startup completed successfully")
	a.LogService.Debug("Debug logging system test")
	a.LogService.Warn("Warning logging system test")

	// Emit startup complete event in background
	a.StartupProgress = "Startup complete"
	if a.WailsEnabled {
		go func() {
			// Listen for confirmed event
			awaitingConfirmation := true
			runtime.EventsOn(a.ctx, "startup-complete-confirmed", func(data ...interface{}) {
				awaitingConfirmation = false
			})

			// emit complete event
			a.StartupComplete = true
			for i := 0; i < 300; i++ {
				if !awaitingConfirmation {
					break
				}
				runtime.EventsEmit(a.ctx, "startup-complete")
				time.Sleep(250 * time.Millisecond)
			}
		}()
	}
}

// handleUILocking handles locking for UI mode (slave)
func (a *App) handleUILocking() error {
	// Check if daemon lock exists
	locked, err := lockfile.IsLocked()
	if err != nil {
		return fmt.Errorf("failed to check lock status: %w", err)
	}

	if !locked {
		// No daemon lock, proceed normally
		return nil
	}

	// Lock exists, wait for daemon to complete startup
	a.StartupProgress = "Waiting for daemon initialization..."
	a.LogService.Info("Waiting for daemon initialization...")

	// Wait up to 10 minutes for the lock to be released
	released, err := lockfile.WaitForLockRelease(10 * time.Minute)
	if err != nil {
		return fmt.Errorf("error while waiting for lock release: %w", err)
	}

	if !released {
		// Timeout reached, check if daemon is actually running
		a.LogService.Warn("Timeout waiting for daemon startup, checking if daemon is running...")
		if !a.IsDaemonRunning() {
			// Daemon not running but lock exists, try to start daemon
			a.LogService.Info("Daemon not running, attempting to start daemon...")
			a.StartupProgress = "Starting daemon..."
			if err := a.StartDaemon(); err != nil {
				return fmt.Errorf("failed to start daemon: %w", err)
			}
			// Wait again for daemon to complete startup
			a.StartupProgress = "Waiting for daemon initialization..."
			released, err := lockfile.WaitForLockRelease(5 * time.Minute)
			if err != nil {
				return fmt.Errorf("error while waiting for daemon startup: %w", err)
			}
			if !released {
				return fmt.Errorf("timeout waiting for daemon startup after starting daemon")
			}
		} else {
			return fmt.Errorf("timeout waiting for daemon startup, but daemon is running")
		}
	}

	a.LogService.Info("Daemon startup complete, proceeding with UI startup...")
	return nil
}

// Centralized error handling
func (a *App) HandleFatalError(message string) {
	if a.WailsEnabled {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Application Error",
			Message: message,
		})
		if a.LogService != nil {
			a.LogService.Error("Fatal error: " + message)
		}
		os.Exit(1)
	} else {
		if a.LogService != nil {
			a.LogService.Error("Fatal error: " + message)
		} else {
			fmt.Println("Fatal error: " + message)
		}
		os.Exit(1)
	}
}

// -- Bind functions - Dont try to fix, just add them here
// -- Hours wasted: 2
func (a *App) GetActivePlaylists() ([]playlist.Playlist, error) {
	return a.PlaylistDB.GetActivePlaylists()
}

func (a *App) OpenDirectory(path string) error {
	return a.Utils.OpenDirectory(path)
}

func (a *App) SelectDirectory() (string, error) {
	return a.Utils.SelectDirectory()
}

func (a *App) GetClipboard() (string, error) {
	return a.Utils.GetClipboard()
}

func (a *App) UpdatePlaylistDirectory(id int, newDirectory string) error {
	return a.PlaylistService.TryUpdatePlaylistDirectory(id, newDirectory)
}

func (a *App) ValidateAndAddPlaylist(url, directory, format string) error {
	return a.PlaylistService.TryAddNewPlaylist(url, directory, format)
}

func (a *App) DeletePlaylist(id int) error {
	return a.PlaylistService.TryDeletePlaylist(id)
}

func (a *App) IsStartupComplete() bool {
	return a.StartupComplete
}

func (a *App) GetSettingString(key string) (string, error) {
	return a.SettingsService.GetSettingString(key)
}

func (a *App) SetSettingPreparsed(key string, value string) error {
	err := a.SettingsService.SetPreparsed(key, value)
	if err != nil {
		return errors.Wrap(err, "failed to set setting")
	}

	// UI Settings changes trigger daemon change signal
	return a.DaemonSignalService.TriggerChange()

}

func (a *App) DirectDownload(url, directory, format string) (string, error) {
	return a.DownloadService.DownloadFile(url, directory, format)
}

func (a *App) GetDownloadHistoryPage(offset int, limit int, showSuccess, showFailed bool) ([]download.Download, error) {
	return a.DownloadDB.GetDownloadHistoryPage(offset, limit, showSuccess, showFailed)
}

func (a *App) SetManualRetry(downloadId int) error {
	return a.DownloadService.SetManualRetry(downloadId)
}

func (a *App) RegisterAllFailedForRetryManual() error {
	return a.DownloadService.RegisterAllFailedForRetryManual()
}

func (a *App) StartDaemon() error {
	if a.isDaemonRunning {
		return nil
	}

	switch goruntime.GOOS {
	case "windows":
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %v", err)
		}

		err = runner.StartDetachedWithFlags(exePath, 0x00000200|0x00000008, "--mode", "daemon")
		if err != nil {
			return fmt.Errorf("failed to start daemon: %v", err)
		}

		// Wait a moment to check if process started successfully
		time.Sleep(500 * time.Millisecond)
		if !a.IsDaemonRunning() {
			return fmt.Errorf("daemon process failed to start")
		}

	case "linux":
		err := runner.RunAndWait("systemctl", "start", LinuxServiceName)
		if err != nil {
			// Fallback to direct execution if service not installed
			err = runner.StartDetached(os.Args[0], "--daemon")
			if err != nil {
				return fmt.Errorf("failed to start daemon: %v", err)
			}
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", goruntime.GOOS)
	}

	a.isDaemonRunning = true
	return nil
}

func (a *App) StopDaemon() error {
	if !a.isDaemonRunning {
		return nil
	}

	switch goruntime.GOOS {
	case "windows":
		selfPid := os.Getpid()
		processes, err := process.Processes()
		if err != nil {
			return fmt.Errorf("failed to list processes: %v", err)
		}

		selfExe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %v", err)
		}
		selfExe = strings.ToLower(selfExe)

		for _, p := range processes {
			if int32(selfPid) == p.Pid {
				continue
			}

			exe, err := p.Exe()
			if err != nil {
				continue
			}

			cmdline, _ := p.Cmdline()
			exe = strings.ToLower(exe)

			// Only kill processes running in daemon mode
			if exe == selfExe && strings.Contains(cmdline, "--mode daemon") {
				if err := p.Kill(); err != nil {
					return fmt.Errorf("failed to kill process %d: %v", p.Pid, err)
				}
				a.LogService.Info(fmt.Sprintf("Killed daemon process: PID=%d", p.Pid))
			}
		}

	case "linux":
		err := runner.RunAndWait("systemctl", "stop", LinuxServiceName)
		if err != nil {
			// Fallback to finding and killing the process
			runner.RunAndWait("pkill", "-f", os.Args[0])
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", goruntime.GOOS)
	}

	a.isDaemonRunning = false
	return nil
}

// GetRecentLogs returns empty array as database logging is no longer supported
// Use GetDaemonLogLines() or GetUILogLines() for file-based logs instead
func (a *App) GetRecentLogs() ([]interface{}, error) {
	return []interface{}{}, nil
}

// GetDaemonLogLines returns the last N lines from daemon.log file
func (a *App) GetDaemonLogLines(lines int) ([]string, error) {
	return a.getLogLinesFromFile("daemon.log", lines)
}

// GetUILogLines returns the last N lines from ui.log file
func (a *App) GetUILogLines(lines int) ([]string, error) {
	return a.getLogLinesFromFile("ui.log", lines)
}

// getLogLinesFromFile reads the last N lines from a log file using proper pathing
func (a *App) getLogLinesFromFile(filename string, lines int) ([]string, error) {
	// Get proper log file path using pathing system
	logFilePath, err := pathing.GetWorkingFile(filename)
	if err != nil {
		return []string{fmt.Sprintf("Error getting log file path: %v", err)}, nil
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"Log file does not exist yet"}, nil
		}
		return nil, fmt.Errorf("failed to open log file %s: %v", logFilePath, err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file %s: %v", logFilePath, err)
	}

	if len(content) == 0 {
		return []string{"Log file is empty"}, nil
	}

	// Split into lines and get the last N lines
	allLines := strings.Split(string(content), "\n")
	
	// Remove empty last line if it exists
	if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
		allLines = allLines[:len(allLines)-1]
	}

	// Get last N lines
	startIndex := 0
	if len(allLines) > lines {
		startIndex = len(allLines) - lines
	}

	result := allLines[startIndex:]
	if len(result) == 0 {
		return []string{"No log entries found"}, nil
	}

	return result, nil
}

func (a *App) IsDaemonRunning() bool {
	switch goruntime.GOOS {
	case "windows":
		selfPid := os.Getpid()
		processes, err := process.Processes()
		if err != nil {
			a.LogService.Error(fmt.Sprintf("Error getting process list: %v", err))
			return false
		}

		selfExe, err := os.Executable()
		if err != nil {
			a.LogService.Error(fmt.Sprintf("Error getting executable path: %v", err))
			return false
		}
		selfExe = strings.ToLower(selfExe)

		for _, p := range processes {
			if int32(selfPid) == p.Pid {
				continue
			}

			exe, err := p.Exe()
			if err != nil {
				continue
			}

			cmdline, _ := p.Cmdline()
			exe = strings.ToLower(exe)

			// Check if it's our executable AND it's running in daemon mode
			if exe == selfExe && strings.Contains(cmdline, "--mode daemon") {
				a.LogService.Debug(fmt.Sprintf("Found daemon process: PID=%d CMD=%s", p.Pid, cmdline))
				return true
			}
		}
		return false

	case "linux":
		err := runner.RunAndWait("systemctl", "is-active", LinuxServiceName)
		if err == nil {
			a.isDaemonRunning = true
			return true
		}
	}

	// Check for running process in case of direct execution
	// This is a simplified check and might need improvement
	a.isDaemonRunning = false
	return false
}
