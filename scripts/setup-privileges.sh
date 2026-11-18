#!/usr/bin/env bash
# FastVPN Linux GUI - Privilege Setup Helper
# Configures passwordless execution of OpenVPN for Linux (Arch/CachyOS and Debian/Ubuntu)

set -e

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run this script with sudo or as root:"
  echo "  sudo bash $0"
  exit 1
fi

CURRENT_USER="${1:-${SUDO_USER:-$USER}}"
if [ -n "$PKEXEC_UID" ] && [ -z "$1" ]; then
  CURRENT_USER=$(id -un "$PKEXEC_UID" 2>/dev/null || echo "$CURRENT_USER")
fi
OPENVPN_BIN=$(type -P openvpn || which openvpn 2>/dev/null || echo "/usr/bin/openvpn")
PKILL_BIN=$(type -P pkill || which pkill 2>/dev/null || echo "/usr/bin/pkill")
KILL_BIN=$(type -P kill || which kill 2>/dev/null || echo "/usr/bin/kill")
KILLALL_BIN=$(type -P killall || which killall 2>/dev/null || echo "/usr/bin/killall")

# Sudoers requires fully qualified absolute paths (e.g. /usr/bin/kill instead of builtin "kill")
[[ "$OPENVPN_BIN" != /* ]] && OPENVPN_BIN="/usr/bin/openvpn"
[[ "$PKILL_BIN" != /* ]] && PKILL_BIN="/usr/bin/pkill"
[[ "$KILL_BIN" != /* ]] && KILL_BIN="/usr/bin/kill"
[[ "$KILLALL_BIN" != /* ]] && KILLALL_BIN="/usr/bin/killall"

echo "==> Configuring OpenVPN permissions for user: $CURRENT_USER"
echo "==> OpenVPN binary: $OPENVPN_BIN"

# 1. Setup Polkit rule
POLKIT_RULES_DIR="/etc/polkit-1/rules.d"
if [ -d "$POLKIT_RULES_DIR" ]; then
  echo "==> Installing Polkit rule to $POLKIT_RULES_DIR/50-fastvpn-openvpn.rules..."
  cat <<EOF > "$POLKIT_RULES_DIR/50-fastvpn-openvpn.rules"
/* Allow user to run and manage OpenVPN without password */
polkit.addRule(function(action, subject) {
    if (action.id == "org.freedesktop.policykit.exec" &&
        (action.lookup("program") == "$OPENVPN_BIN" ||
         action.lookup("program") == "$PKILL_BIN" ||
         action.lookup("program") == "$KILL_BIN" ||
         action.lookup("program") == "$KILLALL_BIN") &&
        subject.user == "$CURRENT_USER") {
        return polkit.Result.YES;
    }
});
EOF
  chmod 644 "$POLKIT_RULES_DIR/50-fastvpn-openvpn.rules"
  echo "    Polkit rule installed successfully."
fi

# 2. Setup sudoers fallback in /etc/sudoers.d/fastvpn-openvpn
SUDOERS_FILE="/etc/sudoers.d/fastvpn-openvpn"
echo "==> Installing sudoers rule to $SUDOERS_FILE..."
cat <<EOF > "$SUDOERS_FILE"
# FastVPN GUI passwordless openvpn execution and termination
$CURRENT_USER ALL=(ALL) NOPASSWD: $OPENVPN_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $PKILL_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $KILL_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $KILLALL_BIN
EOF
chmod 0440 "$SUDOERS_FILE"

# Validate sudoers syntax
if visudo -cf "$SUDOERS_FILE" >/dev/null 2>&1; then
  echo "    Sudoers syntax validated successfully."
else
  echo "Error: Sudoers file syntax invalid, removing..."
  rm -f "$SUDOERS_FILE"
  exit 1
fi

echo "==> Setup complete! FastVPN GUI can now connect and terminate VPN cleanly without password prompts."

