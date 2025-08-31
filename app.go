package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/config"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/lockfile"
	"videoarchiver/backend/domains/logging"
	"videoarchiver/backend/domains/pathing"
	"videoarchiver/backend/domains/playlist"
	"videoarchiver/backend/domains/runner"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/utils"
	"videoarchiver/backend/domains/ytdlp"

	goruntime "runtime" // renamed standard library runtime

	"github.com/NotCoffee418/dbmigrator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	ConfigService       *config.ConfigService
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
	fmt.Printf("NewApp called: wailsEnabled=%v, mode=%s\n", wailsEnabled, mode)
	app := &App{
		WailsEnabled: wailsEnabled,
		mode:         mode,
	}
	// Check initial daemon state
	fmt.Printf("Checking if daemon is running...\n")
	app.isDaemonRunning = app.IsDaemonRunning()
	fmt.Printf("Daemon running status: %v\n", app.isDaemonRunning)

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	fmt.Printf("startup() called for mode: %s, WailsEnabled: %v\n", a.mode, a.WailsEnabled)
	a.ctx = ctx

	// Initialize logging service early so it can be used throughout startup
	a.LogService = logging.NewLogService(a.mode)
	a.LogService.Info(fmt.Sprintf("Starting application startup for mode: %s", a.mode))

	// Start thread for spamming startup progress early
	// We need this because desync between js/backend
	if a.WailsEnabled {
		a.LogService.Info("Starting startup progress thread for UI mode")
		go func() {
			for !a.StartupComplete {
				if a.StartupProgress != "" {
					runtime.EventsEmit(a.ctx, "startup-progress", a.StartupProgress)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	// Create configuration service FIRST
	configService, err := config.NewConfigService()
	if err != nil {
		a.HandleFatalError("Failed to create configuration service: " + err.Error())
	}
	a.ConfigService = configService

	// LogService already initialized early in startup function

	// Create database service using configuration
	dbService, err := db.NewDatabaseService(configService, a.LogService)
	if err != nil {
		a.HandleFatalError("Failed to create database service: " + err.Error())
	}
	a.DB = dbService

	// Create SettingsService using dbService
	a.SettingsService = settings.NewSettingsService(dbService, a.LogService)

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
		a.LogService,
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

	// For UI mode, wait for legal disclaimer acceptance before installing dependencies
	if a.WailsEnabled {
		a.StartupProgress = "Waiting for legal disclaimer acceptance..."
		a.LogService.Info("UI mode: waiting for legal disclaimer acceptance before installing dependencies")
		for {
			accepted, err := a.GetLegalDisclaimerAccepted()
			if err != nil {
				a.LogService.Error(fmt.Sprintf("Failed to check legal disclaimer status: %v", err))
				time.Sleep(5 * time.Second)
				continue
			}
			if accepted {
				a.LogService.Info("Legal disclaimer accepted, proceeding with dependency installation")
				break
			}
			a.LogService.Debug("Waiting for legal disclaimer acceptance before proceeding...")
			time.Sleep(1 * time.Second)
		}

		// Then handle locking for UI mode
		if a.WailsEnabled {
			a.LogService.Info("UI mode detected, checking for daemon locking...")
			// UI mode (slave) - wait for daemon lock if present
			if err := a.handleUILocking(); err != nil {
				a.LogService.Error(fmt.Sprintf("LOG: UI about to exit due to locking error: %v", err))
				a.HandleFatalError("Failed to handle UI locking: " + err.Error())
			}
			a.LogService.Info("UI locking check completed successfully")
		}
	}

	// For daemon mode, wait for legal disclaimer acceptance before installing dependencies
	if !a.WailsEnabled {
		// Create lock before waiting for legal disclaimer acceptance
		a.LogService.Info("Creating lock file to coordinate with UI...")
		if err := lockfile.CreateLock(); err != nil {
			a.HandleFatalError("Failed to create lock file: " + err.Error())
		}

		// Ensure lock is removed on exit
		defer func() {
			if err := lockfile.RemoveLock(); err != nil {
				a.LogService.Warn(fmt.Sprintf("Failed to remove lock file: %v", err))
			}
		}()

		a.StartupProgress = "Waiting for legal disclaimer acceptance..."
		for {
			accepted, err := a.GetLegalDisclaimerAccepted()
			if err != nil {
				a.LogService.Error(fmt.Sprintf("Failed to check legal disclaimer status: %v", err))
				time.Sleep(5 * time.Second)
				continue
			}
			if accepted {
				a.LogService.Info("Legal disclaimer accepted, proceeding with dependency installation")
				break
			}
			a.LogService.Info("Waiting for legal disclaimer acceptance before proceeding...")
			time.Sleep(5 * time.Second)
		}
	}

	// ✅ Install ytdlp in background channel (after legal disclaimer is accepted)
	a.LogService.Info("Starting dependency installation after legal disclaimer acceptance")
	ytdlpUpdateChan := make(chan error)
	go func() {
		defer close(ytdlpUpdateChan)
		err := ytdlp.InstallOrUpdate(false, a.SettingsService, a.LogService)
		if err != nil {
			ytdlpUpdateChan <- err
		}
	}()

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

	a.LogService.Info("Application startup completed successfully")

	// For daemon mode, remove lock file after successful startup
	if !a.WailsEnabled {
		if err := lockfile.RemoveLock(); err != nil {
			a.LogService.Warn(fmt.Sprintf("Failed to remove lock file after startup: %v", err))
		} else {
			a.LogService.Info("Lock file removed, daemon ready for normal operation")
		}
	}

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
	a.LogService.Info("Starting handleUILocking check...")

	// Check if daemon lock exists
	locked, err := lockfile.IsLocked()
	if err != nil {
		a.LogService.Error(fmt.Sprintf("Failed to check lock status: %v", err))
		return fmt.Errorf("failed to check lock status: %w", err)
	}

	a.LogService.Info(fmt.Sprintf("Lock file check result: locked=%v", locked))

	if !locked {
		// No daemon lock, proceed normally
		a.LogService.Info("No daemon lock found, proceeding normally")
		return nil
	}

	// Lock exists - check if daemon is already running
	isDaemonAlreadyRunning := a.IsDaemonRunning()
	a.LogService.Info(fmt.Sprintf("Daemon running status while lock exists: %v", isDaemonAlreadyRunning))

	// Lock exists, wait for daemon to complete startup
	a.LogService.Info("Daemon lock found, waiting for daemon initialization to complete...")
	a.StartupProgress = "Waiting for daemon initialization... This may take a few minutes on first run."
	a.LogService.Info("Waiting for daemon initialization...")

	// Wait up to 10 minutes for the lock to be released
	a.LogService.Info("Starting 10-minute wait for lock release...")
	released, err := lockfile.WaitForLockRelease(10 * time.Minute)
	if err != nil {
		a.LogService.Error(fmt.Sprintf("Error while waiting for lock release: %v", err))
		return fmt.Errorf("error while waiting for lock release: %w", err)
	}

	a.LogService.Info(fmt.Sprintf("Lock release wait completed: released=%v", released))

	if !released {
		// Timeout reached, check if daemon is actually running
		a.LogService.Warn("Timeout waiting for daemon startup, checking if daemon is running...")
		isDaemonRunning := a.IsDaemonRunning()
		a.LogService.Info(fmt.Sprintf("Daemon running check result: %v", isDaemonRunning))

		if !isDaemonRunning {
			// Daemon not running but lock exists, try to start daemon
			a.LogService.Info("Daemon not running, attempting to start daemon...")
			a.StartupProgress = "Starting daemon..."
			if err := a.StartDaemon(); err != nil {
				a.LogService.Error(fmt.Sprintf("Failed to start daemon: %v", err))
				return fmt.Errorf("failed to start daemon: %w", err)
			}
			// Wait again for daemon to complete startup
			a.LogService.Info("Daemon started, waiting for startup completion...")
			a.StartupProgress = "Waiting for daemon initialization..."
			released, err := lockfile.WaitForLockRelease(5 * time.Minute)
			if err != nil {
				a.LogService.Error(fmt.Sprintf("Error while waiting for daemon startup: %v", err))
				return fmt.Errorf("error while waiting for daemon startup: %w", err)
			}
			if !released {
				a.LogService.Error("Timeout waiting for daemon startup after starting daemon")
				return fmt.Errorf("timeout waiting for daemon startup after starting daemon")
			}
		} else {
			a.LogService.Error("LOG: UI about to exit - daemon is running but lock file is not being released")
			a.LogService.Error(fmt.Sprintf("This suggests daemon (PID unknown) has a stale lock file"))
			a.LogService.Error("Possible causes: daemon didn't remove lock properly, race condition, or multiple daemon instances")
			a.LogService.Error("LOG: UI exiting due to: timeout waiting for daemon startup, but daemon is running")
			return fmt.Errorf("timeout waiting for daemon startup, but daemon is running")
		}
	}

	a.LogService.Info("Daemon startup complete, proceeding with UI startup...")
	return nil
}

// Centralized error handling
func (a *App) HandleFatalError(message string) {
	if a.WailsEnabled {
		if a.LogService != nil {
			a.LogService.Error("LOG: HandleFatalError called in UI mode: " + message)
		}
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Application Error",
			Message: message,
		})
		if a.LogService != nil {
			a.LogService.Error("Fatal error: " + message)
		}
		fmt.Printf("LOG: UI exiting due to fatal error: %s\n", message)
		os.Exit(1)
	} else {
		if a.LogService != nil {
			a.LogService.Error("LOG: HandleFatalError called in daemon mode: " + message)
			a.LogService.Error("Fatal error: " + message)
		} else {
			fmt.Println("LOG: Daemon exiting due to fatal error: " + message)
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

// GetDaemonLogLinesWithLevel returns the last N lines from daemon.log file filtered by minimum log level
func (a *App) GetDaemonLogLinesWithLevel(lines int, minLevel string) ([]string, error) {
	return a.getLogLinesFromFileWithLevel("daemon.log", lines, minLevel)
}

// GetUILogLinesWithLevel returns the last N lines from ui.log file filtered by minimum log level
func (a *App) GetUILogLinesWithLevel(lines int, minLevel string) ([]string, error) {
	return a.getLogLinesFromFileWithLevel("ui.log", lines, minLevel)
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

// getLogLinesFromFileWithLevel reads the last N lines from a log file and filters by minimum log level
func (a *App) getLogLinesFromFileWithLevel(filename string, lines int, minLevelStr string) ([]string, error) {
	// Parse minimum log level
	var minLevel logrus.Level
	switch strings.ToLower(minLevelStr) {
	case "debug":
		minLevel = logrus.DebugLevel
	case "info":
		minLevel = logrus.InfoLevel
	case "warn", "warning":
		minLevel = logrus.WarnLevel
	case "error":
		minLevel = logrus.ErrorLevel
	default:
		minLevel = logrus.InfoLevel // Default to info
	}

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

	// Split into lines
	allLines := strings.Split(string(content), "\n")

	// Remove empty last line if it exists
	if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
		allLines = allLines[:len(allLines)-1]
	}

	// Filter lines by log level
	var filteredLines []string
	for _, line := range allLines {
		if line == "" {
			continue
		}

		// Try to parse JSON log entry
		var logEntry struct {
			Level string `json:"level"`
		}
		
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			// If not JSON, include the line (could be plain text log)
			filteredLines = append(filteredLines, line)
			continue
		}

		// Parse log level from JSON
		var logLevel logrus.Level
		switch strings.ToLower(logEntry.Level) {
		case "debug":
			logLevel = logrus.DebugLevel
		case "info":
			logLevel = logrus.InfoLevel
		case "warn", "warning":
			logLevel = logrus.WarnLevel
		case "error":
			logLevel = logrus.ErrorLevel
		default:
			logLevel = logrus.InfoLevel
		}

		// Include line if it meets minimum level requirement
		// In logrus, lower numbers = higher severity, so check <=
		if logLevel <= minLevel {
			filteredLines = append(filteredLines, line)
		}
	}

	// Get last N filtered lines
	startIndex := 0
	if len(filteredLines) > lines {
		startIndex = len(filteredLines) - lines
	}

	result := filteredLines[startIndex:]
	if len(result) == 0 {
		return []string{"No log entries found matching the specified level"}, nil
	}

	return result, nil
}

func (a *App) IsDaemonRunning() bool {
	fmt.Printf("IsDaemonRunning called for mode: %s\n", a.mode)
	switch goruntime.GOOS {
	case "windows":
		selfPid := os.Getpid()
		fmt.Printf("Current process PID: %d\n", selfPid)
		processes, err := process.Processes()
		if err != nil {
			if a.LogService != nil {
				a.LogService.Error(fmt.Sprintf("Error getting process list: %v", err))
			}
			fmt.Printf("Error getting process list: %v\n", err)
			return false
		}

		selfExe, err := os.Executable()
		if err != nil {
			if a.LogService != nil {
				a.LogService.Error(fmt.Sprintf("Error getting executable path: %v", err))
			}
			fmt.Printf("Error getting executable path: %v\n", err)
			return false
		}
		selfExe = strings.ToLower(selfExe)
		fmt.Printf("Current executable path: %s\n", selfExe)

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
				fmt.Printf("Found daemon process: PID=%d CMD=%s\n", p.Pid, cmdline)
				if a.LogService != nil {
					a.LogService.Debug(fmt.Sprintf("Found daemon process: PID=%d CMD=%s", p.Pid, cmdline))
				}
				return true
			}
		}
		fmt.Printf("No daemon processes found\n")
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

func (a *App) GetLegalDisclaimerAccepted() (bool, error) {
	value, err := a.SettingsService.GetSettingString("legal_disclaimer_accepted")
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

func (a *App) SetLegalDisclaimerAccepted(accepted bool) error {
	value := "false"
	if accepted {
		value = "true"
	}
	return a.SettingsService.SetPreparsed("legal_disclaimer_accepted", value)
}

func (a *App) CloseApplication() {
	if a.WailsEnabled {
		runtime.Quit(a.ctx)
	} else {
		os.Exit(0)
	}
}
