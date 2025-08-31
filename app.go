package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"
	"videoarchiver/backend/daemonsignal"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/download"
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
	StartupProgress     string
	isDaemonRunning     bool
}

// NewApp creates a new App application struct
func NewApp(wailsEnabled bool) *App {
	app := &App{
		WailsEnabled: wailsEnabled,
	}
	// Check initial daemon state
	app.isDaemonRunning = app.IsDaemonRunning()
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

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

	// Start thread for spamming startup progress
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

// Centralized error handling
func (a *App) HandleFatalError(message string) {
	if a.WailsEnabled {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Application Error",
			Message: message,
		})
		os.Exit(1)
	} else {
		fmt.Println(message)
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
				fmt.Printf("Killed daemon process: PID=%d\n", p.Pid)
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

func (a *App) IsDaemonRunning() bool {
	switch goruntime.GOOS {
	case "windows":
		selfPid := os.Getpid()
		processes, err := process.Processes()
		if err != nil {
			fmt.Printf("Error getting process list: %v\n", err)
			return false
		}

		selfExe, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
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
				fmt.Printf("Found daemon process: PID=%d CMD=%s\n", p.Pid, cmdline)
				return true
			}
		}
		fmt.Println("No daemon process found")
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
