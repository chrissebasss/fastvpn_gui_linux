import React from "react";
import { AppConfig, ConnectionStatus, VpnProfile } from "@/types/vpn";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { getCountryFlag, cn } from "@/lib/utils";
import { APP_VERSION } from "@/version";
import {
  Search,
  Star,
  Globe2,
  Terminal,
  Settings,
  Power,
  RotateCw,
  Download,
} from "lucide-react";

interface SidebarProps {
  profiles: VpnProfile[];
  config: AppConfig | null;
  selectedProfileId: string | null;
  onSelectProfile: (profile: VpnProfile) => void;
  status: ConnectionStatus;
  protocol: "udp" | "tcp";
  onProtocolChange: (proto: "udp" | "tcp") => void;
  publicIp: string;
  isIpLoading: boolean;
  onRefreshIp: () => void;
  onOpenLogs: () => void;
  onOpenSettings: () => void;
  onOpenServerListModal?: () => void;
  onQuit: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  profiles,
  config,
  selectedProfileId,
  onSelectProfile,
  status,
  protocol,
  onProtocolChange,
  publicIp,
  isIpLoading,
  onRefreshIp,
  onOpenLogs,
  onOpenSettings,
  onOpenServerListModal,
  onQuit,
}) => {
  const [search, setSearch] = React.useState("");
  const [activeTab, setActiveTab] = React.useState<"all" | "favorites">("all");

  const favorites = React.useMemo(() => config?.favorites || [], [config]);

  const isProfileFavorite = React.useCallback(
    (profileId: string) => {
      const base = profileId.replace(/-(UDP|TCP)$/i, "");
      return favorites.some((f) => f === profileId || f.replace(/-(UDP|TCP)$/i, "") === base);
    },
    [favorites]
  );

  const favoriteCount = React.useMemo(() => {
    return profiles.filter((p) => isProfileFavorite(p.id)).length;
  }, [profiles, isProfileFavorite]);

  const filteredProfiles = React.useMemo(() => {
    return profiles
      .filter((p) => {
        if (activeTab === "favorites" && !isProfileFavorite(p.id)) {
          return false;
        }
        if (!search.trim()) return true;
        const q = search.toLowerCase();
        return (
          p.country.toLowerCase().includes(q) ||
          p.city.toLowerCase().includes(q) ||
          p.id.toLowerCase().includes(q) ||
          p.country_code.toLowerCase().includes(q)
        );
      })
      .sort((a, b) => {
        const countryDiff = a.country.localeCompare(b.country, undefined, { sensitivity: "base" });
        if (countryDiff !== 0) return countryDiff;
        const cityDiff = a.city.localeCompare(b.city, undefined, { sensitivity: "base" });
        if (cityDiff !== 0) return cityDiff;
        return a.id.localeCompare(b.id, undefined, { sensitivity: "base" });
      });
  }, [profiles, activeTab, isProfileFavorite, search]);

  const isConnected = status.status === "Connected";
  const isConnecting = status.status === "Connecting";
  const isDisconnecting = status.status === "Disconnecting";
  const activeProfileId = status.data?.profile_id;

  return (
    <aside className="flex flex-col w-[360px] min-w-[340px] max-w-[420px] h-screen bg-[#0C1018] border-r border-border select-text">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3.5 border-b border-border/80">
        <div className="flex items-center space-x-2">
          <span className="font-bold text-base tracking-wide text-foreground">
            Fast<span className="text-primary">VPN</span>
          </span>
          <span className="text-xs text-muted-foreground font-mono">v{APP_VERSION}</span>
        </div>

        {/* Global Status Badge */}
        <div>
          {isConnected ? (
            <Badge variant="connected" className="flex items-center space-x-1 shadow-sm px-2.5 py-0.5 text-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              <span>Connected</span>
            </Badge>
          ) : isConnecting ? (
            <Badge variant="accent" className="flex items-center space-x-1 px-2.5 py-0.5 text-xs">
              <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
              <span>Connecting</span>
            </Badge>
          ) : isDisconnecting ? (
            <Badge variant="secondary" className="flex items-center space-x-1 px-2.5 py-0.5 text-xs">
              <span className="w-2 h-2 rounded-full bg-red-400" />
              <span>Disconnecting</span>
            </Badge>
          ) : (
            <Badge variant="secondary" className="flex items-center space-x-1 px-2.5 py-0.5 text-xs">
              <span className="w-2 h-2 rounded-full bg-slate-500" />
              <span>Disconnected</span>
            </Badge>
          )}
        </div>
      </div>

      {/* Search Input & Protocol Toggle */}
      <div className="px-3 pt-3 pb-2 flex items-center space-x-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search country, city, server..."
            className="pl-9 bg-[#111622] border-border/80 focus-visible:ring-1 focus-visible:ring-primary h-9 text-xs"
          />
          {search && (
            <button
              onClick={() => setSearch("")}
              className="absolute right-2.5 top-2.5 text-sm text-muted-foreground hover:text-foreground"
            >
              ×
            </button>
          )}
        </div>

        {/* Protocol Switcher: [ UDP | TCP ] */}
        <div className="flex items-center bg-[#111622] p-0.5 rounded-lg border border-border/80 shrink-0">
          <button
            type="button"
            disabled={isConnected || isConnecting || isDisconnecting}
            onClick={() => onProtocolChange("udp")}
            className={cn(
              "px-2.5 py-1 text-[11px] font-bold rounded transition-all",
              protocol === "udp"
                ? "bg-primary text-white shadow-xs"
                : "text-muted-foreground hover:text-foreground",
              (isConnected || isConnecting || isDisconnecting) && "opacity-50 cursor-not-allowed"
            )}
            title={isConnected ? "Disconnect to switch protocol" : "Switch to UDP (Fastest)"}
          >
            UDP
          </button>
          <button
            type="button"
            disabled={isConnected || isConnecting || isDisconnecting}
            onClick={() => onProtocolChange("tcp")}
            className={cn(
              "px-2.5 py-1 text-[11px] font-bold rounded transition-all",
              protocol === "tcp"
                ? "bg-primary text-white shadow-xs"
                : "text-muted-foreground hover:text-foreground",
              (isConnected || isConnecting || isDisconnecting) && "opacity-50 cursor-not-allowed"
            )}
            title={isConnected ? "Disconnect to switch protocol" : "Switch to TCP (Firewall bypass)"}
          >
            TCP
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex px-3 pb-2.5 space-x-1.5">
        <button
          onClick={() => setActiveTab("all")}
          className={cn(
            "flex-1 flex items-center justify-center space-x-1.5 py-1.5 rounded-md text-xs font-medium transition-colors",
            activeTab === "all"
              ? "bg-[#182133] text-foreground border border-border/60 shadow-sm"
              : "text-muted-foreground hover:text-foreground hover:bg-[#121826]"
          )}
        >
          <Globe2 className="w-3.5 h-3.5" />
          <span>All Locations</span>
          <span className="text-xs text-muted-foreground">({profiles.length})</span>
        </button>
        <button
          onClick={() => setActiveTab("favorites")}
          className={cn(
            "flex-1 flex items-center justify-center space-x-1.5 py-1.5 rounded-md text-xs font-medium transition-colors",
            activeTab === "favorites"
              ? "bg-[#182133] text-foreground border border-border/60 shadow-sm"
              : "text-muted-foreground hover:text-foreground hover:bg-[#121826]"
          )}
        >
          <Star className="w-3.5 h-3.5 text-amber-400 fill-amber-400" />
          <span>Favorites</span>
          <span className="text-xs text-muted-foreground">({favoriteCount})</span>
        </button>
      </div>

      {/* Profile List */}
      <div className="flex-1 overflow-y-auto px-2.5 space-y-1">
        {filteredProfiles.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-48 text-center px-4">
            <Globe2 className="w-9 h-9 text-muted-foreground/40 mb-2" />
            <p className="text-sm font-medium text-muted-foreground">
              {profiles.length === 0 ? "No server profiles loaded" : "No locations found"}
            </p>
            {profiles.length === 0 ? (
              <Button
                variant="outline"
                size="sm"
                onClick={onOpenServerListModal}
                className="mt-3 text-xs border-primary/40 text-primary hover:bg-primary/10"
              >
                <Download className="w-3.5 h-3.5 mr-1.5" />
                <span>Download Server Lists</span>
              </Button>
            ) : activeTab === "favorites" ? (
              <p className="text-xs text-muted-foreground/70 mt-1">
                Click the star icon on any country to add it to your favorites.
              </p>
            ) : null}
          </div>
        ) : (
          filteredProfiles.map((profile) => {
            const isSelected = selectedProfileId === profile.id;
            const isCountryActive = isConnected && activeProfileId === profile.id;
            const isFavorite = isProfileFavorite(profile.id);

            return (
              <div
                key={profile.id}
                onClick={() => onSelectProfile(profile)}
                className={cn(
                  "flex items-center justify-between px-3 py-2.5 rounded-lg cursor-pointer transition-all border",
                  isSelected
                    ? "bg-[#172030] border-primary/60 text-foreground shadow-sm"
                    : "bg-[#0F1420]/60 border-transparent text-slate-300 hover:bg-[#141A28] hover:text-foreground"
                )}
              >
                <div className="flex items-center space-x-3 min-w-0">
                  <span className="text-2xl leading-none shrink-0">
                    {getCountryFlag(profile.country_code)}
                  </span>
                  <div className="min-w-0">
                    <div className="flex items-center space-x-1.5">
                      <span className="text-sm font-semibold truncate leading-tight">
                        {profile.country}
                      </span>
                      {profile.is_virtual && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-400 font-mono">
                          VIRTUAL
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-muted-foreground truncate leading-tight mt-0.5">
                      {profile.city}
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-2 shrink-0">
                  {isCountryActive && (
                    <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 ring-2 ring-emerald-500/30 animate-pulse" />
                  )}
                  {isFavorite && (
                    <Star className="w-4 h-4 text-amber-400 fill-amber-400 shrink-0" />
                  )}
                  <span className="text-xs font-mono text-muted-foreground bg-card-elevated px-2 py-0.5 rounded border border-border/40">
                    {profile.endpoints.length}
                  </span>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer Controls & IP Status */}
      <div className="p-3.5 border-t border-border bg-[#090D15] space-y-2.5">
        {/* Public IP Box */}
        <div className="flex items-center justify-between px-3 py-2 rounded-md bg-[#111724] border border-border/80">
          <div className="min-w-0">
            <span className="text-[10px] uppercase font-bold tracking-wider text-muted-foreground block leading-tight">
              Public IP
            </span>
            <span className="text-sm font-mono font-medium text-foreground truncate block leading-tight mt-0.5">
              {publicIp}
            </span>
          </div>
          <Button
            size="icon"
            variant="ghost"
            onClick={onRefreshIp}
            disabled={isIpLoading}
            className="h-7 w-7 text-muted-foreground hover:text-foreground"
            title="Refresh Public IP"
          >
            <RotateCw className={cn("w-4 h-4", isIpLoading && "animate-spin")} />
          </Button>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onOpenLogs}
            className="flex-1 space-x-1.5 text-xs h-8"
          >
            <Terminal className="w-4 h-4" />
            <span>Logs</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={onOpenSettings}
            className="flex-1 space-x-1.5 text-xs h-8"
          >
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </Button>

          <Button
            variant="outline"
            size="icon"
            onClick={onQuit}
            className="h-8 w-8 text-muted-foreground hover:text-destructive hover:border-destructive/40"
            title="Quit Application"
          >
            <Power className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </aside>
  );
};
