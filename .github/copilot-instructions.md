# Video Archiver - GitHub Copilot Development Instructions

Video Archiver is a cross-platform desktop application built with Wails v2 that automatically downloads and archives videos from playlists. The application combines a Go backend with a Svelte frontend and can run in both UI and daemon modes.

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the information provided here.**

## Working Effectively

### Prerequisites and Installation
- Install Go 1.25.0 or later
- Install Node.js and npm
- Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Ensure `~/go/bin` is in your PATH: `export PATH=$PATH:~/go/bin`

### Development Environment Setup
- Navigate to repository root: `/home/runner/work/videoarchiver/videoarchiver`
- Install frontend dependencies: `cd frontend && npm install` -- takes 1-2 seconds
- Return to root: `cd ..`

### Build Process

**CRITICAL BUILD LIMITATION:** This application contains Windows-specific code that prevents building on Linux/macOS systems. The following files contain Windows-specific syscall code:
- `backend/domains/ytdlp/ytdlp_instancer.go`
- `backend/domains/utils/utils.go`

**Frontend Only (WORKS ON ALL PLATFORMS):**
- `cd frontend && npm run build` -- takes 2 seconds. Frontend builds successfully.
- Output: Creates `dist/index.html`, `dist/assets/*.css`, `dist/assets/*.js`

**Full Application Build (WINDOWS ONLY):**
- `wails build` -- takes 15-20 minutes on Windows. NEVER CANCEL. Set timeout to 30+ minutes.
- `wails dev` -- WINDOWS ONLY for development with live reload

**Go Binary Build (FAILS ON LINUX/MACOS):**
- `go build .` or `go run .` will fail with: "unknown field HideWindow in struct literal"
- `go vet ./...` will fail with the same error
- `go test ./...` will fail due to build errors (and no tests exist anyway)

### Running the Application

**Development Mode (Windows Only):**
- `wails dev` -- Runs with hot reload. NEVER CANCEL.

**Daemon Mode (Windows Only):**
- `go run . --mode daemon` -- Runs background service for playlist monitoring

**UI Mode (Windows Only):**
- `go run . --mode ui` or `go run .` -- Opens desktop application

### Validation and Testing

**Frontend Validation:**
- Always test frontend changes with: `cd frontend && npm run build`
- Verify build produces: `dist/index.html`, `dist/assets/*.css`, `dist/assets/*.js`

**No Unit Tests:** This repository has no test files. Commands that will fail:
- `go test ./...` -- fails due to build errors, and no test files exist anyway  
- `npm test` -- fails with "Missing script: test"

**Manual Validation (Windows Only):**
- Build and run the application to test functionality
- Test both UI mode and daemon mode
- Test basic playlist addition and video download workflows

### Linting and Code Quality
- Run `go mod tidy` to clean up dependencies -- works on all platforms
- Run `go vet ./...` (will fail on Linux/macOS with "unknown field HideWindow" error)
- Frontend linting: No specific linting tools configured

## Project Structure

### Key Directories
```
├── frontend/           # Svelte frontend application
├── backend/           # Go backend services
├── build/             # Wails build configuration
├── migrations/        # SQLite database migrations
├── data/              # Application constants
├── .vscode/           # VSCode debug/task configuration
└── wails.json         # Wails project configuration
```

### Important Files
- `main.go` - Application entry point with mode selection
- `app.go` - Main application struct and startup logic
- `daemon.go` - Background playlist monitoring service
- `frontend/package.json` - Frontend build configuration
- `wails.json` - Wails project settings

### Backend Services (`backend/domains/`)
- `db/` - Database connection and service
- `download/` - Download management and history
- `playlist/` - Playlist management and monitoring
- `settings/` - Application configuration
- `utils/` - System utilities (file operations, directory opening)
- `ytdlp/` - yt-dlp integration for video downloading

## Common Development Tasks

### Making Code Changes
- **Frontend changes:** Edit files in `frontend/src/`, then run `npm run build` to test
- **Backend changes:** Modify Go files, but note that building requires Windows
- **Database changes:** Add migrations to `migrations/` directory

### Debugging
- Use VSCode with provided launch configurations:
  - "Run Wails Dev" - Full application in development mode (Windows only)
  - "Run Daemon Mode" - Background service mode (Windows only)

### Adding Dependencies
- **Go dependencies:** `go get <package>` then `go mod tidy`
- **Frontend dependencies:** `cd frontend && npm install <package>`

## Platform-Specific Notes

### Windows Development
- Full functionality available
- Can build, run, and debug all modes
- Use Wails dev mode for best experience: `wails dev`

### Linux/macOS Development
- **LIMITED:** Can only work with frontend components
- Backend code review and analysis possible
- Cannot build or run the full application
- Use `cd frontend && npm run build` to validate frontend changes

## Time Expectations

- Frontend dependency installation: 1-2 seconds
- Frontend build: 2 seconds  
- Full Wails build (Windows): 15-20 minutes - NEVER CANCEL, set 30+ minute timeout
- Wails dev startup (Windows): 2-3 minutes - NEVER CANCEL

## Application Features

The Video Archiver provides:
- **Playlist Monitoring:** Automatically checks playlists for new videos
- **Direct Downloads:** Manual download of individual videos
- **Background Processing:** Daemon mode runs continuously
- **Format Options:** Configurable download formats per playlist
- **Desktop UI:** Modern Svelte-based interface

## Troubleshooting

### Build Failures on Linux/macOS
This is expected due to Windows-specific code. The application is designed for Windows development and deployment.

### Frontend Build Issues  
Ensure you're in the frontend directory: `cd frontend && npm install && npm run build`

### Missing Dependencies
Run `go mod tidy` for Go dependencies and `cd frontend && npm install` for frontend dependencies.

## External Dependencies

The application automatically manages:
- **yt-dlp** - Video downloading tool
- **FFmpeg** - Media processing
- **SQLite** - Embedded database

These are downloaded/updated automatically during application startup.

## Common Command Outputs

The following are outputs from frequently run commands. Reference them instead of running bash commands to save time.

### Repository Root
```
ls -la
total 132
drwxr-xr-x 10 runner docker  4096 Aug 31 00:27 .
drwxr-xr-x  3 runner docker  4096 Aug 31 00:17 ..
drwxr-xr-x  7 runner docker  4096 Aug 31 00:26 .git
-rw-r--r--  1 runner docker    14 Aug 31 00:17 .gitattributes
drwxr-xr-x  2 runner docker  4096 Aug 31 00:28 .github
-rw-r--r--  1 runner docker   575 Aug 31 00:17 .gitignore
drwxr-xr-x  2 runner docker  4096 Aug 31 00:17 .vscode
-rw-r--r--  1 runner docker  1191 Aug 31 00:17 README.md
-rw-r--r--  1 runner docker 10319 Aug 31 00:17 app.go           (403 lines)
drwxr-xr-x  5 runner docker  4096 Aug 31 00:17 backend
drwxr-xr-x  4 runner docker  4096 Aug 31 00:17 build
-rw-r--r--  1 runner docker  6420 Aug 31 00:17 daemon.go        (233 lines)
drwxr-xr-x  2 runner docker  4096 Aug 31 00:17 data
drwxr-xr-x  7 runner docker  4096 Aug 31 00:28 frontend
-rw-r--r--  1 runner docker  3306 Aug 31 00:17 go.mod
-rw-r--r--  1 runner docker 42353 Aug 31 00:17 go.sum
-rw-r--r--  1 runner docker  1108 Aug 31 00:17 main.go          (63 lines)
drwxr-xr-x  2 runner docker  4096 Aug 31 00:17 migrations
-rw-r--r--  1 runner docker   387 Aug 31 00:17 wails.json
```

### Go Module Info
```
go mod info:
module videoarchiver
go 1.25.0

Key dependencies:
- github.com/wailsapp/wails/v2 v2.10.2
- github.com/NotCoffee418/dbmigrator v0.2.4  
- modernc.org/sqlite v1.38.2
- github.com/sirupsen/logrus v1.9.3
```