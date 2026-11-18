import { AppConfig, ConnectionStatus, ServerListInfo, VpnProfile } from "@/types/vpn";
import { Call, Events, Window } from "@wailsio/runtime";

if (typeof window !== "undefined") {
  (window as any).wails = { Call, Events, Window };
}

async function callBackend<T = any>(methodName: string, ...args: any[]): Promise<T> {
  console.log(`[FastVPN IPC] -> ${methodName}`, args.length > 0 ? args : "");
  const candidates = [
    `main.VpnAppService.${methodName}`,
    `main.VpnService.${methodName}`,
    `github.com/fastvpn/fastvpn-gui-linux/backend/service.VpnService.${methodName}`,
  ];

  let lastError: any = null;
  for (const fqn of candidates) {
    try {
      const result = await Call.ByName(fqn, ...args);
      console.log(`[FastVPN IPC] <- ${methodName} (${fqn}) success:`, result);
      return result;
    } catch (err: any) {
      lastError = err;
      const msg = String(err?.message || err);
      if (
        msg.includes("does not exist") ||
        msg.includes("not registered") ||
        msg.includes("undefined") ||
        msg.includes("Unknown method")
      ) {
        continue;
      }
      console.error(`[FastVPN IPC] Error calling ${methodName} (${fqn}):`, err);
      throw err;
    }
  }
  console.error(`[FastVPN IPC] Failed all candidate FQNs for ${methodName}:`, lastError);
  throw lastError || new Error(`Failed to invoke backend method: ${methodName}`);
}

export const api = {
  async getProfiles(): Promise<VpnProfile[]> {
    return await callBackend<VpnProfile[]>("GetProfiles");
  },

  async getConfig(): Promise<AppConfig> {
    return await callBackend<AppConfig>("GetConfig");
  },

  async saveCredentials(username: string, password: string): Promise<void> {
    await callBackend("SaveCredentials", username, password);
  },

  async toggleFavorite(profileId: string): Promise<string[]> {
    return await callBackend<string[]>("ToggleFavorite", profileId);
  },

  async connectVpn(profileId: string, endpoint?: string | null): Promise<void> {
    await callBackend("ConnectVPN", profileId, endpoint || "");
  },

  async disconnectVpn(): Promise<void> {
    await callBackend("DisconnectVPN");
  },

  async getConnectionStatus(): Promise<ConnectionStatus> {
    return await callBackend<ConnectionStatus>("GetConnectionStatus");
  },

  async getLogs(): Promise<string[]> {
    return await callBackend<string[]>("GetLogs");
  },

  async checkPrivileges(): Promise<boolean> {
    return await callBackend<boolean>("CheckPrivileges");
  },

  async requestPrivilegeSetup(): Promise<boolean> {
    return await callBackend<boolean>("RequestPrivilegeSetup");
  },

  async openDevTools(): Promise<void> {
    try {
      console.log("[FastVPN Frontend] Invoking Window.OpenDevTools()...");
      await Window.OpenDevTools();
    } catch (err) {
      console.warn("[FastVPN Frontend] Window.OpenDevTools failed, trying backend fallback:", err);
      try {
        await callBackend("OpenDevTools");
      } catch (e) {
        console.error("[FastVPN Frontend] Could not open DevTools:", e);
      }
    }
  },

  async getPublicIp(): Promise<string> {
    const cleanIp = (raw: string): string | null => {
      if (!raw) return null;
      let text = raw.trim();
      if (text.startsWith("{")) {
        try {
          const parsed = JSON.parse(text);
          if (parsed.ip && typeof parsed.ip === "string") {
            text = parsed.ip.trim();
          } else if (parsed.query && typeof parsed.query === "string") {
            text = parsed.query.trim();
          }
        } catch { }
      }
      const match = text.match(/\b(?:\d{1,3}\.){3}\d{1,3}\b/);
      return match ? match[0] : null;
    };

    try {
      const ip = await callBackend<string>("GetPublicIP");
      if (ip && typeof ip === "string") {
        const cleaned = cleanIp(ip);
        if (cleaned) return cleaned;
      }
    } catch (err) {
      console.warn("Backend GetPublicIP failed, trying direct fallback:", err);
    }

    // Direct browser fallback if backend probe failed
    const urls = [
      "https://wmip.dev/api/json",
      "https://api.ipify.org?format=json",
      "https://icanhazip.com",
      "https://ifconfig.me/ip",
    ];
    for (const url of urls) {
      try {
        const res = await fetch(url, { signal: AbortSignal.timeout(3000) });
        if (res.ok) {
          const text = await res.text();
          const cleaned = cleanIp(text);
          if (cleaned) return cleaned;
        }
      } catch { }
    }
    return "Unavailable";
  },

  async getServerListInfo(): Promise<ServerListInfo> {
    return callBackend<ServerListInfo>("GetServerListInfo");
  },

  async downloadServerLists(url?: string): Promise<ServerListInfo> {
    return callBackend<ServerListInfo>("DownloadServerLists", url || "");
  },

  async setProtocol(protocol: "udp" | "tcp"): Promise<VpnProfile[]> {
    return callBackend<VpnProfile[]>("SetProtocol", protocol);
  },

  async setCustomServerListPaths(udpPath: string, tcpPath: string): Promise<void> {
    await callBackend("SetCustomServerListPaths", udpPath, tcpPath);
  },

  async setServerListSource(source: "downloaded" | "custom"): Promise<VpnProfile[]> {
    return callBackend<VpnProfile[]>("SetServerListSource", source);
  },

  onServerListUpdated(callback: (info: ServerListInfo) => void): () => void {
    return Events.On("serverlist:updated", (event: any) => {
      const data = event?.data || event;
      callback(data);
    });
  },

  async forceExit(disconnect: boolean, publicIp?: string | null): Promise<void> {
    await callBackend("ForceExit", disconnect, publicIp || "");
  },

  onStatusChange(callback: (status: ConnectionStatus) => void): () => void {
    return Events.On("vpn:status", (event: any) => {
      const data = event?.data || event;
      callback(data);
    });
  },

  onLog(callback: (logLine: string) => void): () => void {
    return Events.On("vpn:log", (event: any) => {
      const line = typeof event === "string" ? event : event?.data || String(event);
      callback(line);
    });
  },

  onRequestExitConfirm(callback: () => void): () => void {
    return Events.On("request-exit-confirm", () => {
      callback();
    });
  },
};
