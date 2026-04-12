"use client";

import { PromptInput } from "@/components/prompt-input";
import { TerminalProgress } from "@/components/terminal-progress";
import { SlideEditor } from "@/components/slide-editor";
import { useDeckStore } from "@/lib/store/deck-store";
import { Zap } from "lucide-react";

export default function HomePage() {
  const status = useDeckStore((s) => s.progress.status);
  const slideCount = useDeckStore((s) => s.slideOrder.length);
  const isActive = status !== "idle";

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-6 md:p-24 gap-8">
      {/* Hero */}
      <div className="text-center space-y-3 mb-4" id="hero-section">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/5 border border-primary/10 text-primary text-xs font-medium mb-2">
          <Zap className="h-3 w-3" />
          AI Presentation Engine
        </div>
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight bg-gradient-to-br from-foreground to-foreground/60 bg-clip-text text-transparent">
          barq-slides
        </h1>
        <p className="text-base text-muted-foreground max-w-lg mx-auto leading-relaxed">
          Describe your presentation and watch it come to life.
          Real-time HTML preview with 100% editable PPTX export.
        </p>
      </div>

      {/* Prompt Input */}
      <PromptInput />

      {/* Terminal Progress */}
      <TerminalProgress />

      {/* Slide count badge */}
      {isActive && slideCount > 0 && (
        <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 text-xs font-medium animate-fade-in">
          {slideCount} slide{slideCount !== 1 ? "s" : ""} generated
        </div>
      )}

      {/* Slide Editor */}
      <SlideEditor />
    </main>
  );
}
