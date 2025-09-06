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
	app.LogService.Info("Starting daemon loop...")

	// Create context and shutdown handling here
	ctx, _cancelFunc := context.WithCancel(context.Background())
	cancelFunc = _cancelFunc

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		app.LogService.Info("Shutdown signal received")
		cancelFunc()
	}()

	for {
		select {
		case <-ctx.Done():
			app.LogService.Info("Daemon loop shutting down")
			return
		default:
			// Check if we need to do work due to time elapsed
			doWork := lastRun.Add(daemonPlaylistCheckInterval).Before(time.Now())
			if doWork {
				app.LogService.Info("Running iteration: time elapsed since last run")
			}

			// Check if we need to do work due to change signal
			if !doWork {
				isChangeTriggered, err := app.DaemonSignalService.IsChangeTriggered()
				if err != nil {
					app.LogService.Error(fmt.Sprintf("Failed to check if change is triggered: %v", err))
					cancelFunc()
					return
				}
				if isChangeTriggered {
					app.LogService.Info("Running iteration: change triggered by UI")
					err := app.DaemonSignalService.ClearChangeTrigger()
					if err != nil {
						app.LogService.Error(fmt.Sprintf("Failed to clear change trigger: %v", err))
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
	app.LogService.Info("Processing playlists...")

	// Get acive playlists
	activePlaylists, err := app.PlaylistDB.GetActivePlaylists()
	if err != nil {
		app.LogService.Error(fmt.Sprintf("Failed to get active playlists: %v", err))
		return
	}

	// Loop over active playlists
	for _, pl := range activePlaylists {
		app.LogService.Info(fmt.Sprintf("Processing playlist: %s", pl.Name))

		// Get playlist items online
		plInfo, err := ytdlp.GetPlaylistInfoFlat(pl.URL)
		if err != nil {
			app.LogService.Error(fmt.Sprintf("Failed to get playlist info for %s: %v", pl.Name, err))
			continue
		}

		// Check which playlist items are already processed
		existingDls, err := app.DownloadDB.GetDownloadsForPlaylist(pl.ID)
		if err != nil {
			app.LogService.Error(fmt.Sprintf("Failed to get existing downloads for playlist %s: %v", pl.Name, err))
			continue
		}

		// Filter out already downloaded urls
		retryables, undownloadedUrls := getDownloadables(plInfo, existingDls)
		if len(undownloadedUrls) == 0 && len(retryables) == 0 {
			app.LogService.Debug(fmt.Sprintf("No new items or retryable to download for playlist: %s", pl.Name))
			continue
		}
		app.LogService.Info(fmt.Sprintf("Found %d new items and %d retryable items to download for playlist: %s",
			len(undownloadedUrls), len(retryables), pl.Name))

		// Retry any retryable items
		for _, dl := range retryables {
			if shouldStopIteration() {
				return
			}
			_, _ = app.DownloadService.DownloadFile(
				dl.Url, pl.SaveDirectory, pl.OutputFormat, true)
		}

		// Download any new items
		for _, url := range undownloadedUrls {
			if shouldStopIteration() {
				return
			}

			dl := download.NewDownload(pl.ID, url, pl.OutputFormat)
			_, _ = app.DownloadService.DownloadFile(
				dl.Url, pl.SaveDirectory, pl.OutputFormat, true)
		}

	}

	app.LogService.Info("Playlist processing complete.")
}

// Get undownloaded and retryable items from playlist info and existing downloads
func getDownloadables(plInfo *ytdlp.YtdlpPlaylistInfo, existingDls []download.Download) ([]download.Download, []string) {
	// Prepare return values
	retryables := make([]download.Download, 0)
	undownloadedUrls := make([]string, 0)

	// Create map of existing entries for quick lookup
	existingMap := make(map[string]bool)
	for _, existintEntry := range existingDls {
		// Add every existing item to the existing map
		existingMap[existintEntry.Url] = true

		// Add redownloadable items to result
		if existintEntry.Status == download.StFailedAutoRetry || existintEntry.Status == download.StFailedManualRetry {
			retryables = append(retryables, existintEntry)
		}
	}

	// Create download entries for new items
	for _, item := range plInfo.Entries {
		if _, exists := existingMap[item.URL]; !exists {
			undownloadedUrls = append(undownloadedUrls, item.URL)
		}
	}

	// Return results
	return retryables, undownloadedUrls
}

func shouldStopIteration() bool {
	// Check for shutdown signal
	select {
	case <-context.Background().Done():
		app.LogService.Info("Shutdown signal received, stopping downloads")
		return true
	default:
		// Check for daemon change signal
		isChangeTriggered, err := app.DaemonSignalService.IsChangeTriggered()
		if err != nil {
			app.LogService.Error(fmt.Sprintf("Failed to check if change is triggered: %v", err))
			return true
		}
		if isChangeTriggered {
			app.LogService.Info("Change triggered by UI, stopping downloads to restart iteration")
			return true
		}
	}
	return false
}
