# Video Archiver

A simple tool for archiving videos from playlists. Add videos to a playlist, and Video Archiver will automatically download them for you.

## Features

- **Playlist Monitoring**: Watches your playlists and automatically downloads new videos
- **Direct Downloads**: Manual download option for individual videos
- **Background Processing**: Runs in the background
- **Format Options**: Choose your preferred format per playlist

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/NotCoffee418/videoarchiver/releases):

- **Windows**: `videoarchiver-vX.X.X-windows-installer.exe` - Full installer with NSIS
- **Linux (amd64)**: `videoarchiver-vX.X.X-linux-amd64.tar.gz` - For x86_64 systems

### Building from Source

Requirements:
- Go 1.25+
- Node.js 20+
- Wails v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

```bash
# Clone the repository
git clone https://github.com/NotCoffee418/videoarchiver.git
cd videoarchiver

# Install frontend dependencies
cd frontend && npm install && cd ..

# Build the application
wails build

# For Windows installer
wails build -nsis
```

## Usage

1. Add a playlist URL to monitor
2. Video Archiver will check for new videos periodically
3. New videos are automatically downloaded to your specified folder
4. Use the direct download option for one-off videos

## Legal Notice

This software is intended for archiving videos you have the right to download. Users are responsible for ensuring they comply with applicable laws and terms of service.

## Third-Party Software

This software automatically downloads and uses:

- **yt-dlp** (https://github.com/yt-dlp/yt-dlp) - Licensed under The Unlicense
- **FFmpeg** (https://ffmpeg.org/) - Licensed under LGPL 2.1
 - FFmpeg source: https://github.com/FFmpeg/FFmpeg
 - LGPL license: https://www.gnu.org/licenses/old-licenses/lgpl-2.1.html