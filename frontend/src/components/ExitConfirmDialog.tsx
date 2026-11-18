import React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle, PowerOff, ShieldCheck } from "lucide-react";

interface ExitConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onDisconnectAndQuit: () => void;
  onKeepAndQuit: () => void;
  publicIp: string;
}

export const ExitConfirmDialog: React.FC<ExitConfirmDialogProps> = ({
  isOpen,
  onClose,
  onDisconnectAndQuit,
  onKeepAndQuit,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-md bg-[#0D111A] border-border text-foreground">
        <DialogHeader>
          <div className="flex items-center space-x-2 text-amber-400 mb-1">
            <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0" />
            <DialogTitle className="text-base font-bold text-foreground">
              Active VPN Connection
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-muted-foreground leading-relaxed">
            FastVPN is currently connected. What would you like to do before closing the application?
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <div className="p-3 rounded-lg bg-[#111724] border border-border/80 flex items-start space-x-3">
            <PowerOff className="w-4 h-4 text-destructive shrink-0 mt-0.5" />
            <div>
              <h4 className="text-xs font-semibold text-foreground">
                Disconnect & Quit
              </h4>
              <p className="text-[11px] text-muted-foreground">
                Gracefully disconnect OpenVPN, restore original networking, and exit.
              </p>
            </div>
          </div>

          <div className="p-3 rounded-lg bg-[#111724] border border-border/80 flex items-start space-x-3">
            <ShieldCheck className="w-4 h-4 text-primary shrink-0 mt-0.5" />
            <div>
              <h4 className="text-xs font-semibold text-foreground">
                Keep VPN & Quit
              </h4>
              <p className="text-[11px] text-muted-foreground">
                Keep VPN running in the background. The active connection and session will be restored next time you open the app.
              </p>
            </div>
          </div>
        </div>

        <DialogFooter className="flex-col sm:flex-row gap-2 pt-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClose}
            className="text-xs"
          >
            Cancel
          </Button>

          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={onKeepAndQuit}
            className="text-xs bg-[#172030] hover:bg-[#1C273C] text-foreground border border-border"
          >
            Keep VPN & Quit
          </Button>

          <Button
            type="button"
            variant="destructive"
            size="sm"
            onClick={onDisconnectAndQuit}
            className="text-xs"
          >
            Disconnect & Quit
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

