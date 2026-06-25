#!/bin/bash

set -e

echo "╔════════════════════════════════════════════════╗"
echo "║     Protoc Binary Installer (Latest)           ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
if [ "$ARCH" = "x86_64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="aarch_64"
fi

echo "Detected OS: $OS"
echo "Detected Architecture: $ARCH"
echo ""

# Latest version (update as needed)
VERSION="27.0"
DOWNLOAD_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${VERSION}/protoc-${VERSION}-${OS}-${ARCH}.zip"

echo "📥 Downloading protoc v${VERSION}..."
echo "URL: $DOWNLOAD_URL"
echo ""

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cd "$TEMP_DIR"

# Download
if ! curl -L -o protoc.zip "$DOWNLOAD_URL"; then
    echo "❌ Failed to download protoc"
    echo ""
    echo "Try downloading manually from:"
    echo "https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

echo "✓ Downloaded successfully"
echo ""

# Extract
echo "📦 Extracting..."
unzip -q protoc.zip

echo "✓ Extracted"
echo ""

# Install
echo "🔧 Installing to /usr/local..."
sudo cp bin/protoc /usr/local/bin/
sudo cp -r include/google /usr/local/include/

echo "✓ Installed successfully"
echo ""

# Verify
echo "✅ Verifying installation..."
NEW_VERSION=$(/usr/local/bin/protoc --version)
echo "Installed: $NEW_VERSION"
echo ""

# Update symlink if needed
if command -v protoc &> /dev/null; then
    CURRENT_PROTOC=$(which protoc)
    echo "Current protoc symlink: $CURRENT_PROTOC"

    if [ "$CURRENT_PROTOC" != "/usr/local/bin/protoc" ]; then
        echo "Updating /usr/bin/protoc symlink..."
        sudo rm -f /usr/bin/protoc
        sudo ln -s /usr/local/bin/protoc /usr/bin/protoc
        echo "✓ Symlink updated"
    fi
fi

echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║     ✅ Installation Complete!                 ║"
echo "╚════════════════════════════════════════════════╝"
echo ""
echo "Verify: protoc --version"
echo ""
