#!/usr/bin/env bash
# FastVPN Linux GUI - Installation Script (Wails v3)
# Supports Ubuntu, Debian, Linux Mint, Fedora, Arch Linux, CachyOS, openSUSE and more.
# Builds frontend, compiles desktop binary with embedded assets, installs shortcut and configures privileges.

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CURRENT_USER="${SUDO_USER:-$USER}"

echo "=========================================================="
echo "         FastVPN Linux GUI - Installer (Wails v3)         "
echo "=========================================================="

# Detect Linux Distribution
DISTRO="unknown"
if [ -f /etc/os-release ]; then
  . /etc/os-release
  DISTRO="${ID:-unknown}"
  DISTRO_LIKE="${ID_LIKE:-}"
fi

print_distro_deps_help() {
  echo ""
  echo "Missing required build or runtime dependencies."
  echo "Please install them using your distribution's package manager:"
  echo ""
  case "$DISTRO" in
    ubuntu|debian|linuxmint|pop)
      echo "  Runtime only:"
      echo "    sudo apt update && sudo apt install -y openvpn libwebkit2gtk-4.1-0 policykit-1"
      echo "  To build from source:"
      echo "    sudo apt update && sudo apt install -y golang nodejs npm build-essential libwebkit2gtk-4.1-dev libgtk-3-dev pkg-config openvpn"
      ;;
    fedora|rhel|centos)
      echo "  Runtime only:"
      echo "    sudo dnf install -y openvpn webkit2gtk4.1 polkit"
      echo "  To build from source:"
      echo "    sudo dnf install -y golang nodejs npm gcc gcc-c++ webkit2gtk4.1-devel gtk3-devel pkgconf-pkg-config openvpn"
      ;;
    arch|cachyos|manjaro|endeavouros)
      echo "  Runtime only:"
      echo "    sudo pacman -S --needed openvpn webkit2gtk-4.1 polkit"
      echo "  To build from source:"
      echo "    sudo pacman -S --needed go nodejs npm base-devel webkit2gtk-4.1 openvpn"
      ;;
    opensuse*|suse)
      echo "  Runtime only:"
      echo "    sudo zypper install -y openvpn libwebkit2gtk-4_1-0 polkit"
      echo "  To build from source:"
      echo "    sudo zypper install -y go nodejs npm patterns-devel-base-devel_basis libwebkit2gtk-4_1-devel gtk3-devel pkg-config openvpn"
      ;;
    *)
      echo "  Ensure you have: openvpn, webkit2gtk (4.1 or 4.0), gtk3, and if building: go (1.22+), nodejs, npm, gcc."
      ;;
  esac
  echo ""
}

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

# 2. Check if prebuilt release binary exists or build from source
RELEASE_BIN="$REPO_DIR/bin/fastvpn-gui"
NEEDS_BUILD=false

if [ ! -f "$RELEASE_BIN" ]; then
  NEEDS_BUILD=true
fi

# Check for build tools if build is required
if [ "$NEEDS_BUILD" = true ]; then
  if ! command -v go >/dev/null 2>&1 || ! command -v npm >/dev/null 2>&1; then
    echo "Error: 'go' or 'npm' is not installed, and no precompiled binary was found at bin/fastvpn-gui."
    print_distro_deps_help
    exit 1
  fi

  echo ""
  echo "==> [1/4] Building React Vite frontend..."
  if [ "$EUID" -eq 0 ] && [ -n "$SUDO_USER" ]; then
    sudo -u "$SUDO_USER" bash -c "cd '$REPO_DIR/frontend' && npm install && npm run build"
  else
    (cd "$REPO_DIR/frontend" && npm install && npm run build)
  fi

  echo ""
  echo "==> [2/4] Compiling FastVPN desktop binary..."
  mkdir -p "$REPO_DIR/bin"
  if [ "$EUID" -eq 0 ] && [ -n "$SUDO_USER" ]; then
    sudo -u "$SUDO_USER" bash -c "cd '$REPO_DIR' && go build -ldflags=\"-s -w\" -o bin/fastvpn-gui main.go"
  else
    (cd "$REPO_DIR" && go build -ldflags="-s -w" -o bin/fastvpn-gui main.go)
  fi
else
  echo ""
  echo "==> Found existing binary at $RELEASE_BIN. Skipping compilation."
  echo "    (To rebuild, delete bin/fastvpn-gui or run: cd frontend && npm run build && cd .. && go build -o bin/fastvpn-gui main.go)"
fi

if [ ! -f "$RELEASE_BIN" ]; then
  echo "Error: Release binary could not be found or built at $RELEASE_BIN"
  exit 1
fi

# 3. Create destination directories
echo ""
echo "==> [3/4] Installing application and assets..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/scripts"
mkdir -p "$BIN_DIR"
mkdir -p "$DESKTOP_DIR"
mkdir -p "$ICON_DIR"
mkdir -p "$PIXMAP_DIR"

# Copy binary (remove destination first to avoid 'Text file busy' if running)
cp --remove-destination -f "$RELEASE_BIN" "$INSTALL_DIR/fastvpn-gui"
chmod 755 "$INSTALL_DIR/fastvpn-gui"

# Copy bundled scripts so the app and user have local access to privilege and uninstall tools
if [ -f "$REPO_DIR/scripts/setup-privileges.sh" ]; then
  cp -f "$REPO_DIR/scripts/setup-privileges.sh" "$INSTALL_DIR/scripts/setup-privileges.sh"
  chmod 755 "$INSTALL_DIR/scripts/setup-privileges.sh"
fi
if [ -f "$REPO_DIR/scripts/uninstall.sh" ]; then
  cp -f "$REPO_DIR/scripts/uninstall.sh" "$INSTALL_DIR/scripts/uninstall.sh"
  chmod 755 "$INSTALL_DIR/scripts/uninstall.sh"
fi

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

# 4. Install desktop shortcut
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

# Check for OpenVPN runtime dependency
if ! command -v openvpn >/dev/null 2>&1; then
  echo ""
  echo "==> WARNING: 'openvpn' was not found on your system."
  echo "    FastVPN requires openvpn to establish VPN tunnels."
  print_distro_deps_help
fi

# 5. Configure privileges if running as root or print command
echo ""
echo "==> [4/4] Configuring OpenVPN privileges..."
PRIVILEGE_SCRIPT="$INSTALL_DIR/scripts/setup-privileges.sh"
if [ ! -f "$PRIVILEGE_SCRIPT" ]; then
  PRIVILEGE_SCRIPT="$REPO_DIR/scripts/setup-privileges.sh"
fi

if [ "$EUID" -eq 0 ]; then
  bash "$PRIVILEGE_SCRIPT"
else
  echo "    To enable seamless passwordless connection & termination, please run:"
  echo "    sudo bash $PRIVILEGE_SCRIPT"
fi

echo ""
echo "=========================================================="
echo "    FastVPN GUI successfully installed!                  "
echo "    Installed Location: $INSTALL_DIR/fastvpn-gui         "
echo "    Desktop Launcher:   $DESKTOP_FILE                    "
echo "    Launch via desktop menu or run: fastvpn-gui          "
echo "=========================================================="


