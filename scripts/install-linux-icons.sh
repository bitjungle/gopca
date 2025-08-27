#!/bin/bash
#
# install-linux-icons.sh - Install GoPCA and GoCSV icons on Linux systems
#
# This script installs application icons and .desktop files following XDG standards
# for proper Linux desktop integration.

set -e

# Detect if running as root (for system-wide vs user installation)
if [ "$EUID" -eq 0 ]; then
    ICON_DIR="/usr/share/icons/hicolor"
    DESKTOP_DIR="/usr/share/applications"
    echo "Installing system-wide (running as root)"
else
    ICON_DIR="$HOME/.local/share/icons/hicolor"
    DESKTOP_DIR="$HOME/.local/share/applications"
    echo "Installing for current user"
fi

# Function to install icons for an application
install_app_icons() {
    local APP_NAME=$1
    local BUILD_DIR=$2
    local ICON_NAME=$3
    
    echo "Installing $APP_NAME icons..."
    
    # Create icon directories if they don't exist
    for SIZE in 16 32 48 64 128 256; do
        mkdir -p "$ICON_DIR/${SIZE}x${SIZE}/apps"
        
        # Copy icon if it exists
        if [ -f "$BUILD_DIR/linux/icon-${SIZE}.png" ]; then
            cp "$BUILD_DIR/linux/icon-${SIZE}.png" "$ICON_DIR/${SIZE}x${SIZE}/apps/${ICON_NAME}.png"
            echo "  Installed ${SIZE}x${SIZE} icon"
        fi
    done
    
    # Install the base icon as scalable
    if [ -f "$BUILD_DIR/linux/icon.png" ]; then
        mkdir -p "$ICON_DIR/scalable/apps"
        cp "$BUILD_DIR/linux/icon.png" "$ICON_DIR/scalable/apps/${ICON_NAME}.png"
        echo "  Installed scalable icon"
    fi
}

# Function to install desktop file
install_desktop_file() {
    local APP_NAME=$1
    local BUILD_DIR=$2
    local DESKTOP_FILE=$3
    local BINARY_PATH=$4
    
    echo "Installing $APP_NAME desktop file..."
    
    # Create desktop directory if it doesn't exist
    mkdir -p "$DESKTOP_DIR"
    
    # Copy and update desktop file
    if [ -f "$BUILD_DIR/linux/$DESKTOP_FILE" ]; then
        # Update Exec path in desktop file
        sed "s|Exec=/usr/local/bin/.*|Exec=$BINARY_PATH|g" \
            "$BUILD_DIR/linux/$DESKTOP_FILE" > "$DESKTOP_DIR/$DESKTOP_FILE"
        
        # Make it executable (some desktop environments require this)
        chmod +x "$DESKTOP_DIR/$DESKTOP_FILE"
        echo "  Installed desktop file"
    else
        echo "  Warning: Desktop file not found at $BUILD_DIR/linux/$DESKTOP_FILE"
    fi
}

# Main installation
main() {
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
    
    # Install GoPCA Desktop
    if [ -d "$PROJECT_ROOT/cmd/gopca-desktop/build/linux" ]; then
        echo ""
        echo "=== Installing GoPCA Desktop icons ==="
        
        # Determine binary path
        if [ -f "$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA" ]; then
            GOPCA_BIN="$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA"
        elif [ -f "/usr/local/bin/GoPCA" ]; then
            GOPCA_BIN="/usr/local/bin/GoPCA"
        else
            GOPCA_BIN="/usr/local/bin/GoPCA"  # Default expected location
            echo "  Note: GoPCA binary not found, using default path: $GOPCA_BIN"
        fi
        
        install_app_icons "GoPCA" "$PROJECT_ROOT/cmd/gopca-desktop/build" "gopca"
        install_desktop_file "GoPCA" "$PROJECT_ROOT/cmd/gopca-desktop/build" "gopca.desktop" "$GOPCA_BIN"
    else
        echo "GoPCA Desktop Linux build directory not found"
    fi
    
    # Install GoCSV
    if [ -d "$PROJECT_ROOT/cmd/gocsv/build/linux" ]; then
        echo ""
        echo "=== Installing GoCSV icons ==="
        
        # Determine binary path
        if [ -f "$PROJECT_ROOT/cmd/gocsv/build/bin/GoCSV" ]; then
            GOCSV_BIN="$PROJECT_ROOT/cmd/gocsv/build/bin/GoCSV"
        elif [ -f "/usr/local/bin/GoCSV" ]; then
            GOCSV_BIN="/usr/local/bin/GoCSV"
        else
            GOCSV_BIN="/usr/local/bin/GoCSV"  # Default expected location
            echo "  Note: GoCSV binary not found, using default path: $GOCSV_BIN"
        fi
        
        install_app_icons "GoCSV" "$PROJECT_ROOT/cmd/gocsv/build" "gocsv"
        install_desktop_file "GoCSV" "$PROJECT_ROOT/cmd/gocsv/build" "gocsv.desktop" "$GOCSV_BIN"
    else
        echo "GoCSV Linux build directory not found"
    fi
    
    # Update icon cache
    echo ""
    echo "Updating icon cache..."
    if command -v gtk-update-icon-cache &> /dev/null; then
        gtk-update-icon-cache -f "$ICON_DIR" 2>/dev/null || true
        echo "Icon cache updated"
    else
        echo "gtk-update-icon-cache not found, skipping cache update"
    fi
    
    # Update desktop database
    if command -v update-desktop-database &> /dev/null; then
        update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
        echo "Desktop database updated"
    else
        echo "update-desktop-database not found, skipping database update"
    fi
    
    echo ""
    echo "=== Installation complete ==="
    echo "Icons installed to: $ICON_DIR"
    echo "Desktop files installed to: $DESKTOP_DIR"
    echo ""
    echo "You may need to log out and back in for changes to take effect."
}

# Run main function
main "$@"