"use client";

import { PromptInput } from "@/components/prompt-input";
import { SlideEditor } from "@/components/slide-editor";
import { DownloadPPTX } from "@/components/download-pptx";
import { GenerationProgress } from "@/components/generation-progress";
import { useDeckStore } from "@/lib/store/deck-store";

export default function HomePage() {
  const status = useDeckStore((s) => s.progress.status);
  const slideCount = useDeckStore((s) => s.slideOrder.length);

  const isIdle = status === "idle";

  return (
    <main className="flex min-h-screen flex-col items-center p-6 md:p-12 gap-6">
      {/* Hero — only shown on idle */}
      {isIdle && (
        <div className="text-center space-y-3 mt-16 mb-2" id="hero-section">
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight bg-gradient-to-br from-foreground to-foreground/60 bg-clip-text text-transparent">
            Create beautiful presentations in seconds
          </h1>
          <p className="text-muted-foreground text-sm">
            Describe your topic and let AI build the deck.
          </p>
        </div>
      )}

      {/* Prompt Input */}
      <PromptInput />

      {/* Progress panel — visible during and after generation */}
      <GenerationProgress />

      {/* Slide count badge + Download — once slides arrive */}
      {slideCount > 0 && (
        <div className="flex items-center gap-4 animate-fade-in">
          <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 text-xs font-medium">
            {slideCount} slide{slideCount !== 1 ? "s" : ""} generated
          </div>
          {status === "completed" && <DownloadPPTX />}
        </div>
      )}

      {/* Slide Editor — streams in as slides are ready */}
      {slideCount > 0 && <SlideEditor />}
    </main>
  );
}
