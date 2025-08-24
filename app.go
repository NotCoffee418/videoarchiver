package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/playlist"
	"videoarchiver/backend/domains/settings"
	"videoarchiver/backend/domains/utils"
	"videoarchiver/backend/domains/ytdlp"

	"github.com/NotCoffee418/dbmigrator"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:migrations
var migrationFS embed.FS

// App struct
type App struct {
	ctx             context.Context
	StartupComplete bool
	Utils           *utils.Utils
	DB              *db.DatabaseService
	PlaylistDB      *playlist.PlaylistDB
	PlaylistService *playlist.PlaylistService
	SettingsService *settings.SettingsService
	DownloadService *download.DownloadService
	StartupProgress string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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

	// Create PlaylistDB using dbService
	a.PlaylistDB = playlist.NewPlaylistDB(dbService)
	a.PlaylistService = playlist.NewPlaylistService(a.PlaylistDB)

	// Create DownloadService using dbService
	a.DownloadService = download.NewDownloadService(ctx, a.SettingsService)

	// Init utils with context
	a.Utils = utils.NewUtils(ctx)

	// Start thread for spamming startup progress
	// We need this because desync between js/backend
	go func() {
		for !a.StartupComplete {
			if a.StartupProgress != "" {
				runtime.EventsEmit(a.ctx, "startup-progress", a.StartupProgress)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

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

// Centralized error handling
func (a *App) HandleFatalError(message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   "Application Error",
		Message: message,
	})
	os.Exit(1)
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
	if _, err := os.Stat(newDirectory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", newDirectory)
	}
	return a.PlaylistDB.UpdatePlaylistDirectory(id, newDirectory)
}

func (a *App) ValidateAndAddPlaylist(url, directory, format string) error {
	return a.PlaylistService.TryAddNewPlaylist(url, directory, format)
}

func (a *App) DeletePlaylist(id int) error {
	return a.PlaylistDB.DeletePlaylist(id)
}

func (a *App) IsStartupComplete() bool {
	return a.StartupComplete
}

func (a *App) GetSettingString(key string) (string, error) {
	return a.SettingsService.GetSettingString(key)
}

func (a *App) SetSettingPreparsed(key string, value string) error {
	return a.SettingsService.SetPreparsed(key, value)
}

func (a *App) DirectDownload(url, directory, format string) (string, error) {
	return a.DownloadService.DownloadFile(url, directory, format)
}
