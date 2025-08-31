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
  Note that `wails build` and `wails dev` will automatically run this command.

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

**UI Mode (default):** `go run .` or `go run . --mode ui` or `wails dev` (preferred)

- Opens desktop application with Svelte frontend

**Daemon Mode:** `go run . --mode daemon`

- Background service for automatic playlist monitoring

## Project Structure

Key components:

- `main.go` - Entry point with mode selection (`ui` or `daemon`)
- `app.go` - Main application struct and startup logic as well as wails bindings.
- `daemon.go` - Background playlist monitoring service, runs seperately and is the logic for `--daemon` mode.
- `frontend/` - Svelte UI components
- `backend/domains/` - Core services (database, download, playlist management)

## Development Commands

- `go mod tidy` - Clean dependencies
- `cd frontend && npm run build` - Build frontend assets
- `wails dev` - Development mode (set 30+ minute timeout)
- `wails build` - Production build (15-20 minutes)

## Frontend-Backend Interop

**TypeScript Definitions for Wails Bindings:**

When adding new exported functions to `app.go` that will be called from the frontend, you must also add corresponding TypeScript definitions to `frontend/src/vite-env.d.ts`.

**Process:**
1. Add your public function to `app.go` (functions that start with capital letters are automatically bound by Wails)
2. Add the corresponding TypeScript definition to `vite-env.d.ts` in the `App` interface

**TypeScript Mapping Patterns:**
- Go `func() error` → TypeScript `() => Promise<void>`
- Go `func() (string, error)` → TypeScript `() => Promise<string>`
- Go `func() ([]Type, error)` → TypeScript `() => Promise<Array<any>>` (or specific type if available)
- Go `func(param string) error` → TypeScript `(arg1: string) => Promise<void>`
- Go `func(id int, name string) error` → TypeScript `(arg1: number, arg2: string) => Promise<void>`

**Example:**
```go
// In app.go
func (a *App) GetUserName(id int) (string, error) {
    // implementation
}
```

```typescript
// In frontend/src/vite-env.d.ts
GetUserName: (arg1: number) => Promise<string>;
```

This ensures proper TypeScript support and IDE autocompletion for Wails backend calls from the Svelte frontend.

## No Testing Available

- No unit tests exist in this repository
- `go test ./...` and `npm test` will fail
- Validate changes by building and running the application

## Key Features

- **Dual-mode architecture**: UI for user interaction, daemon for background processing
- **Playlist monitoring**: Automatic detection of new videos in configured playlists
- **Download management**: yt-dlp integration with configurable formats
- **SQLite storage**: Embedded database with automatic migrations

## Locking Mechanism

The application uses a file-based locking system to coordinate between UI and daemon modes:

**Lock File Location:**
- Windows: `%LOCALAPPDATA%/videoarchiver/.lock`
- Linux: `$HOME/.local/share/videoarchiver/.lock`

**How It Works:**
- When daemon mode starts (`go run . --mode daemon`), it creates a `.lock` file containing a timestamp
- This prevents multiple daemon instances from running simultaneously
- UI mode waits for the daemon lock to be released before proceeding (up to 10 minutes timeout)
- Lock files older than 5 minutes are automatically considered stale and removed
- Lock is removed after successful daemon startup and when daemon exits

**Important for Manual Testing:**
**Always remove the lock file before manual testing if it exists.** If you interrupt daemon mode (Ctrl+C, kill process, crash), the lock file may remain and cause subsequent tests to wait unnecessarily.

**Quick lock file removal commands:**
- Windows: `del "%LOCALAPPDATA%\videoarchiver\.lock"` 
- Linux: `rm "$HOME/.local/share/videoarchiver/.lock"`
- Or use: `go run . --mode daemon` (will detect and remove stale locks automatically)
