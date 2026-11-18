import React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { ExternalLink, Sparkles, X } from "lucide-react";
import { UpdateInfo } from "@/types/vpn";
import { api } from "@/services/api";

interface UpdateDialogProps {
  isOpen: boolean;
  onClose: () => void;
  updateInfo: UpdateInfo | null;
}

export const UpdateDialog: React.FC<UpdateDialogProps> = ({
  isOpen,
  onClose,
  updateInfo,
}) => {
  if (!updateInfo) return null;

  const handleOpenRelease = () => {
    if (updateInfo.release_url) {
      api.openExternalUrl(updateInfo.release_url);
    }
  };

  const formatDate = (dateStr: string) => {
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg bg-[#0C1018] border-border text-foreground shadow-2xl p-0 overflow-hidden max-h-[88vh] flex flex-col">
        {/* Header */}
        <DialogHeader className="p-5 pb-4 border-b border-border/70 shrink-0 bg-[#090C14]">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2.5">
              <div className="p-2 rounded-lg bg-primary/10 border border-primary/30 text-primary">
                <Sparkles className="w-5 h-5 text-[#FF8A3D]" />
              </div>
              <div>
                <DialogTitle className="text-base sm:text-lg font-bold tracking-tight text-white flex items-center space-x-2">
                  <span>New Version Available</span>
                  <Badge variant="accent" className="bg-[#FF8A3D]/20 text-[#FF8A3D] border-[#FF8A3D]/40 text-xs px-2 py-0.5">
                    v{updateInfo.latest_version}
                  </Badge>
                </DialogTitle>
                <div className="flex items-center space-x-2 text-xs text-muted-foreground mt-0.5">
                  <span>Current: v{updateInfo.current_version}</span>
                  {updateInfo.published_at && (
                    <>
                      <span>•</span>
                      <span>Released: {formatDate(updateInfo.published_at)}</span>
                    </>
                  )}
                </div>
              </div>
            </div>
          </div>
        </DialogHeader>

        {/* Scrollable Changelog / Release Notes */}
        <div className="p-5 overflow-y-auto space-y-3 flex-1 select-text">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
              Release Notes &amp; Fixes
            </span>
          </div>

          <div className="rounded-lg bg-[#111724] border border-border/80 p-4 text-xs sm:text-sm text-slate-300 leading-relaxed font-sans max-h-64 overflow-y-auto whitespace-pre-wrap">
            {updateInfo.release_notes && updateInfo.release_notes.trim().length > 0 ? (
              updateInfo.release_notes
            ) : (
              <span className="text-muted-foreground italic">
                A new version of FastVPN GUI is available for download on GitHub.
              </span>
            )}
          </div>

          <p className="text-xs text-muted-foreground leading-relaxed">
            Download packages (.deb, universal .tar.gz, or Arch recipe) from GitHub Releases or build from source.
          </p>
        </div>

        {/* Footer */}
        <DialogFooter className="p-4 border-t border-border/70 bg-[#090C14] shrink-0 flex items-center justify-between sm:justify-between w-full">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-xs text-muted-foreground hover:text-foreground h-9 px-3"
          >
            <X className="w-4 h-4 mr-1.5" />
            <span>Later</span>
          </Button>

          <Button
            type="button"
            size="sm"
            onClick={handleOpenRelease}
            className="bg-primary hover:bg-primary-hover text-white text-xs sm:text-sm font-semibold h-9 px-4 space-x-2 shadow-sm"
          >
            <span>Download v{updateInfo.latest_version}</span>
            <ExternalLink className="w-4 h-4" />
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
