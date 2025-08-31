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

	// Early logging to track startup mode
	fmt.Printf("Starting application in mode: %s\n", *mode)

	switch *mode {
	case "daemon":
		fmt.Println("Initializing daemon mode...")
		app := NewApp(false, "daemon")
		runDaemon(app)

	case "ui", "":
		fmt.Println("Initializing UI mode...")
		app := NewApp(true, "ui")
		runUI(app)

	default:
		fmt.Println("LOG: Application exiting due to invalid startup mode")
		println("Invalid startup mode. Valid modes: ui, daemon")
		os.Exit(1)
	}
}

func runDaemon(app *App) {
	// Create early logger for daemon startup messages
	earlyLogger := logging.NewLogService("daemon")
	defer earlyLogger.Close()
	
	// Handle daemon locking before initialization - but only check, don't create yet
	earlyLogger.Info("Checking for existing daemon instances...")
	
	// Check if lock already exists
	locked, err := lockfile.IsLocked()
	if err != nil {
		earlyLogger.Error(fmt.Sprintf("Failed to check lock status: %v", err))
		earlyLogger.Error("LOG: Daemon exiting due to lock status check failure")
		os.Exit(1)
	}

	if locked {
		// Another daemon is starting up or running, exit this instance
		earlyLogger.Info("Another daemon instance is starting up. Exiting...")
		earlyLogger.Info("LOG: Daemon exiting due to existing daemon lock")
		os.Exit(0)
	}

	earlyLogger.Info("Initializing application")
	app.startup(context.Background())
	
	app.LogService.Info("Daemon starting")
	startDaemonLoop(app)
}

func runUI(app *App) {
	fmt.Println("Starting Wails UI...")
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
		fmt.Printf("LOG: UI exiting due to Wails error: %v\n", err)
		println("Error:", err.Error())
		os.Exit(1)
	}
}
