package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"videoarchiver/backend/domains/db"

	"github.com/NotCoffee418/dbmigrator"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:migrations
var migrationFS embed.FS

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is savedservices
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set up database service
	dbService, err := db.NewDatabaseService()
	if err != nil {
		a.HandleFatalError("Failed to create database service: " + err.Error())
		os.Exit(1)
	}
	db := dbService.GetDB()

	// Apply database migrations
	dbmigrator.SetDatabaseType(dbmigrator.SQLite)
	<-dbmigrator.MigrateUpCh(
		db,
		migrationFS,
		"migrations",
	)
}

func (a *App) HandleFatalError(message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   "Application Error",
		Message: message,
	})
	os.Exit(1)
}

// todo: DELETEME
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
