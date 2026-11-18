#!/usr/bin/env bash
# FastVPN Linux GUI - Installation Script (Wails v3)
# Builds frontend, compiles desktop binary with embedded assets, installs shortcut and configures privileges.

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CURRENT_USER="${SUDO_USER:-$USER}"

echo "=========================================================="
echo "         FastVPN Linux GUI - Installer (Wails v3)         "
echo "=========================================================="

# 1. Determine installation paths
if [ "$EUID" -eq 0 ]; then
  INSTALL_DIR="/opt/fastvpn-gui"
  BIN_DIR="/usr/local/bin"
  DESKTOP_DIR="/usr/share/applications"
  ICON_DIR="/usr/share/icons/hicolor/512x512/apps"
  PIXMAP_DIR="/usr/share/pixmaps"
  echo "==> Running as root. Installing system-wide to: $INSTALL_DIR"
else
  INSTALL_DIR="${HOME}/.local/share/fastvpn-gui"
  BIN_DIR="${HOME}/.local/bin"
  DESKTOP_DIR="${HOME}/.local/share/applications"
  ICON_DIR="${HOME}/.local/share/icons/hicolor/512x512/apps"
  PIXMAP_DIR="${HOME}/.local/share/pixmaps"
  echo "==> Running as user ($USER). Installing to: $INSTALL_DIR"
fi

# 2. Build frontend if needed
echo ""
echo "==> [1/4] Building React Vite frontend..."
if [ "$EUID" -eq 0 ] && [ -n "$SUDO_USER" ]; then
  sudo -u "$SUDO_USER" bash -c "cd '$REPO_DIR/frontend' && npm install && npm run build"
else
  (cd "$REPO_DIR/frontend" && npm install && npm run build)
fi

# 3. Compile Go desktop binary with embedded assets
echo ""
echo "==> [2/4] Compiling FastVPN desktop binary..."
mkdir -p "$REPO_DIR/bin"
if [ "$EUID" -eq 0 ] && [ -n "$SUDO_USER" ]; then
  sudo -u "$SUDO_USER" bash -c "cd '$REPO_DIR' && go build -ldflags=\"-s -w\" -o bin/fastvpn-gui main.go"
else
  (cd "$REPO_DIR" && go build -ldflags="-s -w" -o bin/fastvpn-gui main.go)
fi

RELEASE_BIN="$REPO_DIR/bin/fastvpn-gui"
if [ ! -f "$RELEASE_BIN" ]; then
  echo "Error: Release binary could not be built at $RELEASE_BIN"
  exit 1
fi

# 4. Create destination directories
echo ""
echo "==> [3/4] Installing application and assets..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$DESKTOP_DIR"
mkdir -p "$ICON_DIR"
mkdir -p "$PIXMAP_DIR"

# Copy binary (remove destination first to avoid 'Text file busy' if running)
cp --remove-destination -f "$RELEASE_BIN" "$INSTALL_DIR/fastvpn-gui"
chmod 755 "$INSTALL_DIR/fastvpn-gui"

# Copy icon
if [ -f "$REPO_DIR/build/appicon.png" ]; then
  cp -f "$REPO_DIR/build/appicon.png" "$ICON_DIR/fastvpn-gui.png"
  cp -f "$REPO_DIR/build/appicon.png" "$PIXMAP_DIR/fastvpn-gui.png"
  cp -f "$REPO_DIR/build/appicon.png" "$INSTALL_DIR/fastvpn-gui.png"
  chmod 644 "$ICON_DIR/fastvpn-gui.png" 2>/dev/null || true
  chmod 644 "$PIXMAP_DIR/fastvpn-gui.png" 2>/dev/null || true
fi

# Create symlink in bin directory
ln -sf "$INSTALL_DIR/fastvpn-gui" "$BIN_DIR/fastvpn-gui"

# 5. Install desktop shortcut
DESKTOP_FILE="$DESKTOP_DIR/fastvpn-gui.desktop"
echo "==> Creating application launcher at $DESKTOP_FILE..."
cat <<EOF > "$DESKTOP_FILE"
[Desktop Entry]
Type=Application
Name=FastVPN GUI
GenericName=VPN Client
Comment=FastVPN OpenVPN Client for Linux
Exec=$BIN_DIR/fastvpn-gui
Icon=fastvpn-gui
Terminal=false
Categories=Network;VPN;Utility;
StartupNotify=true
StartupWMClass=fastvpn-gui
Keywords=vpn;openvpn;fastvpn;network;security;
EOF

chmod 644 "$DESKTOP_FILE"

# Update desktop and icon databases if available
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
fi
HICOLOR_DIR="$(dirname "$(dirname "$ICON_DIR")")"
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t "$HICOLOR_DIR" 2>/dev/null || true
fi
if command -v kbuildsycoca6 >/dev/null 2>&1; then
  kbuildsycoca6 --noincremental 2>/dev/null || true
elif command -v kbuildsycoca5 >/dev/null 2>&1; then
  kbuildsycoca5 --noincremental 2>/dev/null || true
fi

# 6. Configure privileges if running as root
echo ""
echo "==> [4/4] Configuring OpenVPN privileges..."
if [ "$EUID" -eq 0 ]; then
  bash "$REPO_DIR/scripts/setup-privileges.sh"
else
  echo "    To enable passwordless connection & termination, please run:"
  echo "    sudo bash $REPO_DIR/scripts/setup-privileges.sh"
fi

echo ""
echo "=========================================================="
echo "    FastVPN GUI successfully installed!                  "
echo "    Launch via desktop menu or run: fastvpn-gui          "
echo "=========================================================="

