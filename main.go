package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"videoarchiver/backend/domains/lockfile"
	"videoarchiver/backend/domains/logging"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	mode := flag.String("mode", "", "Startup mode: ui, daemon (defaults to ui)")
	flag.Parse()

	switch *mode {
	case "daemon":
		app := NewApp(false, "daemon")
		runDaemon(app)

	case "ui", "":
		app := NewApp(true, "ui")
		runUI(app)

	default:
		println("Invalid startup mode. Valid modes: ui, daemon")
		os.Exit(1)
	}
}

func runDaemon(app *App) {
	// Create early logger for daemon startup messages
	earlyLogger := logging.NewLogService("daemon")
	defer earlyLogger.Close()

	// Handle daemon locking before initialization
	earlyLogger.Info("Checking for existing daemon instances...")

	// Check if lock already exists
	locked, err := lockfile.IsLocked()
	if err != nil {
		earlyLogger.Error(fmt.Sprintf("Failed to check lock status: %v", err))
		os.Exit(1)
	}

	if locked {
		// Another daemon is starting up or running, exit this instance
		earlyLogger.Info("Another daemon instance is starting up. Exiting...")
		os.Exit(0)
	}

	// Create lock file
	if err := lockfile.CreateLock(); err != nil {
		earlyLogger.Error(fmt.Sprintf("Failed to create lock file: %v", err))
		os.Exit(1)
	}

	// Ensure lock is removed on exit
	defer func() {
		if err := lockfile.RemoveLock(); err != nil {
			earlyLogger.Warn(fmt.Sprintf("Failed to remove lock file: %v", err))
		}
	}()

	earlyLogger.Info("Initializing application")
	app.startup(context.Background())

	// Remove lock after successful startup
	if err := lockfile.RemoveLock(); err != nil {
		app.LogService.Warn(fmt.Sprintf("Failed to remove lock file after startup: %v", err))
	}

	app.LogService.Info("Daemon starting")
	startDaemonLoop(app)
}

func runUI(app *App) {
	err := wails.Run(&options.App{
		Title:  "videoarchiver",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
