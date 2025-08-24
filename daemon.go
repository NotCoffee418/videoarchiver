// Daemon loop that runs in the background
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func startDaemonLoop(app *App) {
	// Create context and shutdown handling here
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutdown signal received")
		cancel()
	}()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	// Do work immediately on start
	doTheWork()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Daemon loop shutting down")
			return
		case <-ticker.C:
			doTheWork()
		}
	}
}

func doTheWork() {
	fmt.Println("Starting work...")
	// Your blocking work here
	time.Sleep(5 * time.Second) // Example work
	fmt.Println("Work completed")
}
