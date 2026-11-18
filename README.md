# FastVPN Linux GUI

A native, high-performance desktop application for **FastVPN** on Linux, rebuilt with **Wails v3.0.0 (Go)** and **React 18 + TypeScript + Vite + Tailwind CSS + Radix UI (shadcn/ui)**.

![FastVPN GUI Screenshot](build/appicon.png)

---

## Highlights & Features

- **Dual-Pane Master-Detail Architecture**:
  - **Left Sidebar**: Instant country search, All Locations vs Favorites filtering, country flags, server counts, live public IP detector with one-click refresh, and system controls.
  - **Right Detail View**: Country header, favorite toggle, **Auto / Random Selection Card**, and a 2-column grid of individual server gateways.
- **Dynamic Auto-Routing Gateway Highlight**:
  - Connecting via **Auto / Random** dynamically inspects the OpenVPN TLS handshake and highlights both the Auto card AND the specific server gateway handling your active traffic.
- **Persistent Session & Background VPN**:
  - When closing the application while connected, choose between **Disconnect & Quit** or **Keep VPN & Quit**.
  - **Keep VPN & Quit** preserves the active tunnel in the background. Reopening the app immediately detects the background process, restores state, and displays the connected gateway.
- **Native System Tray Integration**:
  - Full system tray support with dynamic status indicators, quick connect, disconnect, and exit handling.
- **Slide-up Log Drawer**:
  - Real-time streaming OpenVPN terminal logs with color highlighting, auto-scroll, copy, and clear functions.
- **Seamless Privilege Elevation on KDE & GNOME**:
  - Dynamic Polkit and sudoers configuration supporting KDE Plasma 6 Wayland (`kdesu`/`kdialog`), Polkit (`pkexec`), and GNOME (`zenity`).
  - No repeated password prompts required during daily connects and disconnects.
- **Built-in Web Inspector / DevTools**:
  - Open WebKit / Chrome DevTools at any time by pressing <kbd>F12</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>I</kbd>, or via the in-app button in Settings.
- **Self-Contained Native Binary**:
  - Embedded frontend assets in a single 12 MB binary with no external web server dependencies.
- **Dynamic Server Lists & Dual Protocol (UDP / TCP)**:
  - Dynamically downloads 150+ worldwide locations (`udp` and `tcp`) from `https://vpn.ncapi.io/groupedServerList.zip` to `~/.config/fastvpn-gui/profiles/`.
  - Instant protocol switching `[ UDP | TCP ]` directly from the sidebar.
  - Support for custom local server profile directories.
- **Wails v3 MCP Server Support**:
  - Native integration with the Model Context Protocol (`wails3 mcp`) over Streamable HTTP for AI pair-programming and introspection.

---

## Directory Layout

```
.
├── backend/                  # Go service layer
│   ├── config/               # App configuration & session persistence (~/.config/fastvpn-gui)
│   ├── ovpn/                 # OVPN profile parser, OpenVPN process manager & mgmt socket
│   └── service/              # Unified service facade with public IP detector & privilege API
├── frontend/                 # React 18 + TypeScript + Vite + Radix UI
│   ├── src/components/       # Sidebar, GatewayView, LogDrawer, Settings, ServerListModal, ExitDialog
│   ├── src/components/ui/    # Radix UI primitives (shadcn/ui)
│   ├── src/services/api.ts   # Wails v3 IPC client & event listeners (@wailsio/runtime)
│   ├── src/version.ts        # Frontend version constant
│   └── src/index.css         # Dark obsidian styling & responsive 2K scaling
├── scripts/                  # Automation & management scripts
│   ├── install.sh            # Complete build and desktop installation script
│   ├── uninstall.sh          # Clean uninstallation script (with optional --purge)
│   ├── setup-privileges.sh   # Polkit and Sudoers configuration helper
│   ├── update-version.sh     # SemVer version updater across Go, React, and CHANGELOG
│   ├── dev.sh                # Live hot-reloading dev runner (Air + Vite)
│   └── copy-to-shared.sh     # Sync repository to shared folder
├── build/                    # Application icons & branding assets
├── main.go                   # Wails v3 application entry point & tray
├── CHANGELOG.md              # Semantic version history (Keep a Changelog)
├── AGENTS.md                 # Agent architectural reference & instructions
└── go.mod                    # Go dependencies
```

---

## Installation & Setup

### 1. One-Click Installer

To build and install system-wide with an application launcher and desktop shortcut:

```bash
sudo bash scripts/install.sh
```

Or for user-local installation (without root):

```bash
bash scripts/install.sh
```

### 2. Configure Passwordless Privileges

To allow FastVPN to launch and stop OpenVPN without prompting for your `sudo` password:

```bash
sudo bash scripts/setup-privileges.sh
```
*(Alternatively, click "Grant Permissions Automatically" inside the application Settings).*

### 3. Uninstalling FastVPN

To cleanly remove all installed binaries, desktop launchers, icons, downloaded VPN profiles (`~/.config/fastvpn-gui/profiles`), and Polkit/sudoers rules:

```bash
sudo bash scripts/uninstall.sh
```

To **also purge** stored user credentials and settings (`~/.config/fastvpn-gui`):

```bash
sudo bash scripts/uninstall.sh --purge
```

### 4. Updating Application Version

To bump the version across `main.go`, `frontend/package.json`, `frontend/src/version.ts`, and `CHANGELOG.md`:

```bash
bash scripts/update-version.sh 1.0.1
```

---

## Wails v3 MCP Server

Wails v3 includes a built-in Model Context Protocol (MCP) server that agents and automated tooling can launch to inspect and control the running application:

```bash
wails3 mcp
```

### Server Output:
```text
wails3 MCP server ready over Streamable HTTP
endpoint=http://127.0.0.1:<PORT>/mcp
root=/path/to/fastvpn_gui_linux
bearer-token=<BEARER_TOKEN>
```

Agents can connect to the Streamable HTTP `endpoint` with the provided `bearer-token` to query window states, invoke backend methods, and monitor real-time events.

---

## Development & Live Hot Reload (Frontend & Backend)

FastVPN supports instant dual-stack development with **Vite Hot Module Replacement (HMR)** for the React UI and live reloading for the Go backend.

### Option 1: One-Command Live Dev (Recommended)

Run the included dev script:

```bash
./scripts/dev.sh
```

What this does:
1. Spawns Vite dev server on `http://localhost:5173`.
2. Connects the Wails v3 Go runtime to Vite via `FRONTEND_DEVSERVER_URL="http://localhost:5173"`.
3. Opens the desktop window with hot reloading enabled.
4. If [`air`](https://github.com/air-verse/air) is installed, Go backend changes also reload automatically on save.
5. Pressing `Ctrl+C` cleanly terminates all dev processes.

---

### Option 2: Two-Terminal Manual Setup

#### Terminal 1 — Frontend Vite Dev Server (HMR)
```bash
cd frontend
npm run dev
```

#### Terminal 2 — Go Backend Runtime
```bash
export FRONTEND_DEVSERVER_URL="http://localhost:5173"
go run main.go
```

---

## Building from Source

### Prerequisites
- Go 1.24+
- Node.js 18+ and npm
- WebKitGTK 4.1 (`webkit2gtk-4.1` on Arch/CachyOS, `libwebkit2gtk-4.1-dev` on Debian/Ubuntu)
- OpenVPN (`openvpn`)

### Running Backend Tests
```bash
go test ./backend/... -v
```

### Building Production Frontend
```bash
cd frontend
npm install
npm run build
cd ..
```

### Building Stripped Desktop Binary
```bash
go build -ldflags="-s -w" -o bin/fastvpn-gui main.go
```
