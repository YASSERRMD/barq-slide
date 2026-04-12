"use client";

import { AlertTriangle } from "lucide-react";
import { PromptInput } from "@/components/prompt-input";
import { SlideEditor } from "@/components/slide-editor";
import { DownloadPPTX } from "@/components/download-pptx";
import { GenerationProgress } from "@/components/generation-progress";
import { LlmSettingsModal } from "@/components/llm-settings-modal";
import { useDeckStore } from "@/lib/store/deck-store";

export default function HomePage() {
  const status = useDeckStore((s) => s.progress.status);
  const slideCount = useDeckStore((s) => s.slideOrder.length);
  const llmApiKey = useDeckStore((s) => s.llmApiKey);
  const llmProvider = useDeckStore((s) => s.llmProvider);

  const isIdle = status === "idle";
  const needsApiKey = llmProvider !== "ollama" && !llmApiKey;

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

      {/* API key warning banner */}
      {needsApiKey && (
        <div className="w-full max-w-2xl flex items-center gap-3 px-4 py-3 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-600 text-sm animate-fade-in">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span className="flex-1">No API key configured — slides cannot be generated.</span>
          <LlmSettingsModal />
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
