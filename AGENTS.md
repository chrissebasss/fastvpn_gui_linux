# AGENTS.md - FastVPN Linux GUI

Comprehensive architectural reference, development guide, and agent interaction instructions for the FastVPN Linux GUI codebase.

---

## 1. Project Overview & Architecture

FastVPN Linux GUI is a native, high-performance desktop OpenVPN client for Linux distributions (tested on Arch/CachyOS and Ubuntu/Debian, supporting both KDE Plasma 6 Wayland and GNOME/X11).

### Core Stack
- **Desktop Framework**: [Wails v3](https://v3.wails.io) (`github.com/wailsapp/wails/v3 v3.0.0`)
- **Backend**: Go 1.24+ (native binary with embedded assets via `embed.FS`)
- **Frontend**: React 18, TypeScript, Tailwind CSS, Radix UI primitives, Lucide React icons
- **Runtime IPC**: Official `@wailsio/runtime` package (HTTP POST to `/wails/runtime` and event multiplexer)
- **VPN Engine**: Standard Linux `openvpn` with TCP management socket communication (`127.0.0.1:11194`)

### Directory Layout
```
fastvpn_gui_linux/
├── backend/
│   ├── config/             # App settings & session persistence (~/.config/fastvpn-gui)
│   ├── ovpn/               # OpenVPN process control, log monitoring & management socket
│   └── service/            # VpnService exposed to frontend RPC bindings
├── bin/                    # Compiled release binary (fastvpn-gui)
├── build/                  # Application icons (appicon.png)
├── frontend/               # React + Vite application
│   ├── src/
│   │   ├── components/     # UI components (Sidebar, GatewayView, Settings, Logs, ExitModal)
│   │   ├── services/api.ts # Wails RPC client using @wailsio/runtime
│   │   ├── types/vpn.ts    # TypeScript definitions
│   │   └── version.ts      # Frontend version constant
│   └── package.json
├── scripts/
│   ├── install.sh          # Full system-wide or user-level installer
│   ├── uninstall.sh        # Clean uninstaller with optional --purge flag
│   ├── setup-privileges.sh # Polkit and sudoers rule installer
│   ├── update-version.sh   # SemVer updater across Go, React, and CHANGELOG
│   ├── dev.sh              # Live hot-reloading dev runner (Air + Vite)
│   └── copy-to-shared.sh   # Mirror synchronization script
├── main.go                 # Wails v3 application entry point, services, tray & window
├── CHANGELOG.md            # Semantic version history (Keep a Changelog)
├── README.md               # User documentation
└── AGENTS.md               # Agent reference and operational guide
```

---

## 2. Security & Privilege Model

OpenVPN requires root privileges to configure network interfaces (`tun0`) and routing tables. FastVPN uses a passwordless execution architecture so users don't have to enter passwords for subsequent connections:

1. **Elevation Check**: `backend/ovpn/process.go` runs `sudo -n openvpn --version`. If it exits `0`, passwordless access is active.
2. **Elevation Setup**:
   - On KDE Plasma (Wayland/X11), uses `/usr/lib/kf6/kdesu` or `kdialog` with the absolute script path to invoke `scripts/setup-privileges.sh`.
   - On Polkit-capable environments, falls back to `pkexec`.
   - On GNOME/GTK, falls back to `zenity --password`.
3. **Persisted Rules**:
   - `/etc/polkit-1/rules.d/50-fastvpn-openvpn.rules` (Polkit JavaScript rule)
   - `/etc/sudoers.d/fastvpn-openvpn` (NOPASSWD for `openvpn`, `pkill`, `kill`, `killall`)

---

## 3. Development, Build & Test Commands

Always prefix shell commands with `rtk` (per user workflow rules):

### Running Unit Tests
```bash
rtk go test ./backend/... -v
```

### Building the Frontend
```bash
cd frontend && rtk npm run build
```

### Compiling Release Desktop Binary
```bash
rtk go build -ldflags="-s -w" -o bin/fastvpn-gui main.go
```

### Live Development with Hot Reload
```bash
rtk bash scripts/dev.sh
```

### Updating Version
To bump the application version across Go, React, and CHANGELOG:
```bash
rtk bash scripts/update-version.sh 1.0.1
```

### Installation and Removal
```bash
# Install
sudo bash scripts/install.sh

# Uninstall (removes binaries, launchers, icons, Polkit rules, and downloaded VPN profiles)
sudo bash scripts/uninstall.sh
# To also purge configuration and credentials:
sudo bash scripts/uninstall.sh --purge
```

---

## 4. Wails v3 MCP Server

Wails v3 includes a built-in Model Context Protocol (MCP) server that agents can launch to inspect and interact with the application:

### Starting the Server
```bash
wails3 mcp
```

### Expected Output
```text
wails3 MCP server ready over Streamable HTTP
endpoint=http://127.0.0.1:<PORT>/mcp
root=/path/to/fastvpn_gui_linux
bearer-token=<BEARER_TOKEN>
```

### How Agents Connect
1. Connect via **Streamable HTTP** to `endpoint`.
2. Pass the `bearer-token` in the HTTP `Authorization: Bearer <token>` header.
3. The MCP server provides tools to inspect registered services, view running windows, invoke bound methods, and observe application lifecycle events.

---

## 5. Web Inspector / DevTools

- **Keyboard Shortcut**: Press <kbd>F12</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>I</kbd> in the desktop window.
- **Settings UI**: Click the **"DevTools (F12)"** button in the Settings modal footer.
- **Console Global**: `window.wails` is exposed in DevTools containing `{ Call, Events, Window }` for interactive debugging.

