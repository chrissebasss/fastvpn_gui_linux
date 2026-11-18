import React from "react";
import { ConnectionStatus, VpnProfile } from "@/types/vpn";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { getCountryFlag, cn } from "@/lib/utils";
import {
  Shuffle,
  Star,
  CheckCircle2,
  PowerOff,
  Server,
  Zap,
} from "lucide-react";

interface GatewayViewProps {
  profile: VpnProfile | null;
  status: ConnectionStatus;
  isFavorite: boolean;
  onToggleFavorite: () => void;
  onConnect: (profileId: string, endpoint?: string | null) => void;
  onDisconnect: () => void;
}

export const GatewayView: React.FC<GatewayViewProps> = ({
  profile,
  status,
  isFavorite,
  onToggleFavorite,
  onConnect,
  onDisconnect,
}) => {
  if (!profile) {
    return (
      <main className="flex-1 flex flex-col items-center justify-center bg-[#090B10] text-muted-foreground p-8">
        <Server className="w-14 h-14 text-muted-foreground/30 mb-3" />
        <h2 className="text-lg font-semibold text-slate-300">No Location Selected</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Select a country from the left sidebar to view available gateways.
        </p>
      </main>
    );
  }

  const isConnected = status.status === "Connected";
  const isConnecting = status.status === "Connecting";
  const isProfileActive = (isConnected || isConnecting) && status.data?.profile_id === profile.id;
  const currentEndpoint = status.data?.endpoint;
  const resolvedEndpoint = status.data?.resolved_endpoint;
  const isAutoConnected = isProfileActive && (!currentEndpoint || currentEndpoint === "auto");

  // Helper to extract a display label and badge code from endpoint hostname
  const getEndpointDisplay = (endpoint: string) => {
    const host = endpoint.split(":")[0];
    const prefix = host.split(".")[0]; // e.g. "osl-c05"
    let shortCode = prefix;
    if (prefix.includes("-")) {
      const parts = prefix.split("-");
      shortCode = parts[parts.length - 1].toUpperCase(); // "C05"
    }
    return { host, prefix, shortCode };
  };

  return (
    <main className="flex-1 flex flex-col h-screen bg-[#090B10] overflow-y-auto select-text">
      {/* Country Header */}
      <div className="flex items-center justify-between px-7 py-5 border-b border-border/70 bg-[#0C1018]/50">
        <div className="flex items-center space-x-4">
          <span className="text-4xl leading-none">{getCountryFlag(profile.country_code)}</span>
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-xl font-bold text-foreground tracking-tight">
                {profile.country}
              </h1>
              {profile.is_virtual && (
                <Badge variant="secondary" className="text-xs bg-indigo-500/15 text-indigo-400 border-indigo-500/30">
                  Virtual
                </Badge>
              )}
              <Badge variant="outline" className="text-xs text-muted-foreground font-mono">
                {profile.endpoints.length} servers
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">{profile.city} • FastVPN Gateway</p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onToggleFavorite}
            className={cn(
              "space-x-1.5 text-xs h-8 border-border/80",
              isFavorite && "border-amber-500/50 text-amber-400 bg-amber-500/10"
            )}
          >
            <Star className={cn("w-4 h-4", isFavorite && "fill-amber-400 text-amber-400")} />
            <span>{isFavorite ? "Favorited" : "Favorite"}</span>
          </Button>
        </div>
      </div>

      <div className="p-7 space-y-6 max-w-6xl">
        {/* Auto / Random Selection Card */}
        <div
          className={cn(
            "relative flex items-center justify-between p-5 rounded-xl border transition-all",
            isAutoConnected
              ? "bg-[#141B2A] border-primary/70 shadow-lg shadow-primary/10"
              : "bg-[#0E131E] border-border hover:border-border/80"
          )}
        >
          <div className="flex items-center space-x-4">
            <div
              className={cn(
                "flex items-center justify-center w-12 h-12 rounded-xl border",
                isAutoConnected
                  ? "bg-primary/20 border-primary text-primary"
                  : "bg-[#161D2C] border-border text-slate-300"
              )}
            >
              <Shuffle className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center space-x-2.5">
                <h3 className="text-base font-semibold text-foreground">
                  Auto / Random Selection
                </h3>
                {isAutoConnected && (
                  <Badge variant="connected" className="flex items-center space-x-1 text-xs">
                    <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                    <span>Active Gateway</span>
                  </Badge>
                )}
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                Load-balance randomly across all {profile.endpoints.length} servers in {profile.city}.
              </p>
              {isAutoConnected && resolvedEndpoint && (
                <div className="flex items-center space-x-2 mt-2 text-sm text-emerald-400 font-mono">
                  <CheckCircle2 className="w-4 h-4" />
                  <span>
                    Routed via server: <strong className="text-primary font-bold">{resolvedEndpoint}</strong>
                  </span>
                </div>
              )}
            </div>
          </div>

          <div>
            {isAutoConnected ? (
              <Button
                variant="destructive"
                size="sm"
                onClick={onDisconnect}
                className="space-x-1.5 text-xs h-9 px-5 shadow-sm"
              >
                <PowerOff className="w-4 h-4" />
                <span>Disconnect</span>
              </Button>
            ) : (
              <Button
                variant="default"
                size="sm"
                onClick={() => onConnect(profile.id, "auto")}
                disabled={isConnecting}
                className="space-x-1.5 text-xs h-9 px-5"
              >
                <Zap className="w-4 h-4" />
                <span>Connect</span>
              </Button>
            )}
          </div>
        </div>

        {/* Individual Gateways Section */}
        <div className="space-y-3.5">
          <div className="flex items-center justify-between">
            <h3 className="text-sm uppercase font-bold tracking-wider text-muted-foreground">
              Individual Servers ({profile.endpoints.length})
            </h3>
            <span className="text-xs font-mono text-muted-foreground">UDP : 1194</span>
          </div>

          <div className="grid grid-cols-1 xl:grid-cols-2 gap-3.5">
            {profile.endpoints.map((ep) => {
              const { host, prefix, shortCode } = getEndpointDisplay(ep);

              // Server is active if directly selected OR if resolved from auto-connect
              const isDirectActive = isProfileActive && currentEndpoint === ep;
              const isAutoResolvedActive =
                isAutoConnected &&
                Boolean(resolvedEndpoint) &&
                (resolvedEndpoint === prefix || host.startsWith(resolvedEndpoint || ""));
              const isThisServerActive = isDirectActive || isAutoResolvedActive;

              return (
                <div
                  key={ep}
                  className={cn(
                    "flex items-center justify-between p-4 rounded-xl border transition-all",
                    isThisServerActive
                      ? "bg-[#141B2A] border-primary/70 shadow-md shadow-primary/15 ring-1 ring-primary/40"
                      : "bg-[#0E131E] border-border/80 hover:border-border hover:bg-[#121826]"
                  )}
                >
                  <div className="flex items-center space-x-3.5 min-w-0">
                    <div
                      className={cn(
                        "flex items-center justify-center w-10 h-10 rounded-lg font-mono text-xs font-bold border shrink-0",
                        isThisServerActive
                          ? "bg-primary/20 border-primary text-primary"
                          : "bg-[#161D2B] border-border text-slate-400"
                      )}
                    >
                      {shortCode}
                    </div>

                    <div className="min-w-0">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm font-semibold text-foreground truncate">
                          {prefix}
                        </span>
                        {isThisServerActive && (
                          <Badge variant="connected" className="text-[10px] px-1.5 py-0 h-4">
                            Active
                          </Badge>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground font-mono block truncate mt-0.5">
                        {ep}
                      </span>
                    </div>
                  </div>

                  <div className="shrink-0 ml-3">
                    {isThisServerActive ? (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={onDisconnect}
                        className="text-xs h-8 px-3.5 space-x-1"
                      >
                        <PowerOff className="w-3.5 h-3.5" />
                        <span>Disconnect</span>
                      </Button>
                    ) : (
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => onConnect(profile.id, ep)}
                        disabled={isConnecting}
                        className="text-xs h-8 px-3.5 bg-[#172030] hover:bg-primary hover:text-white border border-border"
                      >
                        Connect
                      </Button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </main>
  );
};
