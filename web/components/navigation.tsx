"use client";

import { Zap } from "lucide-react";
import { LlmSettingsModal } from "./llm-settings-modal";

export function Navigation() {
  return (
    <nav className="fixed top-0 inset-x-0 h-14 border-b border-border/40 bg-background/80 backdrop-blur-md z-40 flex items-center justify-between px-6">
      <div className="flex items-center gap-2">
        <div className="flex items-center justify-center w-7 h-7 rounded-lg bg-primary/10 text-primary">
          <Zap className="h-4 w-4" />
        </div>
        <span className="font-semibold text-sm tracking-tight text-foreground">
          barq-slides
        </span>
      </div>
      
      <div className="flex items-center">
        <LlmSettingsModal />
      </div>
    </nav>
  );
}
