#!/bin/bash

# Video Archiver Linux Installation Script
# This script downloads and installs Video Archiver on Linux systems

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${GREEN}=== Video Archiver Linux Installer ===${NC}\n"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root or with sudo!"
        print_error "Please run as a regular user."
        exit 1
    fi
}

# Detect operating system and distribution
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        
        # Detect Linux distribution
        if [[ -f /etc/os-release ]]; then
            . /etc/os-release
            DISTRO=$ID
            DISTRO_VERSION=$VERSION_ID
        elif [[ -f /etc/redhat-release ]]; then
            DISTRO="rhel"
        elif [[ -f /etc/debian_version ]]; then
            DISTRO="debian"
        else
            DISTRO="unknown"
        fi
        
        print_info "Detected Linux distribution: $DISTRO"
    else
        print_error "Unsupported operating system: $OSTYPE"
        print_error "This installer only supports Linux systems."
        exit 1
    fi
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH_SUFFIX="amd64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            print_error "Currently only x86_64 (amd64) is supported."
            exit 1
            ;;
    esac
}



# Check for required dependencies and install WebKit if needed
check_dependencies() {
    print_info "Checking for required dependencies..."
    
    # Check for curl
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed."
        print_error "Please install curl: sudo apt install curl (Ubuntu/Debian) or sudo dnf install curl (Fedora)"
        exit 1
    fi
    
    # Check for systemctl (systemd)
    if ! command -v systemctl &> /dev/null; then
        print_error "systemd is required but not found."
        print_error "This installer requires systemd for daemon management."
        exit 1
    fi
    
    print_success "Basic dependencies found."
    
    # Check and install WebKit dependencies based on distribution
    install_webkit_dependencies
}

# Install WebKit dependencies based on distribution
install_webkit_dependencies() {
    print_info "Checking WebKit dependencies..."
    
    case "$DISTRO" in
        ubuntu|debian)
            install_webkit_ubuntu_debian
            ;;
        fedora)
            install_webkit_fedora
            ;;
        rhel|centos|rocky|almalinux)
            install_webkit_rhel
            ;;
        *)
            print_warning "Unknown distribution '$DISTRO'. WebKit dependencies may need to be installed manually."
            print_warning "Required packages: webkit2gtk development libraries"
            read -p "Continue anyway? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "Installation cancelled."
                exit 0
            fi
            ;;
    esac
}

# Install WebKit for Ubuntu/Debian
install_webkit_ubuntu_debian() {
    # Check if we already have webkit installed
    if pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
        print_success "WebKit 4.1 already installed."
        return 0
    elif pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
        print_success "WebKit 4.0 already installed."
        return 0
    fi
    
    print_info "WebKit libraries not found. Installing..."
    
    # Update package list
    print_info "Updating package list..."
    if ! sudo apt-get update; then
        print_warning "Failed to update package list. Continuing..."
    fi
    
    # Try webkit2gtk-4.1 first (for newer Ubuntu versions like 24.04)
    print_info "Attempting to install libwebkit2gtk-4.1-dev..."
    if sudo apt-get install -y libwebkit2gtk-4.1-dev libgtk-3-dev build-essential pkg-config; then
        print_success "Successfully installed WebKit 4.1 dependencies."
        return 0
    fi
    
    # Fallback to webkit2gtk-4.0
    print_warning "WebKit 4.1 not available, trying WebKit 4.0..."
    if sudo apt-get install -y libwebkit2gtk-4.0-dev libgtk-3-dev build-essential pkg-config; then
        print_success "Successfully installed WebKit 4.0 dependencies."
        return 0
    fi
    
    print_error "Failed to install WebKit dependencies."
    print_error "Please install manually:"
    print_error "  Ubuntu 24.04+: sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev"
    print_error "  Ubuntu 22.04-: sudo apt install libwebkit2gtk-4.0-dev libgtk-3-dev"
    exit 1
}

# Install WebKit for Fedora
install_webkit_fedora() {
    # Check if we already have webkit installed
    if pkg-config --exists webkit2gtk-4.1 2>/dev/null || pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
        print_success "WebKit libraries already installed."
        return 0
    fi
    
    print_info "WebKit libraries not found. Installing..."
    
    # Install WebKit and development tools
    if sudo dnf install -y webkit2gtk4.1-devel gtk3-devel gcc gcc-c++ pkg-config; then
        print_success "Successfully installed WebKit dependencies."
        return 0
    fi
    
    print_error "Failed to install WebKit dependencies."
    print_error "Please install manually: sudo dnf install webkit2gtk4.1-devel gtk3-devel"
    exit 1
}

# Install WebKit for RHEL/CentOS/Rocky/AlmaLinux
install_webkit_rhel() {
    # Check if we already have webkit installed
    if pkg-config --exists webkit2gtk-4.1 2>/dev/null || pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
        print_success "WebKit libraries already installed."
        return 0
    fi
    
    print_info "WebKit libraries not found. Installing..."
    
    # Enable EPEL for additional packages
    if ! rpm -q epel-release &>/dev/null; then
        print_info "Enabling EPEL repository..."
        sudo dnf install -y epel-release || sudo yum install -y epel-release
    fi
    
    # Install WebKit and development tools
    if sudo dnf install -y webkit2gtk3-devel gtk3-devel gcc gcc-c++ pkg-config || sudo yum install -y webkit2gtk3-devel gtk3-devel gcc gcc-c++ pkg-config; then
        print_success "Successfully installed WebKit dependencies."
        return 0
    fi
    
    print_error "Failed to install WebKit dependencies."
    print_error "Please install manually with: sudo dnf install webkit2gtk3-devel gtk3-devel"
    exit 1
}

# Get latest release information
get_latest_release() {
    print_info "Getting latest release information..."
    
    # We don't need to parse version anymore, just download the latest assets
    print_info "Using latest available release assets."
}

# Create installation directory
create_install_dir() {
    INSTALL_DIR="$HOME/.local/share/videoarchiver"
    print_info "Creating installation directory: $INSTALL_DIR"
    
    mkdir -p "$INSTALL_DIR"
    
    if [[ $? -ne 0 ]]; then
        print_error "Failed to create installation directory."
        exit 1
    fi
    
    print_success "Installation directory created."
}

# Download binary
download_binary() {
    BINARY_NAME="videoarchiver-linux-${ARCH_SUFFIX}"
    DOWNLOAD_URL="https://github.com/NotCoffee418/videoarchiver/releases/latest/download/${BINARY_NAME}"
    BINARY_PATH="$INSTALL_DIR/videoarchiver"
    
    print_info "Downloading Video Archiver binary..."
    print_info "URL: $DOWNLOAD_URL"
    
    # Check if service exists and stop it before updating binary
    SERVICE_FILE="$HOME/.config/systemd/user/video-archiver.service"
    if [[ -f "$SERVICE_FILE" ]]; then
        print_info "Existing installation detected. Stopping service before update..."
        
        # Check if service is active and stop it
        if systemctl --user is-active --quiet video-archiver.service 2>/dev/null; then
            print_info "Stopping video-archiver service..."
            systemctl --user stop video-archiver.service
            
            if [[ $? -eq 0 ]]; then
                print_success "Service stopped successfully."
            else
                print_warning "Failed to stop service, but continuing with update..."
            fi
        fi
        
        # Wait a moment for the service to fully stop
        sleep 1
    fi
    
    # Download with better error handling
    HTTP_CODE=$(curl -L -w "%{http_code}" -o "$BINARY_PATH" "$DOWNLOAD_URL")
    
    if [[ "$HTTP_CODE" != "200" ]]; then
        print_error "Failed to download binary (HTTP $HTTP_CODE)."
        print_error "Please check your internet connection and try again."
        print_error "If the problem persists, the release may not be available for your architecture."
        exit 1
    fi
    
    # Verify the download was successful and file exists
    if [[ ! -f "$BINARY_PATH" ]]; then
        print_error "Binary file was not created successfully."
        exit 1
    fi
    
    # Check if the file is actually a binary (not an error page)
    if [[ $(file "$BINARY_PATH" 2>/dev/null) != *"executable"* ]]; then
        print_warning "Downloaded file may not be a valid binary."
        print_warning "Continuing anyway..."
    fi
    
    # Make binary executable
    chmod +x "$BINARY_PATH"
    
    if [[ $? -ne 0 ]]; then
        print_error "Failed to make binary executable."
        exit 1
    fi
    
    print_success "Binary downloaded and made executable."
}

# Download application icon
download_icon() {
    ICON_PATH="$INSTALL_DIR/appicon.png"
    ICON_URL="https://raw.githubusercontent.com/NotCoffee418/videoarchiver/refs/heads/main/build/appicon.png"
    
    print_info "Downloading application icon..."
    
    curl -L -o "$ICON_PATH" "$ICON_URL"
    
    if [[ $? -ne 0 ]]; then
        print_warning "Failed to download application icon. Desktop entry will use default icon."
        return 1
    fi
    
    print_success "Application icon downloaded."
    return 0
}

# Create systemd user service
create_systemd_service() {
    print_info "Creating systemd user service..."
    
    # Create systemd user directory
    SYSTEMD_DIR="$HOME/.config/systemd/user"
    mkdir -p "$SYSTEMD_DIR"
    
    SERVICE_FILE="$SYSTEMD_DIR/video-archiver.service"
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Video Archiver Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/videoarchiver --mode daemon
Restart=always
RestartSec=5
Environment=HOME=$HOME

[Install]
WantedBy=default.target
EOF

    if [[ $? -ne 0 ]]; then
        print_error "Failed to create systemd service file."
        exit 1
    fi
    
    # Reload systemd user daemon
    systemctl --user daemon-reload
    
    if [[ $? -ne 0 ]]; then
        print_error "Failed to reload systemd user daemon."
        exit 1
    fi
    
    print_success "Systemd service created."
}

# Create desktop entry
create_desktop_entry() {
    print_info "Creating desktop entry..."
    
    # Create applications directory
    DESKTOP_DIR="$HOME/.local/share/applications"
    mkdir -p "$DESKTOP_DIR"
    
    DESKTOP_FILE="$DESKTOP_DIR/videoarchiver.desktop"
    ICON_PATH="$INSTALL_DIR/appicon.png"
    
    # Use custom icon if available, otherwise use default
    if [[ -f "$ICON_PATH" ]]; then
        ICON_ENTRY="Icon=$ICON_PATH"
    else
        ICON_ENTRY="Icon=video-x-generic"
    fi
    
    cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=Video Archiver
Comment=Archive videos from playlists automatically
Exec=$INSTALL_DIR/videoarchiver
$ICON_ENTRY
Terminal=false
Categories=AudioVideo;Video;
StartupNotify=true
EOF

    if [[ $? -ne 0 ]]; then
        print_error "Failed to create desktop entry."
        exit 1
    fi
    
    # Make desktop file executable
    chmod +x "$DESKTOP_FILE"
    
    print_success "Desktop entry created."
}

# Enable and start systemd service
enable_service() {
    print_info "Enabling and starting systemd service..."
    
    # Enable service
    systemctl --user enable video-archiver.service
    
    if [[ $? -ne 0 ]]; then
        print_error "Failed to enable systemd service."
        exit 1
    fi
    
    # Start service
    systemctl --user start video-archiver.service
    
    if [[ $? -ne 0 ]]; then
        print_warning "Failed to start systemd service. You can start it manually later with:"
        print_warning "systemctl --user start video-archiver.service"
    else
        print_success "Systemd service enabled and started."
    fi
}

# Launch application
launch_app() {
    print_info "Launching Video Archiver..."
    
    # Only launch if we have a desktop environment
    if [[ -n "$DISPLAY" || -n "$WAYLAND_DISPLAY" ]]; then
        # Give the daemon service a moment to fully start
        sleep 1
        
        nohup "$INSTALL_DIR/videoarchiver" --mode ui > /dev/null 2>&1 &
        
        if [[ $? -eq 0 ]]; then
            print_success "Video Archiver launched successfully!"
            print_info "The application should appear shortly."
        else
            print_warning "Failed to launch Video Archiver automatically."
            print_info "You can start it manually from: $INSTALL_DIR/videoarchiver"
        fi
    else
        print_info "No desktop environment detected. Skipping UI launch."
        print_info "The daemon service has been started and will run in the background."
    fi
}

# Print completion message
print_completion() {
    print_header
    print_success "Video Archiver installation completed successfully!"
    echo
    print_info "Installation location: $INSTALL_DIR/videoarchiver"
    print_info "Systemd service: video-archiver.service"
    print_info "Desktop entry: Available in your application menu"
    echo
    print_info "Usage:"
    print_info "  UI Mode:     $INSTALL_DIR/videoarchiver"
    print_info "  Daemon Mode: systemctl --user start video-archiver.service"
    echo
    print_info "Service management:"
    print_info "  Start:   systemctl --user start video-archiver.service"
    print_info "  Stop:    systemctl --user stop video-archiver.service"
    print_info "  Status:  systemctl --user status video-archiver.service"
    print_info "  Logs:    journalctl --user -u video-archiver.service"
    echo
    print_info "The daemon service has been automatically started and will run on boot."
}

# Main installation function
main() {
    print_header
    
    check_root
    detect_os
    detect_arch
    check_dependencies
    get_latest_release
    create_install_dir
    download_binary
    download_icon
    create_systemd_service
    create_desktop_entry
    enable_service
    
    # Small delay to let service start
    sleep 2
    
    launch_app
    print_completion
}

# Run main function
main "$@"