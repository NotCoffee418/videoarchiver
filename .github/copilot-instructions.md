# Video Archiver - GitHub Copilot Development Instructions

Video Archiver is a desktop application built with Wails v2 that downloads and archives videos from playlists. **This is a single application with two runtime modes: `--mode ui` (desktop interface) and `--mode daemon` (background service).**

**Always reference these instructions first before using search or bash commands.**

## Prerequisites and Setup

- Install Go 1.25+ and Node.js/npm
- Install Wails v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Add to PATH: `export PATH=$PATH:~/go/bin`
- Frontend setup: `cd frontend && npm install` (2 seconds)

## Build Process

**Frontend (works everywhere):**

- `cd frontend && npm run build` -- Creates dist/ assets (2 seconds)

**Full Application:**

- `wails build` -- Complete application build (15-20 minutes on any platform)
- `wails dev` -- Development mode with hot reload (NEVER CANCEL)

**Windows Installer:**

- `wails build -nsis` -- Validate that the installer works if changes are made to it.

**Build Limitations:**
Some Windows-specific syscall code may cause build failures on Linux/macOS with "unknown field HideWindow" errors in:

- `backend/domains/ytdlp/ytdlp_instancer.go`
- `backend/domains/utils/utils.go`

This is a known issue that may be fixed in issue #22.

## Application Modes

**UI Mode (default):** `go run .` or `go run . --mode ui`

- Opens desktop application with Svelte frontend

**Daemon Mode:** `go run . --mode daemon`

- Background service for automatic playlist monitoring

## Project Structure

Key components:

- `main.go` - Entry point with mode selection (`ui` or `daemon`)
- `app.go` - Main application struct and startup logic
- `daemon.go` - Background playlist monitoring service
- `frontend/` - Svelte UI components
- `backend/domains/` - Core services (database, download, playlist management)

## Development Commands

- `go mod tidy` - Clean dependencies
- `cd frontend && npm run build` - Build frontend assets
- `wails dev` - Development mode (set 30+ minute timeout)
- `wails build` - Production build (15-20 minutes)

## No Testing Available

- No unit tests exist in this repository
- `go test ./...` and `npm test` will fail
- Validate changes by building and running the application

## Key Features

- **Dual-mode architecture**: UI for user interaction, daemon for background processing
- **Playlist monitoring**: Automatic detection of new videos in configured playlists
- **Download management**: yt-dlp integration with configurable formats
- **SQLite storage**: Embedded database with automatic migrations
