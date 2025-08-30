// Daemon loop that runs in the background
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/ytdlp"
)

var (
	app        *App
	cancelFunc context.CancelFunc
	lastRun    time.Time = time.Time{}
)

const (
	daemonWorkCheckInterval     = 5 * time.Second
	daemonPlaylistCheckInterval = 30 * time.Minute
)

func startDaemonLoop(_app *App) {
	app = _app
	fmt.Println("Starting daemon loop...")

	// Create context and shutdown handling here
	ctx, _cancelFunc := context.WithCancel(context.Background())
	cancelFunc = _cancelFunc

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutdown signal received")
		cancelFunc()
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Daemon loop shutting down")
			return
		default:
			// Check if we need to do work due to time elapsed
			doWork := lastRun.Add(daemonPlaylistCheckInterval).Before(time.Now())
			if doWork {
				fmt.Println("Running iteration: time elapsed since last run")
			}

			// Check if we need to do work due to change signal
			if !doWork {
				isChangeTriggered, err := app.DaemonSignalService.IsChangeTriggered()
				if err != nil {
					fmt.Printf("Error: Failed to check if change is triggered: %v\n", err)
					cancelFunc()
					return
				}
				if isChangeTriggered {
					fmt.Println("Running iteration: change triggered by UI")
					err := app.DaemonSignalService.ClearChangeTrigger()
					if err != nil {
						fmt.Printf("Error: Failed to clear change trigger: %v\n", err)
						cancelFunc()
						return
					}
					doWork = true
				}
			}

			// Do work if needed
			if doWork {
				lastRun = time.Now()
				processActivePlaylists()
			}

			// Then wait 5s (or until cancelled)
			select {
			case <-ctx.Done():
				// Break out of the inner select triggering the outer one
				return
			case <-time.After(daemonWorkCheckInterval):
				// Continue to next iteration
				continue
			}
		}
	}
}

// Process playlists
func processActivePlaylists() {
	fmt.Println("Processing playlists...")

	// Get acive playlists
	activePlaylists, err := app.PlaylistDB.GetActivePlaylists()
	if err != nil {
		fmt.Printf("Error: Failed to get active playlists: %v\n", err)
		return
	}

	// Loop over active playlists
	for _, pl := range activePlaylists {
		fmt.Printf("Processing playlist: %s\n", pl.Name)

		// Get playlist items online
		plInfo, err := ytdlp.GetPlaylistInfoFlat(pl.URL)
		if err != nil {
			fmt.Printf("Error: Failed to get playlist info for %s: %v\n", pl.Name, err)
			continue
		}

		// Check which playlist items are already processed
		existingDls, err := app.DownloadDB.GetDownloadsForPlaylist(pl.ID)
		if err != nil {
			fmt.Printf("Error: Failed to get existing downloads for playlist %s: %v\n", pl.Name, err)
			continue
		}

		// Filter out already downloaded urls
		unprocessedUrls := filterUndownloadedUrls(plInfo, existingDls)
		if len(unprocessedUrls) == 0 {
			fmt.Printf("No new items to download for playlist: %s\n", pl.Name)
			continue
		}
		fmt.Printf("Found %d new items to download for playlist: %s\n", len(unprocessedUrls), pl.Name)

		// todo
		fmt.Println("niy")
		os.Exit(0)

	}

	fmt.Println("Playlist processing complete.")
}

func filterUndownloadedUrls(plInfo *ytdlp.YtdlpPlaylistInfo, existingDls []download.Download) []string {
	// Create map for actual playlist items
	plItemsMap := make(map[string]bool)
	for _, item := range plInfo.Entries {
		plItemsMap[item.URL] = false
	}

	// Indicate which items are already downloaded
	for _, dl := range existingDls {
		if _, exists := plItemsMap[dl.VideoID]; exists {
			plItemsMap[dl.VideoID] = true
		}
	}

	// Return urls that are not yet downloaded
	unprocessedUrls := make([]string, 0)
	for url, isDownloaded := range plItemsMap {
		if !isDownloaded {
			unprocessedUrls = append(unprocessedUrls, url)
		}
	}
	return unprocessedUrls
}
