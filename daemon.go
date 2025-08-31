// Daemon loop that runs in the background
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"videoarchiver/backend/domains/download"
	"videoarchiver/backend/domains/playlist"
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
		retryables, undownloadedUrls := getDownloadables(plInfo, existingDls)
		if len(undownloadedUrls) == 0 && len(retryables) == 0 {
			fmt.Printf("No new items or retryable to download for playlist: %s\n", pl.Name)
			continue
		}
		fmt.Printf("Found %d new items and %d retryable items to download for playlist: %s\n",
			len(undownloadedUrls), len(retryables), pl.Name)

		// Retry any retryable items
		for _, dl := range retryables {
			if shouldStopIteration() {
				return
			}
			downloadItem(&dl, &pl)
		}

		// Download any new items
		for _, url := range undownloadedUrls {
			if shouldStopIteration() {
				return
			}

			dl := download.NewDownload(pl.ID, url, pl.OutputFormat)
			downloadItem(dl, &pl)
		}

	}

	fmt.Println("Playlist processing complete.")
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
		fmt.Println("Shutdown signal received, stopping downloads")
		return true
	default:
		// Check for daemon change signal
		isChangeTriggered, err := app.DaemonSignalService.IsChangeTriggered()
		if err != nil {
			fmt.Printf("Error: Failed to check if change is triggered: %v\n", err)
			return true
		}
		if isChangeTriggered {
			fmt.Println("Change triggered by UI, stopping downloads to restart iteration")
			return true
		}
	}
	return false
}

func downloadItem(dl *download.Download, pl *playlist.Playlist) {
	fmt.Printf("Downloading new item: %s\n", dl.Url)
	outputFilePath, err := app.DownloadService.DownloadFile(
		dl.Url, pl.SaveDirectory, pl.OutputFormat)
	if err != nil {
		fmt.Printf("Error: Failed to download item %s: %v\n", dl.Url, err)
		dl.SetFail(app.DownloadDB, err.Error())
	} else {
		// Calculate MD5 of downloaded file
		md5, err := download.CalculateMD5(outputFilePath)
		if err != nil {
			fmt.Printf("Error: Failed to calculate MD5 for item %s: %v\n", dl.Url, err)
			err = dl.SetFail(app.DownloadDB, fmt.Sprintf("Failed to calculate MD5: %v", err))
			// Optionally delete the file if MD5 calculation fails
			os.Remove(outputFilePath)
		} else {
			fmt.Printf("Download successful for item %s, saved to %s\n", dl.Url, outputFilePath)
			fileName := filepath.Base(outputFilePath)
			
			// Check for duplicate by MD5 hash
			existingDownload, err := app.DownloadDB.CheckDuplicateByMD5(md5, fileName, dl.PlaylistID)
			if err != nil {
				fmt.Printf("Error checking for duplicates for item %s: %v\n", dl.Url, err)
				err = dl.SetSuccess(app.DownloadDB, fileName, md5)
			} else if existingDownload != nil {
				// Check if the existing file still exists on disk
				existingFilePath := filepath.Join(pl.SaveDirectory, existingDownload.OutputFilename.String)
				if _, err := os.Stat(existingFilePath); err == nil {
					fmt.Printf("Duplicate detected for item %s (matches existing file: %s)\n", dl.Url, existingDownload.OutputFilename.String)
					// Remove the newly downloaded duplicate file
					os.Remove(outputFilePath)
					err = dl.SetSuccessDuplicate(app.DownloadDB, fileName, md5)
				} else {
					// Existing file no longer exists, proceed as normal download
					fmt.Printf("Existing file %s no longer exists, proceeding as new download\n", existingFilePath)
					err = dl.SetSuccess(app.DownloadDB, fileName, md5)
				}
			} else {
				// No duplicate found, proceed as normal
				err = dl.SetSuccess(app.DownloadDB, fileName, md5)
			}
		}

		// Handle DB errors
		if err != nil {
			fmt.Printf("failed to update database after download %s: %v\n", dl.Url, err)
		}
	}
}
