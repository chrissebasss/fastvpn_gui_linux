import React from "react";
import { Button } from "@/components/ui/button";
import { Copy, Trash2, X, Terminal as TerminalIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface LogDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  logs: string[];
  onClearLogs: () => void;
}

export const LogDrawer: React.FC<LogDrawerProps> = ({
  isOpen,
  onClose,
  logs,
  onClearLogs,
}) => {
  const terminalEndRef = React.useRef<HTMLDivElement>(null);
  const [copied, setCopied] = React.useState(false);

  React.useEffect(() => {
    if (isOpen) {
      terminalEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs, isOpen]);

  const handleCopy = () => {
    navigator.clipboard.writeText(logs.join("\n"));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-x-0 bottom-0 z-40 h-[320px] bg-[#080B11] border-t border-border shadow-2xl flex flex-col animate-in slide-in-from-bottom-5 duration-200">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-border/80 bg-[#0C1018]">
        <div className="flex items-center space-x-2">
          <TerminalIcon className="w-4 h-4 text-primary" />
          <span className="text-xs font-semibold text-foreground">OpenVPN Logs</span>
          <span className="text-[10px] text-muted-foreground font-mono">
            ({logs.length} lines)
          </span>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopy}
            className="h-6 px-2 text-xs text-muted-foreground hover:text-foreground space-x-1"
          >
            <Copy className="w-3 h-3" />
            <span>{copied ? "Copied!" : "Copy"}</span>
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={onClearLogs}
            className="h-6 px-2 text-xs text-muted-foreground hover:text-destructive space-x-1"
          >
            <Trash2 className="w-3 h-3" />
            <span>Clear</span>
          </Button>

          <Button
            variant="ghost"
            size="icon"
            onClick={onClose}
            className="h-6 w-6 text-muted-foreground hover:text-foreground"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Terminal View */}
      <div className="flex-1 overflow-y-auto p-4 font-mono text-[11px] leading-relaxed text-slate-300 space-y-0.5 select-text">
        {logs.length === 0 ? (
          <div className="text-muted-foreground italic">No logs recorded yet.</div>
        ) : (
          logs.map((line, idx) => {
            const isError = line.includes("ERROR") || line.includes("AUTH_FAILED") || line.includes("[STDERR]");
            const isSuccess = line.includes("Initialization Sequence Completed") || line.includes("VPN Connected");
            const isInfo = line.startsWith("==>");

            return (
              <div
                key={idx}
                className={cn(
                  "break-all",
                  isError && "text-red-400 font-semibold",
                  isSuccess && "text-emerald-400 font-semibold",
                  isInfo && "text-primary font-medium"
                )}
              >
                {line}
              </div>
            );
          })
        )}
        <div ref={terminalEndRef} />
      </div>
    </div>
  );
};

