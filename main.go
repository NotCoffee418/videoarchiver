package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	mode := flag.String("mode", "", "Startup mode: ui, daemon, or empty (defaults to ui)")
	installService := flag.Bool("install-service", false, "Install Windows service")
	removeService := flag.Bool("remove-service", false, "Remove Windows service")
	flag.Parse()

	// Handle service installation flags
	if *installService {
		if err := installWindowsService(); err != nil {
			fmt.Printf("Failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully")
		os.Exit(0)
	}

	if *removeService {
		if err := removeWindowsService(); err != nil {
			fmt.Printf("Failed to remove service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service removed successfully")
		os.Exit(0)
	}

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
	fmt.Println("Initializing application")
	app.startup(context.Background())
	fmt.Println("Daemon starting")
	startDaemonLoop(app)
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

func installWindowsService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	// Ensure we use the installed location, not the temporary NSIS location
	targetExe := filepath.Join(filepath.Dir(exe), "videoarchiver.exe")

	config := mgr.Config{
		DisplayName: "Video Archiver",
		Description: "Background service for Video Archiver that handles automatic downloads",
		StartType:   mgr.StartAutomatic,
	}

	s, err := m.CreateService(
		WindowsServiceName,
		targetExe,
		config,
		"--mode", "daemon",
	)
	if err != nil {
		return fmt.Errorf("could not create service: %v", err)
	}
	defer s.Close()

	// Setup event logging
	if err := eventlog.Install(
		WindowsServiceName,
		targetExe,
		false, // Don't support message-file messages
		eventlog.Error|eventlog.Warning|eventlog.Info, // Support all message types
	); err != nil {
		s.Delete()
		return fmt.Errorf("could not setup event logging: %v", err)
	}

	return nil
}

func removeWindowsService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(WindowsServiceName)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	// First try to stop the service
	status, err := s.Control(svc.Stop)
	if err != nil {
		fmt.Printf("Warning: Could not stop service: %v\n", err)
		// Continue anyway to try deletion
	}

	// Wait a bit for the service to stop
	if status.State != svc.Stopped {
		fmt.Println("Waiting for service to stop...")
		time.Sleep(5 * time.Second)
	}

	if err := s.Delete(); err != nil {
		return fmt.Errorf("could not delete service: %v", err)
	}

	if err := eventlog.Remove(WindowsServiceName); err != nil {
		fmt.Printf("Warning: Could not remove event log: %v\n", err)
		// Not critical, continue
	}

	return nil
}
