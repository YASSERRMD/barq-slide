"use client";

import { useDeckStore } from "@/lib/store/deck-store";
import { Loader2, Brain, Palette, Layers, CheckCircle2, XCircle } from "lucide-react";

const PHASE_CONFIG: Record<string, { label: string; icon: React.ReactNode; color: string }> = {
  started:    { label: "Connecting…",          icon: <Loader2 className="h-5 w-5 animate-spin" />, color: "text-blue-500" },
  planning:   { label: "Building slide plan…", icon: <Brain className="h-5 w-5 animate-pulse" />,  color: "text-violet-500" },
  theming:    { label: "Applying design…",     icon: <Palette className="h-5 w-5 animate-pulse" />, color: "text-pink-500" },
  generating: { label: "Generating slides…",   icon: <Layers className="h-5 w-5 animate-pulse" />, color: "text-emerald-500" },
  completed:  { label: "Done!",                icon: <CheckCircle2 className="h-5 w-5" />,          color: "text-emerald-500" },
  error:      { label: "Generation failed",    icon: <XCircle className="h-5 w-5" />,              color: "text-red-500" },
};

export function GenerationProgress() {
  const progress = useDeckStore((s) => s.progress);
  const slideOrder = useDeckStore((s) => s.slideOrder);
  const plan = useDeckStore((s) => s.plan);

  const { status, message, progressPct, error } = progress;
  const phase = PHASE_CONFIG[status] ?? PHASE_CONFIG.started;
  const slidesSoFar = slideOrder.length;
  const totalExpected = plan?.totalSlides ?? 0;

  if (status === "idle") return null;

  return (
    <div className="w-full max-w-2xl mx-auto mt-6 rounded-2xl border border-border/50 bg-card shadow-sm overflow-hidden animate-fade-in">
      {/* Header row */}
      <div className="flex items-center gap-3 px-5 py-4 border-b border-border/30">
        <span className={phase.color}>{phase.icon}</span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-foreground">{phase.label}</p>
          {message && (
            <p className="text-xs text-muted-foreground mt-0.5 truncate">{message}</p>
          )}
        </div>
        {status !== "error" && status !== "completed" && (
          <span className="text-xs tabular-nums font-mono text-muted-foreground bg-muted/50 px-2 py-0.5 rounded-md">
            {progressPct}%
          </span>
        )}
      </div>

      {/* Progress bar */}
      {status !== "error" && (
        <div className="h-1.5 w-full bg-muted/40">
          <div
            className="h-full bg-gradient-to-r from-primary to-primary/70 transition-all duration-500 ease-out"
            style={{ width: `${Math.max(3, progressPct)}%` }}
          />
        </div>
      )}

      {/* Slide count during generation */}
      {slidesSoFar > 0 && (
        <div className="px-5 py-3 flex items-center gap-3 flex-wrap">
          {Array.from({ length: totalExpected || slidesSoFar }, (_, i) => (
            <div
              key={i}
              title={`Slide ${i + 1}`}
              className={`h-2 w-2 rounded-full transition-all duration-300 ${
                i < slidesSoFar
                  ? "bg-primary scale-110"
                  : "bg-muted/60"
              }`}
            />
          ))}
          <span className="text-xs text-muted-foreground ml-1">
            {slidesSoFar}{totalExpected > 0 ? ` / ${totalExpected}` : ""} slide{slidesSoFar !== 1 ? "s" : ""} ready
          </span>
        </div>
      )}

      {/* Pipeline steps */}
      <div className="px-5 py-3 flex gap-6 text-xs text-muted-foreground">
        {(["started", "planning", "theming", "generating"] as const).map((step) => {
          const steps = ["started", "planning", "theming", "generating", "completed"];
          const currentIdx = steps.indexOf(status);
          const stepIdx = steps.indexOf(step);
          const done = currentIdx > stepIdx;
          const active = step === status;
          return (
            <div key={step} className={`flex items-center gap-1.5 transition-colors ${active ? "text-foreground font-medium" : done ? "text-emerald-500" : ""}`}>
              <span className={`h-1.5 w-1.5 rounded-full ${active ? "bg-primary animate-pulse" : done ? "bg-emerald-500" : "bg-muted/50"}`} />
              {step === "started" ? "Intent" : step === "planning" ? "Plan" : step === "theming" ? "Design" : "Slides"}
            </div>
          );
        })}
      </div>

      {/* Error message */}
      {status === "error" && (
        <div className="px-5 pb-4 space-y-1">
          {message && (
            <p className="text-xs text-red-500 font-medium">{message}</p>
          )}
          {error && error !== message && (
            <p className="text-xs text-muted-foreground font-mono break-all">{error}</p>
          )}
        </div>
      )}
    </div>
  );
}
