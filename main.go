package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"log"
	"os"

	"github.com/fastvpn/fastvpn-gui-linux/backend/ovpn"
	"github.com/fastvpn/fastvpn-gui-linux/backend/service"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var rawAssets embed.FS

//go:embed build/appicon.png
var appIcon []byte

// Version of the FastVPN GUI application
const Version = "1.1.0"

// VpnAppService wraps service.VpnService in the main package so it is accessible as main.VpnAppService
type VpnAppService struct {
	*service.VpnService
	window *application.WebviewWindow
}

func (s *VpnAppService) OpenDevTools() {
	if s.window != nil {
		s.window.OpenDevTools()
	}
}

func main() {
	distFS, err := fs.Sub(rawAssets, "frontend/dist")
	if err != nil {
		log.Fatalf("failed to locate embedded frontend: %v", err)
	}

	vpnService := service.NewVpnService()
	appService := &VpnAppService{VpnService: vpnService}

	log.Println("[FastVPN] ==================================================")
	log.Printf("[FastVPN] FastVPN Linux GUI (Wails v3) v%s", Version)
	log.Printf("[FastVPN] Desktop: %s, User: %s", os.Getenv("XDG_CURRENT_DESKTOP"), os.Getenv("USER"))
	hasPriv := vpnService.CheckPrivileges()
	if hasPriv {
		log.Println("[FastVPN] Passwordless OpenVPN Privileges: ACTIVE (sudo openvpn works without password)")
	} else {
		log.Println("[FastVPN] Passwordless OpenVPN Privileges: NOT ACTIVE (authorization required)")
	}
	log.Println("[FastVPN] ==================================================")

	app := application.New(application.Options{
		Name:        "fastvpn-gui",
		Description: "FastVPN OpenVPN Client",
		Icon:        appIcon,
		Linux: application.LinuxOptions{
			ProgramName: "fastvpn-gui",
		},
		Services: []application.Service{
			application.NewService(vpnService),
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(distFS),
		},
	})

	// Webview Desktop Window
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "FastVPN",
		Width:            1200,
		Height:           800,
		MinWidth:         960,
		MinHeight:        640,
		URL:              "/",
		DevToolsEnabled:  true,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(9, 11, 16),
	})
	appService.window = window

	// Intercept Window Closing if VPN is Connected
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		status := vpnService.GetConnectionStatus()
		if status.Status == "Connected" {
			e.Cancel()
			app.Event.Emit("request-exit-confirm")
			window.Focus()
		}
	})

	// System Tray
	systemTray := app.SystemTray.New()
	systemTray.SetIcon(appIcon)

	trayMenu := app.NewMenu()
	trayMenu.Add("Show FastVPN").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})
	trayMenu.AddSeparator()

	statusItem := trayMenu.Add("Status: Disconnected").SetEnabled(false)

	quickConnectItem := trayMenu.Add("Quick Connect (Auto)").OnClick(func(ctx *application.Context) {
		cfg, _ := vpnService.GetConfig()
		targetProfile := "NCVPN-NO-Oslo-UDP"
		if cfg != nil && cfg.LastSelectedProfile != "" {
			targetProfile = cfg.LastSelectedProfile
		}
		_ = vpnService.ConnectVPN(targetProfile, nil)
	})

	disconnectItem := trayMenu.Add("Disconnect VPN").OnClick(func(ctx *application.Context) {
		_ = vpnService.DisconnectVPN()
	})
	disconnectItem.SetEnabled(false)

	trayMenu.AddSeparator()

	trayMenu.Add("Quit FastVPN").OnClick(func(ctx *application.Context) {
		status := vpnService.GetConnectionStatus()
		if status.Status == "Connected" {
			window.Show()
			window.Focus()
			app.Event.Emit("request-exit-confirm")
		} else {
			app.Quit()
		}
	})

	systemTray.SetMenu(trayMenu)

	// Update tray menu dynamically when status changes
	vpnService.SetEventEmitter(func(event string, data interface{}) {
		app.Event.Emit(event, data)

		if event == "vpn:status" {
			if s, ok := data.(ovpn.ConnectionStatus); ok {
				statusItem.SetLabel("Status: " + s.Status)
				if s.Status == "Connected" {
					quickConnectItem.SetEnabled(false)
					disconnectItem.SetEnabled(true)
				} else {
					quickConnectItem.SetEnabled(true)
					disconnectItem.SetEnabled(false)
				}
			}
		}
	})

	// Check if restored session on launch
	currStatus := vpnService.GetConnectionStatus()
	if currStatus.Status == "Connected" {
		statusItem.SetLabel("Status: Connected")
		quickConnectItem.SetEnabled(false)
		disconnectItem.SetEnabled(true)
	}

	if err := app.Run(); err != nil {
		log.Printf("Application terminated: %v", err)
		os.Exit(1)
	}
}
