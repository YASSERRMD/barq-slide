"use client";

import { useEffect, useRef, useState } from "react";
import {
  CheckCircle,
  XCircle,
  Loader2,
  Sparkles,
  Palette,
  LayoutGrid,
  FileCheck,
  Zap,
} from "lucide-react";
import { useDeckStore, type GenerationStatus } from "@/lib/store/deck-store";

interface LogEntry {
  id: number;
  status: GenerationStatus;
  message: string;
  timestamp: Date;
}

const STATUS_CONFIG: Record<
  GenerationStatus,
  { icon: React.ElementType; color: string; label: string }
> = {
  idle: { icon: Zap, color: "text-muted-foreground", label: "Ready" },
  started: { icon: Loader2, color: "text-blue-400", label: "Starting" },
  planning: { icon: Sparkles, color: "text-violet-400", label: "Planning" },
  theming: { icon: Palette, color: "text-amber-400", label: "Theming" },
  generating: {
    icon: LayoutGrid,
    color: "text-cyan-400",
    label: "Generating",
  },
  completed: { icon: CheckCircle, color: "text-emerald-400", label: "Done" },
  error: { icon: XCircle, color: "text-red-400", label: "Error" },
};

export function TerminalProgress() {
  const progress = useDeckStore((s) => s.progress);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const scrollRef = useRef<HTMLDivElement>(null);
  const prevStatusRef = useRef<string>("");
  const idCounterRef = useRef(0);

  // Append log entries when progress changes.
  useEffect(() => {
    const key = `${progress.status}:${progress.message}`;
    if (key === prevStatusRef.current) return;
    prevStatusRef.current = key;

    if (progress.status === "idle" && !progress.message) return;

    setLogs((prev) => [
      ...prev,
      {
        id: ++idCounterRef.current,
        status: progress.status,
        message: progress.message,
        timestamp: new Date(),
      },
    ]);
  }, [progress.status, progress.message]);

  // Auto-scroll to bottom.
  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  }, [logs]);

  const config = STATUS_CONFIG[progress.status];
  const StatusIcon = config.icon;
  const isActive =
    progress.status !== "idle" && progress.status !== "completed" && progress.status !== "error";

  if (progress.status === "idle" && logs.length === 0) {
    return null;
  }

  return (
    <div
      className="w-full max-w-2xl mx-auto mt-4"
      id="terminal-progress"
    >
      <div className="rounded-xl border border-border/50 bg-card overflow-hidden shadow-sm">
        {/* Header bar */}
        <div className="flex items-center gap-2 px-4 py-2.5 border-b border-border/30 bg-muted/30">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full bg-red-400/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-amber-400/60" />
            <div className="w-2.5 h-2.5 rounded-full bg-emerald-400/60" />
          </div>
          <span className="text-xs font-medium text-muted-foreground ml-2">
            Generation Pipeline
          </span>
          <div className="ml-auto flex items-center gap-1.5">
            <StatusIcon
              className={`h-3.5 w-3.5 ${config.color} ${
                isActive ? "animate-spin" : ""
              }`}
            />
            <span className={`text-xs font-medium ${config.color}`}>
              {config.label}
            </span>
          </div>
        </div>

        {/* Progress bar */}
        {isActive && (
          <div className="h-0.5 bg-muted">
            <div
              className="h-full bg-gradient-to-r from-primary/80 to-primary transition-all duration-500 ease-out"
              style={{ width: `${progress.progressPct}%` }}
            />
          </div>
        )}

        {/* Log entries */}
        <div
          ref={scrollRef}
          className="max-h-48 overflow-y-auto px-4 py-2 space-y-1 font-mono text-xs"
          id="terminal-log"
        >
          {logs.map((entry) => {
            const c = STATUS_CONFIG[entry.status];
            const Icon = c.icon;
            return (
              <div
                key={entry.id}
                className="flex items-start gap-2 py-0.5 animate-fade-in"
              >
                <span className="text-muted-foreground/40 flex-shrink-0 w-16 text-right tabular-nums">
                  {entry.timestamp.toLocaleTimeString("en-US", {
                    hour12: false,
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                  })}
                </span>
                <Icon className={`h-3.5 w-3.5 mt-0.5 flex-shrink-0 ${c.color}`} />
                <span
                  className={`${
                    entry.status === "error"
                      ? "text-red-400"
                      : "text-foreground/80"
                  }`}
                >
                  {entry.message}
                </span>
              </div>
            );
          })}
        </div>

        {/* Error detail */}
        {progress.status === "error" && progress.error && (
          <div className="px-4 py-2 border-t border-red-500/10 bg-red-500/5">
            <p className="text-xs text-red-400 font-mono">
              Error code: {progress.error}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
