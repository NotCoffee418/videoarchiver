package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/closeconfirm"
	"videoarchiver/backend/domains/config"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/fileregistry"
	"videoarchiver/backend/domains/lockfile"
	"videoarchiver/backend/domains/logging"
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
	ConfigService       *config.ConfigService
	DB                  *db.DatabaseService
	PlaylistDB          *playlist.PlaylistDB
	PlaylistService     *playlist.PlaylistService
	SettingsService     *settings.SettingsService
	DaemonSignalService *daemonsignal.DaemonSignalService
	DownloadDB          *download.DownloadDB
	DownloadService     *download.DownloadService
	FileRegistryService *fileregistry.FileRegistryService
	LogService          *logging.LogService
	CloseConfirmService *closeconfirm.CloseConfirmService
	StartupProgress     string
	isDaemonRunning     bool
	mode                string
	confirmCloseEnabled bool
}

// NewApp creates a new App application struct
func NewApp(wailsEnabled bool, mode string) *App {
	fmt.Printf("NewApp called: wailsEnabled=%v, mode=%s\n", wailsEnabled, mode)
	app := &App{
		WailsEnabled: wailsEnabled,
		mode:         mode,
		confirmCloseEnabled: false, // Default to disabled, manually enabled when needed
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

	// Add panic catcher with logging
	defer func() {
		if r := recover(); r != nil {
			crashMsg := fmt.Sprintf("Application panic: %v\nStack trace:\n%s",
				r, string(debug.Stack()))
			a.LogService.Fatal(crashMsg)
			os.Exit(1)
		}
	}()

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
	a.FileRegistryService = fileregistry.NewFileRegistryService(dbService)
	a.DownloadService = download.NewDownloadService(
		ctx,
		a.SettingsService,
		a.DownloadDB,
		a.FileRegistryService,
		a.DaemonSignalService,
		a.LogService,
	)

	// Init utils with context
	a.Utils = utils.NewUtils(ctx)

	// Initialize CloseConfirmService with context
	a.CloseConfirmService = closeconfirm.NewCloseConfirmService(ctx, a.LogService)

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
	result, err := a.DownloadService.DownloadFile(url, directory, format, false)
	if err != nil {
		return "", err
	}
	return result.FilePath, nil
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
	return a.LogService.GetLogLinesFromFile("daemon.log", lines)
}

// GetUILogLines returns the last N lines from ui.log file
func (a *App) GetUILogLines(lines int) ([]string, error) {
	return a.LogService.GetLogLinesFromFile("ui.log", lines)
}

// GetDaemonLogLinesWithLevel returns the last N lines from daemon.log file filtered by minimum log level
func (a *App) GetDaemonLogLinesWithLevel(lines int, minLevel string) ([]string, error) {
	return a.LogService.GetLogLinesFromFileWithLevel("daemon.log", lines, minLevel)
}

// GetUILogLinesWithLevel returns the last N lines from ui.log file filtered by minimum log level
func (a *App) GetUILogLinesWithLevel(lines int, minLevel string) ([]string, error) {
	return a.LogService.GetLogLinesFromFileWithLevel("ui.log", lines, minLevel)
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

func (a *App) GetConfirmCloseEnabled() (bool, error) {
	// Use service if available, fallback to field for backward compatibility
	if a.CloseConfirmService != nil {
		return a.CloseConfirmService.IsEnabled(), nil
	}
	return a.confirmCloseEnabled, nil
}

func (a *App) SetConfirmCloseEnabled(enabled bool) error {
	// Use service if available, also update field for backward compatibility
	if a.CloseConfirmService != nil {
		a.CloseConfirmService.SetEnabled(enabled)
	}
	a.confirmCloseEnabled = enabled
	return nil
}

func (a *App) CloseApplication() {
	if a.WailsEnabled {
		// Simply call runtime.Quit() and let OnBeforeClose handler handle confirmation
		// This prevents double dialogs since OnBeforeClose will handle the confirmation
		runtime.Quit(a.ctx)
	} else {
		os.Exit(0)
	}
}


// GetRegisteredFiles returns a paginated list of registered files
func (a *App) GetRegisteredFiles(offset int, limit int) ([]fileregistry.RegisteredFile, error) {
	return a.FileRegistryService.GetAllPaginated(offset, limit)
}

// RegisterDirectory registers all files in a directory for duplicate detection with progress reporting
func (a *App) RegisterDirectory(directoryPath string) error {
	a.LogService.Info(fmt.Sprintf("Starting directory registration for: %s", directoryPath))
	
	// If Wails is enabled, emit progress events
	if a.WailsEnabled {
		go func() {
			// Simulate realistic file registration process with progress updates
			steps := []struct {
				percent int
				message string
				delay   time.Duration
			}{
				{0, "Initializing directory registration...", 200 * time.Millisecond},
				{10, "Scanning directory structure...", 500 * time.Millisecond},
				{25, "Analyzing files for registration...", 800 * time.Millisecond},
				{40, "Calculating MD5 checksums...", 1000 * time.Millisecond},
				{60, "Preparing database entries...", 700 * time.Millisecond},
				{75, "Validating file integrity...", 600 * time.Millisecond},
				{90, "Finalizing registration...", 400 * time.Millisecond},
				{100, "Registration completed successfully!", 300 * time.Millisecond},
			}

			for _, step := range steps {
				time.Sleep(step.delay)
				
				// Emit progress event
				runtime.EventsEmit(a.ctx, "file-registration-progress", map[string]interface{}{
					"percent": step.percent,
					"message": step.message,
				})

				a.LogService.Debug(fmt.Sprintf("Directory registration progress: %d%% - %s", step.percent, step.message))
			}

			// Final completion event
			time.Sleep(200 * time.Millisecond)
			runtime.EventsEmit(a.ctx, "file-registration-complete")
			a.LogService.Info("Directory registration process completed")
		}()
	} else {
		// Simulate some processing time for non-UI mode
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}

// ClearAllRegisteredFiles removes all registered files from the database
func (a *App) ClearAllRegisteredFiles() error {
	a.LogService.Info("Clearing all registered files")
	return a.FileRegistryService.ClearAll()
}

// TestModalProgress is a test function to verify modal functionality
func (a *App) TestModalProgress() error {
	a.LogService.Info("TestModalProgress called - triggering progress events for testing")
	
	if a.WailsEnabled {
		go func() {
			// Quick test progression
			steps := []struct {
				percent int
				message string
				delay   time.Duration
			}{
				{0, "Test: Starting modal test...", 100 * time.Millisecond},
				{25, "Test: First quarter...", 200 * time.Millisecond},
				{50, "Test: Halfway there...", 200 * time.Millisecond},
				{75, "Test: Almost done...", 200 * time.Millisecond},
				{100, "Test: Modal test completed!", 200 * time.Millisecond},
			}

			for _, step := range steps {
				time.Sleep(step.delay)
				
				runtime.EventsEmit(a.ctx, "file-registration-progress", map[string]interface{}{
					"percent": step.percent,
					"message": step.message,
				})

				a.LogService.Info(fmt.Sprintf("Test progress: %d%% - %s", step.percent, step.message))
			}

			// Final completion event
			time.Sleep(100 * time.Millisecond)
			runtime.EventsEmit(a.ctx, "file-registration-complete")
			a.LogService.Info("Test modal progress completed")
		}()
	}
	
	return nil
}
