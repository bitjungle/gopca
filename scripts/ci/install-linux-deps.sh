#!/bin/bash
#
# install-linux-deps.sh - Robust installation of Linux dependencies for GoPCA
#
# This script handles the various webkit package versions across different Ubuntu releases
# and provides clear error messages when packages are not available.

set -e

echo "=== Installing Linux Dependencies for GoPCA ==="
echo "OS Info: $(lsb_release -d 2>/dev/null || echo 'Unknown')"

# Update package lists
echo "Updating package lists..."
sudo apt-get update

# Install common dependencies
echo "Installing common dependencies..."
sudo apt-get install -y \
    build-essential \
    pkg-config \
    libgtk-3-dev

# Try to install webkit - handle different versions
echo "Detecting and installing webkit..."

# List of webkit packages in order of preference
WEBKIT_PACKAGES=(
    "libwebkit2gtk-4.0-dev"
    "libwebkit2gtk-4.1-dev"
    "libwebkitgtk-6.0-dev"
    "webkit2gtk-driver"
)

installed=false
for pkg in "${WEBKIT_PACKAGES[@]}"; do
    echo "Checking for $pkg..."
    if apt-cache show "$pkg" &>/dev/null; then
        echo "Found $pkg - installing..."
        if sudo apt-get install -y "$pkg"; then
            echo "Successfully installed $pkg"
            installed=true
            
            # Create compatibility symlinks if needed
            if [ "$pkg" = "libwebkit2gtk-4.1-dev" ]; then
                echo "Creating compatibility symlinks for webkit 4.0..."
                # Find the actual location of the .pc file
                pc_file=$(find /usr -name "webkit2gtk-4.1.pc" 2>/dev/null | head -1)
                if [ -n "$pc_file" ]; then
                    pc_dir=$(dirname "$pc_file")
                    sudo ln -sf "$pc_file" "$pc_dir/webkit2gtk-4.0.pc" || true
                fi
            fi
            break
        else
            echo "Failed to install $pkg, trying next option..."
        fi
    fi
done

if [ "$installed" = false ]; then
    echo "ERROR: No suitable webkit package could be installed!"
    echo ""
    echo "Available webkit packages on this system:"
    apt-cache search webkit | grep -E "webkit.*gtk.*dev" || echo "No webkit packages found"
    echo ""
    echo "This usually means you're using an Ubuntu version we haven't tested yet."
    echo "Please report this issue with the output above."
    exit 1
fi

# Verify installation
echo ""
echo "=== Verification ==="
echo "Installed packages:"
dpkg -l | grep -E "(gtk-3|webkit)" | grep -E "dev" || true

echo ""
echo "=== Linux dependencies successfully installed ==="