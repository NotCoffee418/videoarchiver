package main

import (
	"context"
	"embed"
	"flag"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	mode := flag.String("mode", "", "Startup mode: ui, daemon, or empty (defaults to ui)")
	flag.Parse()

	switch *mode {
	case "daemon":
		app := NewApp(false)
		runDaemon(app)
	case "ui", "":
		app := NewApp(true)
		runUI(app)
	default:
		println("Invalid startup mode: " + *mode)
		os.Exit(1)
	}
}

func runDaemon(app *App) {
	app.startup(context.Background())
	daemonStartup(app)
}

func runUI(app *App) {
	// ✅ Create application with options
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
			app, // ✅ Bind only the App
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
