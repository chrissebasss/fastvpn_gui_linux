# FastVPN Linux GUI

<p align="center">
  <img src="build/appicon.png" alt="FastVPN GUI Logo" width="128" height="128" />
</p>

<p align="center">
  <strong>Native, high-performance Linux desktop client for Namecheap FastVPN / OpenVPN.</strong>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version" /></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=black" alt="React 18" /></a>
  <a href="https://vite.dev"><img src="https://img.shields.io/badge/Vite-6.0-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite" /></a>
  <a href="https://v3.wails.io"><img src="https://img.shields.io/badge/Wails-v3.0.0-DF0000?style=flat-square&logo=wails&logoColor=white" alt="Wails v3" /></a>
  <a href="https://www.linux.org"><img src="https://img.shields.io/badge/Platform-Linux-FCC624?style=flat-square&logo=linux&logoColor=black" alt="Linux" /></a>
  <a href="https://openvpn.net"><img src="https://img.shields.io/badge/Engine-OpenVPN-EA7E20?style=flat-square&logo=openvpn&logoColor=white" alt="OpenVPN" /></a>
</p>

---

## 🌟 Overview

**FastVPN Linux GUI** is a modern, responsive, and lightweight desktop OpenVPN client engineered specifically for Linux distributions. Built using **Wails v3 (Go)** and **React 18 + TypeScript + Vite + Tailwind CSS**, it pairs the raw speed and low footprint of native Go with a clean, dark-obsidian UI.

Whether you run **Arch Linux / CachyOS**, **Ubuntu / Debian / Linux Mint**, or **Fedora**, FastVPN Linux GUI delivers seamless connection management, automatic profile downloads, and passwordless tunnel operation without frustrating permission dialogs.

---

## ✨ Features

- **Dual-Pane Master-Detail Interface**:
  - **Left Sidebar**: Real-time country search, All Locations vs. Favorites filter, country flags, server counts, public IP monitor, and protocol selector.
  - **Right Detail View**: Gateway cards with latency info, individual server endpoints, and an **Auto / Random Gateway Card**.
- **Dynamic Auto-Routing Gateway Highlight**:
  - When connecting via **Auto / Random Selection**, the app inspects the OpenVPN TLS handshake in real-time and highlights both the Auto card and the specific server gateway actively routing your traffic.
- **Cross-Protocol Favorites Sync**:
  - Favoriting a location (e.g. *United States*) seamlessly syncs across both **UDP** and **TCP** profile sets.
- **Persistent Session & Background VPN**:
  - Choose between **Disconnect & Quit** or **Keep VPN & Quit**.
  - Background tunnels stay alive when the window closes; reopening the app instantly re-attaches to the active session and streams live statistics.
- **System Tray Integration**:
  - Minimal system tray icon with quick connect, disconnect, gateway information, and exit controls.
- **Streaming Live Terminal Logs**:
  - Slide-up log drawer with real-time colored OpenVPN terminal output, auto-scroll lock, copy, and clear functionality.
- **Seamless Passwordless Privileges**:
  - One-time automated elevation setup using Polkit (`pkexec`) and standard Sudoers with strict `visudo` validation.
  - Connect and disconnect anytime without entering your system root password repeatedly.
- **Dynamic Server Profile Bundle (UDP & TCP)**:
  - 1-click download of 150+ worldwide locations directly from FastVPN's official bundle endpoint (`~/.config/fastvpn-gui/profiles`).
  - Full support for custom user-provided OpenVPN directories.
- **Built-in DevTools & MCP Server**:
  - Press <kbd>F12</kbd> or <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>I</kbd> to launch WebKit DevTools.
  - Native Model Context Protocol server (`wails3 mcp`) for AI pair programming.

---

## 📦 System Requirements & Distro Packages

FastVPN Linux GUI runs on any modern Linux distribution with standard WebKitGTK and OpenVPN packages.

### Supported Distributions Matrix

| Distribution | Runtime Dependencies | Build Prerequisites |
| :--- | :--- | :--- |
| **Ubuntu (22.04 / 24.04 LTS)**<br>**Debian 12 (Bookworm)**<br>**Linux Mint (21 / 22)**<br>**Pop!_OS** | `openvpn`<br>`libwebkit2gtk-4.1-0`<br>`policykit-1` (or `polkitd`) | `golang`<br>`nodejs` (>=18) + `npm`<br>`build-essential`<br>`libwebkit2gtk-4.1-dev`<br>`libgtk-3-dev`<br>`pkg-config` |
| **Arch Linux**<br>**CachyOS**<br>**Manjaro**<br>**EndeavourOS** | `openvpn`<br>`webkit2gtk-4.1`<br>`polkit` | `go`<br>`nodejs` (>=18) + `npm`<br>`base-devel`<br>`webkit2gtk-4.1` |
| **Fedora (39 / 40 / 41)**<br>**RHEL 9 / CentOS Stream** | `openvpn`<br>`webkit2gtk4.1`<br>`polkit` | `golang`<br>`nodejs` (>=18) + `npm`<br>`gcc` + `gcc-c++`<br>`webkit2gtk4.1-devel`<br>`gtk3-devel`<br>`pkgconf-pkg-config` |
| **openSUSE (Tumbleweed / Leap)** | `openvpn`<br>`libwebkit2gtk-4_1-0`<br>`polkit` | `go`<br>`nodejs` + `npm`<br>`patterns-devel-base-devel_basis`<br>`libwebkit2gtk-4_1-devel`<br>`gtk3-devel`<br>`pkg-config` |

---

### Installing Dependencies by Distribution

#### 🔵 Ubuntu, Debian, Linux Mint & Pop!_OS
```bash
# 1. Runtime dependencies (if using pre-compiled binary):
sudo apt update && sudo apt install -y openvpn libwebkit2gtk-4.1-0 policykit-1

# 2. Build dependencies (if compiling from source):
sudo apt update && sudo apt install -y golang nodejs npm build-essential libwebkit2gtk-4.1-dev libgtk-3-dev pkg-config openvpn
```

#### 🏹 Arch Linux, CachyOS & Manjaro
```bash
# 1. Runtime dependencies:
sudo pacman -S --needed openvpn webkit2gtk-4.1 polkit

# 2. Build dependencies:
sudo pacman -S --needed go nodejs npm base-devel webkit2gtk-4.1 openvpn
```

#### 🎩 Fedora & Red Hat Enterprise Linux
```bash
# 1. Runtime dependencies:
sudo dnf install -y openvpn webkit2gtk4.1 polkit

# 2. Build dependencies:
sudo dnf install -y golang nodejs npm gcc gcc-c++ webkit2gtk4.1-devel gtk3-devel pkgconf-pkg-config openvpn
```

#### 🦎 openSUSE (Tumbleweed / Leap)
```bash
# 1. Runtime dependencies:
sudo zypper install -y openvpn libwebkit2gtk-4_1-0 polkit

# 2. Build dependencies:
sudo zypper install -y go nodejs npm patterns-devel-base-devel_basis libwebkit2gtk-4_1-devel gtk3-devel pkg-config openvpn
```

---

## 🚀 Quick Start & Installation

Clone this repository:
```bash
git clone https://github.com/your-username/fastvpn_gui_linux.git
cd fastvpn_gui_linux
```

### Option A: User-Local Installation (Recommended)
Installs to `~/.local/share/fastvpn-gui`, creates application menu entries, and links `~/.local/bin/fastvpn-gui`:
```bash
bash scripts/install.sh
```

### Option B: System-Wide Installation
Installs to `/opt/fastvpn-gui`, creates `/usr/share/applications/fastvpn-gui.desktop`, and links `/usr/local/bin/fastvpn-gui`:
```bash
sudo bash scripts/install.sh
```

---

## 🔐 Passwordless Privilege Model

OpenVPN requires administrative privileges to create virtual network interfaces (`tun0`) and manipulate kernel routing tables. FastVPN eliminates repetitive password prompts through a transparent, secure authorization helper:

```bash
sudo bash scripts/setup-privileges.sh
```
*(You can also click **"Grant Permissions Automatically"** inside the application Settings).*

### What this script configures:
1. **Polkit JavaScript Rule**:
   - Written to `/etc/polkit-1/rules.d/50-fastvpn-openvpn.rules`.
   - Allows only your local user account to invoke OpenVPN and termination tools without a password.
2. **Sudoers Rule**:
   - Written to `/etc/sudoers.d/fastvpn-openvpn`.
   - Grants passwordless execution strictly limited to:
     ```text
     <username> ALL=(ALL) NOPASSWD: /usr/bin/openvpn
     <username> ALL=(ALL) NOPASSWD: /usr/bin/pkill
     <username> ALL=(ALL) NOPASSWD: /usr/bin/kill
     <username> ALL=(ALL) NOPASSWD: /usr/bin/killall
     ```
3. **Integrity Validation**:
   - Validated automatically with `visudo -cf` before application to guarantee sudoers file syntax safety.

---

## 🛠️ Building From Source

If you want to manually build the binary without running the full installer:

```bash
# 1. Run backend unit tests
go test ./backend/... -v

# 2. Build the React frontend
cd frontend
npm install
npm run build
cd ..

# 3. Compile the stripped, self-contained desktop binary
go build -ldflags="-s -w" -o bin/fastvpn-gui main.go
```

The resulting single binary in `bin/fastvpn-gui` has all frontend assets embedded and can be launched immediately:
```bash
./bin/fastvpn-gui
```

---

## ⚡ Live Hot-Reloading Development

FastVPN provides a synchronized dual-stack development workflow with **Vite Hot Module Replacement (HMR)** for the UI and Go binary reloading:

```bash
bash scripts/dev.sh
```

This single command:
1. Launches the Vite development server on `http://localhost:5173`.
2. Connects the Wails v3 Go runtime to Vite via `FRONTEND_DEVSERVER_URL`.
3. Opens the desktop application with live HMR. Changes to React code reflect instantly in the app window without recompiling Go.
4. If [`air`](https://github.com/air-verse/air) is installed, backend Go code changes automatically reload on save.

---

## 🤖 Wails v3 Model Context Protocol (MCP) Server

FastVPN includes a built-in MCP server for AI coding agents and automated inspection tools:

```bash
wails3 mcp
```

Example Output:
```text
wails3 MCP server ready over Streamable HTTP
endpoint=http://127.0.0.1:<PORT>/mcp
root=/path/to/fastvpn_gui_linux
bearer-token=<BEARER_TOKEN>
```

Agents can connect via Streamable HTTP with the authorization bearer token to inspect window state, call registered services, and listen to runtime events.

---

## 🧹 Uninstallation

To cleanly remove all installed binaries, desktop shortcuts, icons, and downloaded server profiles:

```bash
sudo bash scripts/uninstall.sh
```

To **also purge** stored user credentials, favorites, and settings (`~/.config/fastvpn-gui`):
```bash
sudo bash scripts/uninstall.sh --purge
```

---

## 📁 Repository Structure

```
fastvpn_gui_linux/
├── backend/                  # Go service architecture
│   ├── config/               # App configuration & session persistence (~/.config/fastvpn-gui)
│   ├── ovpn/                 # OVPN profile parser, OpenVPN process manager & mgmt socket
│   └── service/              # Unified service facade with public IP detector & privilege API
├── frontend/                 # React 18 + TypeScript + Vite + Tailwind CSS
│   ├── src/components/       # UI components (Sidebar, GatewayView, Settings, Logs, ExitModal)
│   ├── src/services/api.ts   # Wails v3 IPC client & event listeners (@wailsio/runtime)
│   ├── src/version.ts        # Frontend SemVer constant
│   └── src/index.css         # Dark obsidian styling & responsive scaling
├── scripts/                  # Shell automation & tooling
│   ├── install.sh            # Universal installer with distro package detection
│   ├── uninstall.sh          # Clean uninstaller (with optional --purge)
│   ├── setup-privileges.sh   # Polkit and Sudoers permission configuration
│   ├── dev.sh                # Dual-stack live development runner (Vite + Air)
│   ├── update-version.sh     # SemVer synchronizer across Go, React, and CHANGELOG
│   └── copy-to-shared.sh     # Mirror synchronization script
├── build/                    # Branding assets & application icon (512x512)
├── main.go                   # Wails v3 desktop application entry point & system tray
├── CHANGELOG.md              # Semantic version history (Keep a Changelog)
├── AGENTS.md                 # Agent architectural guide
└── README.md                 # Documentation
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
OpenVPN is a registered trademark of OpenVPN Inc. FastVPN is a trademark of Namecheap, Inc. This project is an independent open-source client.
