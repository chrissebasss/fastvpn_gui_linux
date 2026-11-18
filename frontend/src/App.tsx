import React from "react";
import { AppConfig, ConnectionStatus, ServerListInfo, UpdateInfo, VpnProfile } from "@/types/vpn";
import { api } from "@/services/api";
import { Sidebar } from "@/components/Sidebar";
import { GatewayView } from "@/components/GatewayView";
import { LogDrawer } from "@/components/LogDrawer";
import { SettingsDialog } from "@/components/SettingsDialog";
import { ExitConfirmDialog } from "@/components/ExitConfirmDialog";
import { ServerListModal } from "@/components/ServerListModal";
import { UpdateDialog } from "@/components/UpdateDialog";

export default function App() {
  const [profiles, setProfiles] = React.useState<VpnProfile[]>([]);
  const [config, setConfig] = React.useState<AppConfig | null>(null);
  const [serverListInfo, setServerListInfo] = React.useState<ServerListInfo | null>(null);
  const [protocol, setProtocol] = React.useState<"udp" | "tcp">("udp");
  const [selectedProfile, setSelectedProfile] = React.useState<VpnProfile | null>(null);
  const [status, setStatus] = React.useState<ConnectionStatus>({ status: "Disconnected" });
  const [logs, setLogs] = React.useState<string[]>([]);
  const [publicIp, setPublicIp] = React.useState<string>("Detecting...");
  const [isIpLoading, setIsIpLoading] = React.useState<boolean>(false);
  const [hasPrivileges, setHasPrivileges] = React.useState<boolean>(false);
  const [updateInfo, setUpdateInfo] = React.useState<UpdateInfo | null>(null);

  // Dialog states
  const [isLogsOpen, setIsLogsOpen] = React.useState<boolean>(false);
  const [isSettingsOpen, setIsSettingsOpen] = React.useState<boolean>(false);
  const [isServerListModalOpen, setIsServerListModalOpen] = React.useState<boolean>(false);
  const [isUpdateDialogOpen, setIsUpdateDialogOpen] = React.useState<boolean>(false);
  const [isExitConfirmOpen, setIsExitConfirmOpen] = React.useState<boolean>(false);
  const [dismissedPrivilegeBanner, setDismissedPrivilegeBanner] = React.useState<boolean>(false);

  // Initial load
  React.useEffect(() => {
    let mounted = true;

    async function initialize() {
      try {
        const [loadedProfiles, loadedConfig, initialInfo, initialStatus, initialLogs, privs] =
          await Promise.all([
            api.getProfiles().catch(() => []),
            api.getConfig().catch(() => null),
            api.getServerListInfo().catch(() => null),
            api.getConnectionStatus().catch(() => ({ status: "Disconnected" as const })),
            api.getLogs().catch(() => []),
            api.checkPrivileges().catch(() => false),
          ]);

        if (!mounted) return;

        setProfiles(loadedProfiles);
        setConfig(loadedConfig);
        setServerListInfo(initialInfo);
        if (initialInfo?.protocol) {
          setProtocol(initialInfo.protocol);
        } else if (loadedConfig?.protocol) {
          setProtocol(loadedConfig.protocol);
        }
        setStatus(initialStatus);
        setLogs(initialLogs);
        setHasPrivileges(privs);

        // Check if server lists are completely missing -> trigger onboarding modal
        if (loadedProfiles.length === 0 && (!initialInfo || !initialInfo.has_profiles)) {
          console.log("[FastVPN Frontend] No server profiles found on startup. Opening ServerListModal...");
          setIsServerListModalOpen(true);
        }

        console.log("[FastVPN Frontend] Passwordless OpenVPN Privileges detected on launch:", privs);
        if (!privs) {
          console.log("[FastVPN Frontend] Privileges not yet configured. Automatically requesting authorization...");
          handleGrantPrivileges().catch((e) => {
            console.warn("[FastVPN Frontend] Startup privilege request:", e);
          });
        }

        // Silent background update check
        api.checkForUpdates()
          .then((info) => {
            if (mounted && info) {
              setUpdateInfo(info);
            }
          })
          .catch(() => {});

        // Select initial profile
        if (loadedProfiles.length > 0) {
          const lastId = loadedConfig?.last_selected_profile;
          const found = loadedProfiles.find((p) => p.id === lastId);
          setSelectedProfile(found || loadedProfiles[0]);
        }
      } catch (err) {
        console.error("Initialization error:", err);
      }
    }

    initialize();

    // Fetch initial IP
    refreshPublicIp();

    return () => {
      mounted = false;
    };
  }, []);

  // Listen to Wails events
  React.useEffect(() => {
    const unsubStatus = api.onStatusChange((newStatus) => {
      setStatus(newStatus);
      if (newStatus.status === "Connected") {
        setTimeout(refreshPublicIp, 1200);
      }
    });

    const unsubLog = api.onLog((line) => {
      setLogs((prev) => {
        const next = [...prev, line];
        return next.length > 500 ? next.slice(-400) : next;
      });
    });

    const unsubServerList = api.onServerListUpdated((info) => {
      setServerListInfo(info);
      if (info.protocol) {
        setProtocol(info.protocol);
      }
    });

    const unsubExit = api.onRequestExitConfirm(() => {
      setStatus((current) => {
        if (current.status === "Connected") {
          setIsExitConfirmOpen(true);
        } else {
          api.forceExit(false);
        }
        return current;
      });
    });

    return () => {
      unsubStatus();
      unsubLog();
      unsubServerList();
      unsubExit();
    };
  }, []);

  const refreshPublicIp = React.useCallback(async () => {
    setIsIpLoading(true);
    try {
      const ip = await api.getPublicIp();
      setPublicIp(ip);
    } catch {
      setPublicIp("Unavailable");
    } finally {
      setIsIpLoading(false);
    }
  }, []);

  const handleSelectProfile = (profile: VpnProfile) => {
    setSelectedProfile(profile);
  };

  const handleToggleFavorite = async () => {
    if (!selectedProfile) return;
    try {
      const updatedFavorites = await api.toggleFavorite(selectedProfile.id);
      setConfig((prev) => (prev ? { ...prev, favorites: updatedFavorites } : null));
    } catch (err) {
      console.error("Failed to toggle favorite:", err);
    }
  };

  const handleConnect = async (profileId: string, endpoint?: string | null) => {
    if (!config?.username || !config?.password) {
      setIsSettingsOpen(true);
      return;
    }
    try {
      await api.connectVpn(profileId, endpoint);
    } catch (err) {
      console.error("Connect failed:", err);
    }
  };

  const handleDisconnect = async () => {
    try {
      await api.disconnectVpn();
      setTimeout(refreshPublicIp, 1200);
    } catch (err) {
      console.error("Disconnect failed:", err);
    }
  };

  const handleSaveCredentials = async (u: string, p: string) => {
    await api.saveCredentials(u, p);
    setConfig((prev) => (prev ? { ...prev, username: u, password: p } : { favorites: [], username: u, password: p }));
    const privs = await api.checkPrivileges();
    setHasPrivileges(privs);
  };

  // Global shortcut for opening Web Inspector / DevTools (F12 or Ctrl+Shift+I)
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "F12" || (e.ctrlKey && e.shiftKey && e.key === "I")) {
        console.log("[FastVPN Frontend] Opening DevTools via shortcut");
        api.openDevTools();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handleGrantPrivileges = async () => {
    try {
      console.log("[FastVPN Frontend] Invoking api.requestPrivilegeSetup()...");
      const ok = await api.requestPrivilegeSetup();
      console.log("[FastVPN Frontend] requestPrivilegeSetup returned:", ok);
      const privs = await api.checkPrivileges();
      setHasPrivileges(privs);
    } catch (err) {
      console.warn("[FastVPN Frontend] Privilege setup canceled or failed:", err);
      throw err;
    }
  };

  const handleProtocolChange = async (newProtocol: "udp" | "tcp") => {
    try {
      const updatedProfiles = await api.setProtocol(newProtocol);
      setProtocol(newProtocol);
      setProfiles(updatedProfiles);
      if (updatedProfiles.length > 0) {
        const found = selectedProfile
          ? updatedProfiles.find(
              (p) => p.id === selectedProfile.id || p.country_code === selectedProfile.country_code
            )
          : null;
        setSelectedProfile(found || updatedProfiles[0]);
      } else {
        setSelectedProfile(null);
      }
      const info = await api.getServerListInfo();
      setServerListInfo(info);
    } catch (err) {
      console.error("Failed to change protocol:", err);
    }
  };

  const handleDownloadServerLists = async (customUrl?: string) => {
    const info = await api.downloadServerLists(customUrl);
    setServerListInfo(info);
    if (info?.protocol) {
      setProtocol(info.protocol);
    }
    const updatedProfiles = await api.getProfiles();
    setProfiles(updatedProfiles);
    if (updatedProfiles.length > 0) {
      setSelectedProfile(updatedProfiles[0]);
    }
  };

  const handleSaveCustomPaths = async (udpPath: string, tcpPath: string) => {
    await api.setCustomServerListPaths(udpPath, tcpPath);
    const info = await api.getServerListInfo();
    setServerListInfo(info);
    const updatedProfiles = await api.getProfiles();
    setProfiles(updatedProfiles);
    if (updatedProfiles.length > 0) {
      setSelectedProfile(updatedProfiles[0]);
    }
  };

  const handleQuitRequest = () => {
    if (status.status === "Connected") {
      setIsExitConfirmOpen(true);
    } else {
      api.forceExit(false);
    }
  };

  const isFavorite = Boolean(
    selectedProfile &&
      config?.favorites?.some(
        (f) =>
          f === selectedProfile.id ||
          f.replace(/-(UDP|TCP)$/i, "") === selectedProfile.id.replace(/-(UDP|TCP)$/i, "")
      )
  );

  return (
    <div className="flex h-screen w-screen bg-[#090B10] text-foreground overflow-hidden font-sans">
      {/* Sidebar: Master Column */}
      <Sidebar
        profiles={profiles}
        config={config}
        selectedProfileId={selectedProfile?.id || null}
        onSelectProfile={handleSelectProfile}
        status={status}
        protocol={protocol}
        onProtocolChange={handleProtocolChange}
        publicIp={publicIp}
        isIpLoading={isIpLoading}
        onRefreshIp={refreshPublicIp}
        onOpenLogs={() => setIsLogsOpen(!isLogsOpen)}
        onOpenSettings={() => setIsSettingsOpen(true)}
        onOpenServerListModal={() => setIsServerListModalOpen(true)}
        updateInfo={updateInfo}
        onOpenUpdateDialog={() => setIsUpdateDialogOpen(true)}
        onQuit={handleQuitRequest}
      />

      {/* Detail Column */}
      <div className="flex-1 flex flex-col h-screen min-w-0">
        {!hasPrivileges && !dismissedPrivilegeBanner && (
          <div className="flex items-center justify-between px-4 py-2 bg-amber-500/10 border-b border-amber-500/30 text-amber-300 text-xs shrink-0 select-text">
            <div className="flex items-center space-x-2 truncate">
              <span className="font-semibold">Authorization needed:</span>
              <span className="truncate">Enable seamless connection without root password prompts.</span>
            </div>
            <div className="flex items-center space-x-2 shrink-0 ml-3">
              <button
                onClick={handleGrantPrivileges}
                className="px-2.5 py-1 rounded bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold text-[11px] transition-colors shadow-sm"
              >
                Grant Permissions
              </button>
              <button
                onClick={() => setDismissedPrivilegeBanner(true)}
                className="text-amber-400/70 hover:text-amber-200 text-xs px-1"
              >
                Dismiss
              </button>
            </div>
          </div>
        )}

        <GatewayView
          profile={selectedProfile}
          status={status}
          isFavorite={isFavorite}
          onToggleFavorite={handleToggleFavorite}
          onConnect={handleConnect}
          onDisconnect={handleDisconnect}
        />
      </div>

      {/* Slide-up Log Drawer */}
      <LogDrawer
        isOpen={isLogsOpen}
        onClose={() => setIsLogsOpen(false)}
        logs={logs}
        onClearLogs={() => setLogs([])}
      />

      {/* Settings Modal */}
      <SettingsDialog
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        initialUsername={config?.username || ""}
        initialPassword={config?.password || ""}
        hasPrivileges={hasPrivileges}
        serverListInfo={serverListInfo}
        updateInfo={updateInfo}
        onOpenUpdateDialog={() => {
          setIsSettingsOpen(false);
          setIsUpdateDialogOpen(true);
        }}
        onUpdateInfoChange={(info) => setUpdateInfo(info)}
        onSave={handleSaveCredentials}
        onRequestGrantPrivileges={handleGrantPrivileges}
        onDownloadServerLists={handleDownloadServerLists}
        onOpenServerListModal={() => {
          setIsSettingsOpen(false);
          setIsServerListModalOpen(true);
        }}
      />

      {/* Server List Modal */}
      <ServerListModal
        isOpen={isServerListModalOpen}
        onClose={() => setIsServerListModalOpen(false)}
        isInitialOnboarding={profiles.length === 0 && (!serverListInfo || !serverListInfo.has_profiles)}
        info={serverListInfo}
        onDownload={handleDownloadServerLists}
        onSaveCustomPaths={handleSaveCustomPaths}
      />

      {/* Exit Confirmation Modal */}
      <ExitConfirmDialog
        isOpen={isExitConfirmOpen}
        onClose={() => setIsExitConfirmOpen(false)}
        publicIp={publicIp}
        onDisconnectAndQuit={() => api.forceExit(true)}
        onKeepAndQuit={() => api.forceExit(false, publicIp)}
      />

      {/* Update Available Modal */}
      <UpdateDialog
        isOpen={isUpdateDialogOpen}
        onClose={() => setIsUpdateDialogOpen(false)}
        updateInfo={updateInfo}
      />
    </div>
  );
}
