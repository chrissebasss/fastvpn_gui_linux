# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-11-18

### Added
- **New Official Brand Logo**: Updated application branding across the desktop client with the new shield and WiFi logo (`logo.png`).
- **High-Resolution Icon Packaging**: Generated 512x512 squircle icons with smooth antialiased corners for Linux desktop environments (KDE Plasma, GNOME), system tray, and webview favicon.
- **Sidebar Brand Integration**: Displayed official app icon in the sidebar header alongside app version and status badge.
- **Dynamic Server List Downloader**: Replaced static bundled server profiles with dynamic 1-click downloading of official server lists directly from `https://vpn.ncapi.io/groupedServerList.zip`.
- **Dual Protocol Support (UDP & TCP)**: Added protocol switcher `[ UDP | TCP ]` in the sidebar header with instant profile list reloading (150+ locations for each protocol).
- **First-Launch Onboarding Modal**: Automatically guides users to download official server lists or configure local disk directories on initial setup.
- **Alphabetical Profile Sorting**: Server profiles are automatically sorted A-Z by country name and city across both backend and UI.
- **Server Profiles Management in Settings**: Displays profile counts, source status, formatted "Last Updated" timestamp, and a 1-click "Update Official Lists" button.
- **Custom Local Server Directories**: Support for custom user-selected paths for UDP and TCP OpenVPN configuration folders.

### Fixed
- **Public IP Parsing**: Fixed raw JSON string display (`{"ip": "...","type": "IPv4"}`) from `wmip.dev` API by implementing structured JSON decoding and IPv4 regex validation in both Go backend and React frontend fallback.
- **Connection Hang & Mutex Deadlock**: Fixed non-reentrant mutex deadlock in `ProcessManager.Disconnect()` and `StartConnection()`, allowing connections to launch immediately.
- **OpenVPN Password Prompt Freeze**: Removed `--management-query-passwords` so OpenVPN uses `--auth-user-pass` without waiting indefinitely on the TCP management socket.
- **Terminal & App Logs**: Added standard logging output (`log.Println`) so logs appear in both terminal and in-app logs viewer.
- **Wayland Font Blurriness**: Resolved subpixel rendering blurriness in dialogs and settings by replacing CSS translate transforms with flexbox centering.
- **Selectable Text**: Enabled text selection globally so users can copy IP addresses, error messages, server paths, and logs.
- **Settings Typography & Spacing**: Scaled up font sizes and input heights across Settings and Server Profiles dialogs for enhanced readability.

## [1.0.0] - 2026-11-10

### Added
- **Complete Native Rebuild**: Redesigned FastVPN GUI using Wails v3, React 18, Tailwind CSS, and Radix UI.
- **KDE Plasma 6 & Wayland Support**: Native Polkit & KDE `kdesu`/`kdialog` authorization flow with passwordless sudo configuration.
- **DevTools Integration**: Enabled WebKit Web Inspector inside the native desktop app with <kbd>F12</kbd> / <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>I</kbd> and an in-app button in Settings.
- **Wails MCP Server**: Integrated Wails v3 MCP server capabilities (`wails3 mcp`) for agentic introspection and control over Streamable HTTP.
- **Session Persistence**: Ability to keep VPN connected in the background when the GUI window is closed, with seamless state recovery on re-launch.
