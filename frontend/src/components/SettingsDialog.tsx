import React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Eye, EyeOff, ShieldCheck, ShieldAlert, Copy, Check, Terminal, Globe, RefreshCw, FolderOpen, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { api } from "@/services/api";
import { ServerListInfo } from "@/types/vpn";

interface SettingsDialogProps {
  isOpen: boolean;
  onClose: () => void;
  initialUsername?: string;
  initialPassword?: string;
  hasPrivileges: boolean;
  serverListInfo?: ServerListInfo | null;
  onSave: (username: string, password: string) => Promise<void>;
  onRequestGrantPrivileges?: () => Promise<void>;
  onDownloadServerLists?: () => Promise<void>;
  onOpenServerListModal?: () => void;
}

export const SettingsDialog: React.FC<SettingsDialogProps> = ({
  isOpen,
  onClose,
  initialUsername = "",
  initialPassword = "",
  hasPrivileges,
  serverListInfo,
  onSave,
  onRequestGrantPrivileges,
  onDownloadServerLists,
  onOpenServerListModal,
}) => {
  const [username, setUsername] = React.useState(initialUsername);
  const [password, setPassword] = React.useState(initialPassword);
  const [showPassword, setShowPassword] = React.useState(false);
  const [isSaving, setIsSaving] = React.useState(false);
  const [isGranting, setIsGranting] = React.useState(false);
  const [isUpdatingLists, setIsUpdatingLists] = React.useState(false);
  const [updateFeedback, setUpdateFeedback] = React.useState<"idle" | "success" | "error">("idle");
  const [copiedCmd, setCopiedCmd] = React.useState(false);
  const [setupCmd, setSetupCmd] = React.useState("sudo bash scripts/setup-privileges.sh");
  const [grantFeedback, setGrantFeedback] = React.useState<{ status: "idle" | "success" | "error"; msg: string }>({
    status: "idle",
    msg: "",
  });

  React.useEffect(() => {
    setUsername(initialUsername);
    setPassword(initialPassword);
    setGrantFeedback({ status: "idle", msg: "" });
    if (isOpen) {
      api.getPrivilegeSetupCommand().then((cmd) => {
        if (cmd) setSetupCmd(cmd);
      });
    }
  }, [initialUsername, initialPassword, isOpen]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      await onSave(username, password);
      onClose();
    } finally {
      setIsSaving(false);
    }
  };

  const handleCopyCmd = () => {
    navigator.clipboard.writeText(setupCmd);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg bg-[#0D111A] border-border text-foreground max-h-[88vh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="p-5 pb-3 border-b border-border/70 shrink-0">
          <DialogTitle className="text-lg font-bold text-foreground">
            FastVPN Settings
          </DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground leading-normal mt-1">
            Configure your FastVPN service credentials and Linux security privileges.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSave} className="flex flex-col flex-1 min-h-0 overflow-hidden">
          <div className="flex-1 overflow-y-auto p-5 space-y-4">
            {/* Credentials */}
            <div className="space-y-3.5">
              <div className="space-y-1.5">
                <label className="text-sm font-semibold text-slate-200">
                  FastVPN Username / Email
                </label>
                <Input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="vpn username"
                  className="bg-[#121824] border-border text-sm h-10 px-3"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-semibold text-slate-200">
                  FastVPN Password
                </label>
                <div className="relative">
                  <Input
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••••••"
                    className="bg-[#121824] border-border pr-10 text-sm h-10 px-3"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            </div>

          {/* Server Profiles & Protocol Section */}
          <div className="p-3.5 rounded-lg bg-[#111724] border border-border/80 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold text-slate-200 flex items-center space-x-2">
                <Globe className="w-4 h-4 text-primary" />
                <span>Server Profiles</span>
              </span>
              <Badge variant="outline" className="text-xs px-2.5 py-0.5 text-slate-300 border-border">
                {serverListInfo?.server_list_source === "custom" ? "Custom Folders" : "Official Lists"}
              </Badge>
            </div>

            <div className="text-xs sm:text-sm text-slate-300 space-y-2 select-text">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Available Profiles:</span>
                <span className="font-mono text-slate-200 font-medium">
                  {serverListInfo ? `${serverListInfo.udp_count} UDP, ${serverListInfo.tcp_count} TCP` : "0"}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Last Updated:</span>
                <span className="font-mono text-slate-200 font-medium">
                  {serverListInfo?.last_server_list_update
                    ? new Date(serverListInfo.last_server_list_update * 1000).toLocaleString()
                    : "Not yet updated"}
                </span>
              </div>
              <div className="p-2.5 rounded bg-[#090C14] border border-border/80 flex items-center justify-between font-mono text-xs text-slate-300">
                <span className="truncate mr-2 select-all">https://vpn.ncapi.io/groupedServerList.zip</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => {
                    navigator.clipboard.writeText("https://vpn.ncapi.io/groupedServerList.zip");
                    setCopiedCmd(true);
                    setTimeout(() => setCopiedCmd(false), 2000);
                  }}
                  className="h-6 w-6 text-muted-foreground hover:text-primary shrink-0"
                  title="Copy URL"
                >
                  {copiedCmd ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                </Button>
              </div>
            </div>

            <div className="flex items-center space-x-2 pt-1">
              {onDownloadServerLists && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={isUpdatingLists}
                  onClick={async () => {
                    setIsUpdatingLists(true);
                    setUpdateFeedback("idle");
                    try {
                      await onDownloadServerLists();
                      setUpdateFeedback("success");
                      setTimeout(() => setUpdateFeedback("idle"), 2500);
                    } catch (e) {
                      setUpdateFeedback("error");
                      setTimeout(() => setUpdateFeedback("idle"), 3500);
                    } finally {
                      setIsUpdatingLists(false);
                    }
                  }}
                  className={cn(
                    "flex-1 text-sm h-9 space-x-2 font-medium transition-colors",
                    updateFeedback === "success"
                      ? "border-emerald-500/50 text-emerald-400 bg-emerald-950/20"
                      : updateFeedback === "error"
                      ? "border-red-500/50 text-red-400 bg-red-950/20"
                      : "border-primary/40 text-primary hover:bg-primary/10 hover:text-primary"
                  )}
                >
                  {isUpdatingLists ? (
                    <RefreshCw className="w-4 h-4 animate-spin text-primary" />
                  ) : updateFeedback === "success" ? (
                    <Check className="w-4 h-4 text-emerald-400" />
                  ) : updateFeedback === "error" ? (
                    <AlertCircle className="w-4 h-4 text-red-400" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                  <span>
                    {isUpdatingLists
                      ? "Updating..."
                      : updateFeedback === "success"
                      ? "Updated!"
                      : updateFeedback === "error"
                      ? "Update Failed"
                      : "Update Official Lists"}
                  </span>
                </Button>
              )}

              {onOpenServerListModal && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={onOpenServerListModal}
                  className="text-sm h-9 text-slate-300 hover:text-foreground font-medium"
                  title="Configure custom directories or re-download"
                >
                  <FolderOpen className="w-4 h-4 mr-1.5" />
                  <span>Manage</span>
                </Button>
              )}
            </div>
          </div>

          {/* Linux Privilege Status */}
          <div className="p-3.5 rounded-lg bg-[#111724] border border-border/80 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold text-slate-200">
                Passwordless Privileges
              </span>
              {hasPrivileges ? (
                <Badge variant="connected" className="flex items-center space-x-1.5 text-xs px-2.5 py-0.5">
                  <ShieldCheck className="w-3.5 h-3.5" />
                  <span>Configured</span>
                </Badge>
              ) : (
                <Badge variant="accent" className="flex items-center space-x-1.5 text-xs px-2.5 py-0.5 bg-amber-500/20 text-amber-400 border-amber-500/40">
                  <ShieldAlert className="w-3.5 h-3.5" />
                  <span>Password Prompt Active</span>
                </Badge>
              )}
            </div>

            <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
              To connect and disconnect seamlessly without system password popups, run the privilege setup script:
            </p>

            {onRequestGrantPrivileges && (
              <Button
                type="button"
                variant={hasPrivileges ? "outline" : "default"}
                size="sm"
                disabled={isGranting}
                onClick={async () => {
                  setIsGranting(true);
                  setGrantFeedback({ status: "idle", msg: "" });
                  try {
                    await onRequestGrantPrivileges();
                    setGrantFeedback({
                      status: "success",
                      msg: "Elevation dialog completed successfully! Privileges are verified.",
                    });
                  } catch (err: any) {
                    setGrantFeedback({
                      status: "error",
                      msg: err?.message || "Elevation dialog failed or was cancelled.",
                    });
                  } finally {
                    setIsGranting(false);
                  }
                }}
                className={cn(
                  "w-full text-sm h-9 space-x-2 shadow-sm font-medium",
                  !hasPrivileges && "bg-primary hover:bg-primary-hover text-white"
                )}
              >
                <ShieldCheck className="w-4 h-4" />
                <span>
                  {isGranting
                    ? "Opening Elevation Dialog..."
                    : hasPrivileges
                    ? "Re-run / Test Elevation Dialog (kdesu)"
                    : "Grant Permissions Automatically (One-Time)"}
                </span>
              </Button>
            )}

            {grantFeedback.status !== "idle" && (
              <div
                className={cn(
                  "text-xs p-2.5 rounded border leading-relaxed",
                  grantFeedback.status === "success"
                    ? "bg-emerald-950/40 border-emerald-500/40 text-emerald-400"
                    : "bg-red-950/40 border-red-500/40 text-red-400"
                )}
              >
                {grantFeedback.msg}
              </div>
            )}

            <div className="flex items-center justify-between px-3 py-2 rounded bg-[#090C14] border border-border/70 font-mono text-xs text-slate-300">
              <span className="truncate select-all">{setupCmd}</span>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={handleCopyCmd}
                className="h-6 w-6 text-muted-foreground hover:text-primary shrink-0 ml-2"
                title="Copy command"
              >
                {copiedCmd ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              </Button>
            </div>

            {/* Transparency Note: explain what the privilege script modifies */}
            <div className="rounded-lg bg-[#080B12] p-2.5 border border-border/50 text-[11px] text-muted-foreground space-y-1">
              <span className="font-semibold text-slate-300 block">Security &amp; Permissions Details:</span>
              <ul className="list-disc list-inside space-y-0.5 text-slate-400">
                <li>Installs Polkit rule to <code className="text-slate-300 font-mono">/etc/polkit-1/rules.d/50-fastvpn-openvpn.rules</code></li>
                <li>Installs Sudoers rule to <code className="text-slate-300 font-mono">/etc/sudoers.d/fastvpn-openvpn</code></li>
                <li>Restricted strictly to <code className="text-slate-300 font-mono">openvpn</code> &amp; <code className="text-slate-300 font-mono">pkill -x openvpn</code> for your user account.</li>
              </ul>
            </div>
          </div>
        </div>

        <DialogFooter className="p-4 pt-3 border-t border-border/70 bg-[#0A0D15] shrink-0 flex items-center justify-between sm:justify-between w-full">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => api.openDevTools()}
            className="text-sm text-muted-foreground hover:text-foreground flex items-center space-x-2 h-9 px-2.5"
            title="Open Chrome/WebKit DevTools console"
          >
            <Terminal className="w-4 h-4" />
            <span>DevTools (F12)</span>
          </Button>

          <div className="flex items-center space-x-2.5">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              className="text-sm h-9 px-4"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isSaving}
              className="text-sm h-9 px-4 font-medium"
            >
              {isSaving ? "Saving..." : "Save Credentials"}
            </Button>
          </div>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
);
};

