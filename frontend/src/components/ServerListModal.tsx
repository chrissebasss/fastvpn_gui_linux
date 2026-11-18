import React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ServerListInfo } from "@/types/vpn";
import { Download, FolderOpen, RefreshCw, Globe, ChevronDown, ChevronUp, AlertCircle, CheckCircle2, Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface ServerListModalProps {
  isOpen: boolean;
  onClose: () => void;
  isInitialOnboarding?: boolean;
  info: ServerListInfo | null;
  onDownload: (url?: string) => Promise<void>;
  onSaveCustomPaths: (udpPath: string, tcpPath: string) => Promise<void>;
}

export const ServerListModal: React.FC<ServerListModalProps> = ({
  isOpen,
  onClose,
  isInitialOnboarding = false,
  info,
  onDownload,
  onSaveCustomPaths,
}) => {
  const [downloadUrl, setDownloadUrl] = React.useState("https://vpn.ncapi.io/groupedServerList.zip");
  const [copiedUrl, setCopiedUrl] = React.useState(false);
  const [isDownloading, setIsDownloading] = React.useState(false);
  const [showCustomAccordion, setShowCustomAccordion] = React.useState(false);
  const [udpPath, setUdpPath] = React.useState(info?.custom_udp_path || "");
  const [tcpPath, setTcpPath] = React.useState(info?.custom_tcp_path || "");
  const [isSavingCustom, setIsSavingCustom] = React.useState(false);
  const [feedback, setFeedback] = React.useState<{ type: "success" | "error"; msg: string } | null>(null);

  React.useEffect(() => {
    if (info) {
      setUdpPath(info.custom_udp_path || "");
      setTcpPath(info.custom_tcp_path || "");
    }
  }, [info]);

  const handleDownload = async () => {
    setIsDownloading(true);
    setFeedback(null);
    try {
      await onDownload(downloadUrl.trim());
      setFeedback({
        type: "success",
        msg: "Server lists successfully downloaded and loaded!",
      });
      if (isInitialOnboarding) {
        setTimeout(() => onClose(), 1200);
      }
    } catch (err: any) {
      setFeedback({
        type: "error",
        msg: err?.message || "Failed to download server lists. Please check your internet connection.",
      });
    } finally {
      setIsDownloading(false);
    }
  };

  const handleCustomSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!udpPath.trim() && !tcpPath.trim()) {
      setFeedback({
        type: "error",
        msg: "Please specify at least one directory path containing .ovpn files.",
      });
      return;
    }
    setIsSavingCustom(true);
    setFeedback(null);
    try {
      await onSaveCustomPaths(udpPath.trim(), tcpPath.trim());
      setFeedback({
        type: "success",
        msg: "Custom server lists configured successfully!",
      });
      setTimeout(() => onClose(), 1000);
    } catch (err: any) {
      setFeedback({
        type: "error",
        msg: err?.message || "Failed to load profiles from specified paths.",
      });
    } finally {
      setIsSavingCustom(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => (!open && !isInitialOnboarding ? onClose() : null)}>
      <DialogContent
        className={cn(
          "sm:max-w-lg bg-[#0D111A] border-border text-foreground shadow-2xl max-h-[88vh] flex flex-col p-0 overflow-hidden",
          isInitialOnboarding && "[&>button]:hidden"
        )}
      >
        <DialogHeader className="text-left space-y-1.5 p-5 pb-3 border-b border-border/70 shrink-0">
          <div className="flex items-center space-x-2.5">
            <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-primary/20 border border-primary/50 text-primary">
              <Globe className="w-5 h-5" />
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {isInitialOnboarding ? "Setup Server Profiles" : "OpenVPN Server Profiles"}
            </DialogTitle>
          </div>
          <DialogDescription className="text-sm text-muted-foreground leading-relaxed mt-1">
            {isInitialOnboarding
              ? "FastVPN requires OpenVPN server profiles (.ovpn) to establish secure connections. Download the latest official server bundle (UDP & TCP) or point to local folders on your system."
              : "Update your official server lists or configure custom folders on your disk."}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {/* Option 1: 1-Click Official Download */}
          <div className="p-4 rounded-xl bg-[#121824] border border-border/80 space-y-3.5">
            <div className="flex items-start justify-between">
              <div>
                <h4 className="text-sm font-semibold text-slate-200 flex items-center space-x-2">
                  <Download className="w-4 h-4 text-primary" />
                  <span>Download Official Server Bundle</span>
                </h4>
                <p className="text-xs sm:text-sm text-muted-foreground mt-1">
                  Downloads ~150 worldwide locations with both UDP and TCP endpoints.
                </p>
              </div>
            </div>

            {/* Download URL display & copy */}
            <div className="space-y-1.5 pt-0.5">
              <div className="flex items-center justify-between text-xs sm:text-sm">
                <span className="text-slate-300 font-semibold">Download URL:</span>
                <button
                  type="button"
                  onClick={() => {
                    navigator.clipboard.writeText(downloadUrl);
                    setCopiedUrl(true);
                    setTimeout(() => setCopiedUrl(false), 2000);
                  }}
                  className="flex items-center space-x-1.5 text-muted-foreground hover:text-primary transition-colors cursor-pointer text-xs"
                  title="Copy URL to clipboard"
                >
                  {copiedUrl ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedUrl ? "Copied" : "Copy URL"}</span>
                </button>
              </div>
              <Input
                type="text"
                value={downloadUrl}
                onChange={(e) => setDownloadUrl(e.target.value)}
                className="bg-[#090C14] border-border text-xs sm:text-sm font-mono h-9 select-all text-slate-300 px-3"
                placeholder="https://vpn.ncapi.io/groupedServerList.zip"
              />
            </div>

            <Button
              type="button"
              disabled={isDownloading}
              onClick={handleDownload}
              className="w-full h-10 bg-primary hover:bg-primary-hover text-white text-sm font-semibold shadow-md flex items-center justify-center space-x-2"
            >
              {isDownloading ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  <span>Downloading & Extracting Server Lists...</span>
                </>
              ) : (
                <>
                  <Download className="w-4 h-4" />
                  <span>Download Server Lists (1-Click)</span>
                </>
              )}
            </Button>
          </div>

          {/* Option 2: Custom Local Folders Accordion */}
          <div className="border border-border/70 rounded-xl overflow-hidden bg-[#0A0E17]">
            <button
              type="button"
              onClick={() => setShowCustomAccordion(!showCustomAccordion)}
              className="w-full flex items-center justify-between p-3.5 text-left hover:bg-[#121824] transition-colors"
            >
              <div className="flex items-center space-x-2 text-sm font-medium text-slate-200">
                <FolderOpen className="w-4 h-4 text-muted-foreground" />
                <span>Or select local server directories on this machine</span>
              </div>
              {showCustomAccordion ? (
                <ChevronUp className="w-4 h-4 text-muted-foreground" />
              ) : (
                <ChevronDown className="w-4 h-4 text-muted-foreground" />
              )}
            </button>

            {showCustomAccordion && (
              <form onSubmit={handleCustomSubmit} className="p-3.5 pt-1 border-t border-border/60 space-y-3.5">
                <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
                  If you have downloaded <code className="text-slate-300 bg-slate-900 px-1 py-0.5 rounded">groupedServerList.zip</code> with your browser, extract it and enter the folder paths below:
                </p>
                <div className="space-y-1.5">
                  <label className="text-xs sm:text-sm font-semibold text-slate-200">
                    UDP Profiles Folder Path
                  </label>
                  <Input
                    type="text"
                    value={udpPath}
                    onChange={(e) => setUdpPath(e.target.value)}
                    placeholder="/path/to/ovpn/udp"
                    className="bg-[#121824] border-border text-xs sm:text-sm h-9 select-all px-3"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs sm:text-sm font-semibold text-slate-200">
                    TCP Profiles Folder Path
                  </label>
                  <Input
                    type="text"
                    value={tcpPath}
                    onChange={(e) => setTcpPath(e.target.value)}
                    placeholder="/path/to/ovpn/tcp"
                    className="bg-[#121824] border-border text-xs sm:text-sm h-9 select-all px-3"
                  />
                </div>

                <Button
                  type="submit"
                  size="sm"
                  variant="outline"
                  disabled={isSavingCustom}
                  className="w-full text-sm h-9 space-x-2 font-medium"
                >
                  <FolderOpen className="w-4 h-4" />
                  <span>{isSavingCustom ? "Applying..." : "Use Custom Folders"}</span>
                </Button>
              </form>
            )}
          </div>

          {/* Feedback messages */}
          {feedback && (
            <div
              className={cn(
                "p-3 rounded-lg border text-xs flex items-start space-x-2.5 select-text",
                feedback.type === "success"
                  ? "bg-emerald-950/40 border-emerald-500/40 text-emerald-400"
                  : "bg-red-950/40 border-red-500/40 text-red-400"
              )}
            >
              {feedback.type === "success" ? (
                <CheckCircle2 className="w-4 h-4 shrink-0 mt-0.5" />
              ) : (
                <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
              )}
              <div className="flex-1 space-y-1 break-all">
                <div className="font-semibold">{feedback.type === "success" ? "Success" : "Connection Error"}</div>
                <div className="text-[11px] opacity-90 leading-relaxed font-mono select-all">{feedback.msg}</div>
              </div>
            </div>
          )}
        </div>

        {!isInitialOnboarding && (
          <div className="p-4 pt-3 border-t border-border/70 bg-[#090C14] shrink-0 flex justify-end">
            <Button variant="ghost" size="sm" onClick={onClose} className="text-sm">
              Close
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
};

