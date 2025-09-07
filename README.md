# Video Archiver

Video Archiver automatically downloads new videos from your playlists in the background. Simply add playlists to monitor, and the app handles everything else - no manual intervention needed. Works with [any platform supported by yt-dlp](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md).

![Video Archiver Preview](https://raw.githubusercontent.com/NotCoffee418/videoarchiver/refs/heads/main/preview.jpg)

## Features

- **Playlist Monitoring**: Watches your playlists and automatically downloads new videos
- **Direct Downloads**: Manual download option for individual videos
- **Background Processing**: Runs in the background
- **Format Options**: Choose your preferred format per playlist

## Installation

### Windows

Download the latest Windows installer from the [Releases page](https://github.com/NotCoffee418/videoarchiver/releases):

- **Windows**: `videoarchiver-windows-installer.exe` - Full installer with NSIS

### Linux

**Quick Install (Recommended)**

Run the installation script with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/NotCoffee418/videoarchiver/refs/heads/main/install.sh | bash
```

**Manual Installation**

Download the latest release for your platform from the [Releases page](https://github.com/NotCoffee418/videoarchiver/releases):

- **Linux (amd64)**: `videoarchiver-linux-amd64` - For x86_64 systems

## Usage

1. Add a playlist URL to monitor
2. Video Archiver will check for new videos periodically
3. New videos are automatically downloaded to your specified folder
4. Use the direct download option for one-off videos

## Building from Source

Requirements:
- Go 1.25+
- Node.js 22
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

## Legal Notice

This software is intended for archiving videos you have the right to download. Users are responsible for ensuring they comply with applicable laws and terms of service.

## Third-Party Software

This software automatically downloads and uses:

- **yt-dlp** (https://github.com/yt-dlp/yt-dlp) - Licensed under The Unlicense
- **FFmpeg** (https://ffmpeg.org/) - Licensed under LGPL 2.1
 - FFmpeg source: https://github.com/FFmpeg/FFmpeg
 - LGPL license: https://www.gnu.org/licenses/old-licenses/lgpl-2.1.html

## Troubleshooting

**Linux: Missing WebKit Dependencies**

If you encounter issues with the application not starting or displaying properly on Linux, you may need to install WebKit dependencies manually:

- Ubuntu 24.04+: `sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev`
- Ubuntu 22.04-: `sudo apt install libwebkit2gtk-4.0-dev libgtk-3-dev`  
- Fedora: `sudo dnf install webkit2gtk4.1-devel gtk3-devel`