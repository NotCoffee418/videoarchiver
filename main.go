package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
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

// Global shutdown channel for service communication
var serviceShutdown chan struct{}

// Windows service handler
type windowsService struct {
	app *App
}

func (ws *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Initialize your app
	ws.app.startup(context.Background())

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Start your daemon loop in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		startDaemonLoop(ws.app)
	}()

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				if serviceShutdown != nil {
					close(serviceShutdown)
				}
				break loop
			default:
				// Log unexpected control request
			}
		case <-done:
			break loop
		}
	}

	changes <- svc.Status{State: svc.Stopped}
	return
}

func main() {
	mode := flag.String("mode", "", "Startup mode: ui, daemon, install-service, remove-service (defaults to ui)")
	flag.Parse()

	switch *mode {
	case "install-service":
		if err := installWindowsService(); err != nil {
			fmt.Printf("Failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully")
		os.Exit(0)

	case "remove-service":
		if err := removeWindowsService(); err != nil {
			fmt.Printf("Failed to remove service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service removed successfully")
		os.Exit(0)

	case "daemon":
		app := NewApp(false)

		// Check if we're running as a Windows service
		isWindowsService, err := svc.IsWindowsService()
		if err != nil {
			fmt.Printf("Failed to determine if running as service: %v\n", err)
			os.Exit(1)
		}

		if isWindowsService {
			// Run as Windows service
			runWindowsService(app)
		} else {
			// Run as regular daemon (Linux or manual Windows)
			runDaemon(app)
		}

	case "ui", "":
		app := NewApp(true)
		runUI(app)

	default:
		println("Invalid startup mode. Valid modes: ui, daemon, install-service, remove-service")
		os.Exit(1)
	}
}

func runWindowsService(app *App) {
	err := svc.Run(WindowsServiceName, &windowsService{app: app})
	if err != nil {
		fmt.Printf("Service failed: %v\n", err)
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

	config := mgr.Config{
		DisplayName: "Video Archiver",
		Description: "Background service for Video Archiver that handles automatic downloads",
		StartType:   mgr.StartAutomatic,
	}

	s, err := m.CreateService(
		WindowsServiceName,
		exe,
		config,
		"--mode", "daemon",
	)
	if err != nil {
		return fmt.Errorf("could not create service: %v", err)
	}
	defer s.Close()

	if err := eventlog.Install(
		WindowsServiceName,
		exe,
		false,
		eventlog.Error|eventlog.Warning|eventlog.Info,
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

	status, err := s.Control(svc.Stop)
	if err != nil {
		fmt.Printf("Warning: Could not stop service: %v\n", err)
	}

	if status.State != svc.Stopped {
		fmt.Println("Waiting for service to stop...")
		time.Sleep(5 * time.Second)
	}

	if err := s.Delete(); err != nil {
		return fmt.Errorf("could not delete service: %v", err)
	}

	if err := eventlog.Remove(WindowsServiceName); err != nil {
		fmt.Printf("Warning: Could not remove event log: %v\n", err)
	}

	return nil
}
