package main

import (
	"context"
	"embed"
	"os"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/playlist"
	"videoarchiver/backend/domains/utils"

	"github.com/NotCoffee418/dbmigrator"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:migrations
var migrationFS embed.FS

// App struct
type App struct {
	ctx        context.Context
	Utils      *utils.Utils
	DB         *db.DatabaseService
	PlaylistDB *playlist.PlaylistDB
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// ✅ Create database service ONCE
	dbService, err := db.NewDatabaseService()
	if err != nil {
		a.HandleFatalError("Failed to create database service: " + err.Error())
	}
	a.DB = dbService

	// ✅ Create PlaylistDB using dbService
	a.PlaylistDB = playlist.NewPlaylistDB(dbService)

	// ✅ Init utils with context
	a.Utils = utils.NewUtils(ctx)

	// ✅ Apply database migrations (AFTER setting up DB)
	db := dbService.GetDB()
	dbmigrator.SetDatabaseType(dbmigrator.SQLite)
	<-dbmigrator.MigrateUpCh(
		db,
		migrationFS,
		"migrations",
	)
}

// ✅ Centralized error handling
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
func (a *App) GetPlaylists() ([]playlist.Playlist, error) {
	return a.PlaylistDB.GetPlaylists()
}

func (a *App) OpenDirectory(path string) error {
	return a.Utils.OpenDirectory(path)
}

func (a *App) SelectDirectory() (string, error) {
	return a.Utils.SelectDirectory()
}
