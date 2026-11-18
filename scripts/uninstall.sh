#!/usr/bin/env bash
# FastVPN Linux GUI - Uninstallation Script
# Cleanly removes all installed application binaries, desktop launchers, icons,
# sudoers rules, Polkit permissions, and temporary runtime files.

set -e

PURGE_CONFIG=false
for arg in "$@"; do
  case "$arg" in
    --purge|-p)
      PURGE_CONFIG=true
      shift
      ;;
    --help|-h)
      echo "Usage: sudo bash $0 [--purge]"
      echo ""
      echo "Options:"
      echo "  --purge, -p   Also delete user configuration (~/.config/fastvpn-gui)"
      exit 0
      ;;
  esac
done

echo "=========================================================="
echo "        FastVPN Linux GUI - Uninstaller                   "
echo "=========================================================="

CURRENT_USER="${SUDO_USER:-$USER}"
USER_HOME=$(eval echo "~$CURRENT_USER")

# 1. Terminate running instances & VPN connection
echo "==> [1/5] Stopping running FastVPN and OpenVPN processes..."
if [ -f "/tmp/fastvpn.pid" ]; then
  FASTVPN_PID=$(cat /tmp/fastvpn.pid 2>/dev/null || true)
  if [ -n "$FASTVPN_PID" ]; then
    kill -15 "$FASTVPN_PID" 2>/dev/null || true
  fi
  rm -f "/tmp/fastvpn.pid"
fi
rm -f /tmp/fastvpn.sock /tmp/fastvpn_auth.txt /tmp/fastvpn_setup_privileges.sh 2>/dev/null || true

# 2. Remove application files
echo "==> [2/5] Removing installed binaries, launchers, and icons..."

# System-wide paths (Root or elevated)
if [ "$EUID" -eq 0 ]; then
  rm -rf "/opt/fastvpn-gui"
  rm -f "/usr/local/bin/fastvpn-gui"
  rm -f "/usr/share/applications/fastvpn-gui.desktop"
  rm -f "/usr/share/icons/hicolor/512x512/apps/fastvpn-gui.png"
  rm -f "/usr/share/pixmaps/fastvpn-gui.png"

  # Sudoers & Polkit rules
  echo "==> [3/5] Removing privilege elevation rules..."
  rm -f "/etc/sudoers.d/fastvpn-openvpn"
  rm -f "/etc/polkit-1/rules.d/50-fastvpn-openvpn.rules"

  # Update system caches
  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database /usr/share/applications 2>/dev/null || true
  fi
  if command -v gtk-update-icon-cache >/dev/null 2>&1; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor 2>/dev/null || true
  fi
elif [ -d "/opt/fastvpn-gui" ] || [ -f "/usr/share/applications/fastvpn-gui.desktop" ] || [ -f "/etc/sudoers.d/fastvpn-openvpn" ]; then
  echo "==> Notice: System-wide files detected. Removing with sudo..."
  sudo rm -rf "/opt/fastvpn-gui" "/usr/local/bin/fastvpn-gui" "/usr/share/applications/fastvpn-gui.desktop" "/usr/share/icons/hicolor/512x512/apps/fastvpn-gui.png" "/usr/share/pixmaps/fastvpn-gui.png" "/etc/sudoers.d/fastvpn-openvpn" "/etc/polkit-1/rules.d/50-fastvpn-openvpn.rules" 2>/dev/null || true
fi

# User-specific paths
rm -rf "$USER_HOME/.local/share/fastvpn-gui"
rm -f "$USER_HOME/.local/bin/fastvpn-gui"
rm -f "$USER_HOME/.local/share/applications/fastvpn-gui.desktop"
rm -f "$USER_HOME/.local/share/icons/hicolor/512x512/apps/fastvpn-gui.png"
rm -f "$USER_HOME/.local/share/pixmaps/fastvpn-gui.png"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$USER_HOME/.local/share/applications" 2>/dev/null || true
fi

# 3. Always remove VPN profiles by default
echo "==> [4/5] Removing downloaded VPN server profiles..."
if [ -d "$USER_HOME/.config/fastvpn-gui/profiles" ]; then
  rm -rf "$USER_HOME/.config/fastvpn-gui/profiles"
  echo "    Removed: $USER_HOME/.config/fastvpn-gui/profiles"
fi
if [ -d "$USER_HOME/.local/share/fastvpn-gui/profiles" ]; then
  rm -rf "$USER_HOME/.local/share/fastvpn-gui/profiles"
fi
if [ -d "/root/.config/fastvpn-gui/profiles" ]; then
  rm -rf "/root/.config/fastvpn-gui/profiles"
fi
rm -rf "/opt/fastvpn-gui/serverlistUDP" "/usr/share/fastvpn-gui/serverlistUDP" 2>/dev/null || true

# 4. Handle user configuration (credentials, favorites, settings)
if [ "$PURGE_CONFIG" = true ]; then
  echo "==> [5/5] Purging user credentials and configuration..."
  rm -rf "$USER_HOME/.config/fastvpn-gui"
  rm -rf "/root/.config/fastvpn-gui" 2>/dev/null || true
  echo "    Removed: $USER_HOME/.config/fastvpn-gui"
else
  echo "==> [5/5] Preserving user credentials at: $USER_HOME/.config/fastvpn-gui/config.json"
  echo "    (To remove credentials as well, run: sudo bash $0 --purge)"
fi

echo ""
echo "=========================================================="
echo "    FastVPN GUI has been successfully uninstalled!        "
echo "=========================================================="

