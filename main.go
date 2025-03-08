package main

import (
	"embed"
	"os"
	"videoarchiver/backend/domains/db"
	"videoarchiver/backend/domains/playlist"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create database service and playlist DB
	dbService, err := db.NewDatabaseService()
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
	playlistDB := playlist.NewPlaylistDB(dbService)

	// Create application with options
	err = wails.Run(&options.App{
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
			playlistDB,
		},
	})
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}

}
