"use client";

import { useState } from "react";
import {
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  LayoutGrid,
} from "lucide-react";
import { useDeckStore } from "@/lib/store/deck-store";
import { SlideCanvas } from "./slide-canvas";

/**
 * SlideEditor provides a thumbnail sidebar listing all slides on the left,
 * a main canvas viewport for the selected slide, and a regeneration panel.
 */
export function SlideEditor() {
  const slideOrder = useDeckStore((s) => s.slideOrder);
  const slidesById = useDeckStore((s) => s.slidesById);
  const tokens = useDeckStore((s) => s.tokens);
  const selectedIdx = useDeckStore((s) => s.selectedSlideIndex);
  const setSelectedIdx = useDeckStore((s) => s.setSelectedSlideIndex);
  const progress = useDeckStore((s) => s.progress);

  const [regenerateInstruction, setRegenerateInstruction] = useState("");
  const [isRegenerating, setIsRegenerating] = useState(false);

  if (slideOrder.length === 0) return null;

  const selectedSlide = slidesById[slideOrder[selectedIdx]];
  const totalSlides = slideOrder.length;

  const handleRegenerate = async () => {
    if (!selectedSlide || !regenerateInstruction.trim()) return;
    setIsRegenerating(true);

    const baseUrl =
      process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

    try {
      const res = await fetch(
        `${baseUrl}/barq.v1.HeliosService/RegenerateSlide`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Connect-Protocol-Version": "1",
            "X-Llm-Provider": useDeckStore.getState().llmProvider || "anthropic",
            "X-Llm-Api-Key": useDeckStore.getState().llmApiKey || "",
            "X-Llm-Model": useDeckStore.getState().llmModel || "",
            "X-Llm-Base-Url": useDeckStore.getState().llmBaseUrl || "",
          },
          body: JSON.stringify({
            requestId: selectedSlide.requestId,
            slideId: selectedSlide.id,
            instruction: regenerateInstruction.trim(),
          }),
        }
      );

      if (!res.ok) {
        console.error("Regeneration failed:", res.status);
      }
      // Future: parse streaming response and upsert slide.
      setRegenerateInstruction("");
    } catch (err: unknown) {
      console.error("Regeneration error:", err instanceof Error ? err.message : err);
    } finally {
      setIsRegenerating(false);
    }
  };

  const goPrev = () => setSelectedIdx(Math.max(0, selectedIdx - 1));
  const goNext = () => setSelectedIdx(Math.min(totalSlides - 1, selectedIdx + 1));

  return (
    <div
      className="w-full max-w-6xl mx-auto mt-6 animate-fade-in"
      id="slide-editor"
    >
      <div className="flex gap-4 h-[600px]">
        {/* Thumbnail sidebar */}
        <div className="w-48 flex-shrink-0 rounded-xl border border-border/50 bg-card overflow-hidden">
          <div className="px-3 py-2.5 border-b border-border/30 bg-muted/30 flex items-center gap-2">
            <LayoutGrid className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs font-medium text-muted-foreground">
              Slides ({totalSlides})
            </span>
          </div>
          <div
            className="overflow-y-auto p-2 space-y-2"
            style={{ maxHeight: "calc(100% - 42px)" }}
            id="slide-thumbnails"
          >
            {slideOrder.map((id, idx) => {
              const s = slidesById[id];
              if (!s) return null;
              const isSelected = idx === selectedIdx;
              return (
                <button
                  key={id}
                  id={`thumbnail-${idx}`}
                  onClick={() => setSelectedIdx(idx)}
                  className={`w-full rounded-lg border-2 transition-all duration-150 overflow-hidden ${
                    isSelected
                      ? "border-primary shadow-sm shadow-primary/10"
                      : "border-transparent hover:border-border/60"
                  }`}
                >
                  {/* Clip container: 960*0.175 × 540*0.175 */}
                  <div style={{ width: "168px", height: "95px", overflow: "hidden", position: "relative" }}>
                    <SlideCanvas slide={s} tokens={tokens} scale={0.175} />
                    <div className="absolute top-1 left-1 px-1.5 py-0.5 rounded bg-black/50 text-white text-[9px] font-mono">
                      {idx + 1}
                    </div>
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Main canvas viewport */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Toolbar */}
          <div className="flex items-center justify-between px-4 py-2 rounded-t-xl border border-b-0 border-border/50 bg-card">
            <div className="flex items-center gap-2">
              <button
                onClick={goPrev}
                disabled={selectedIdx === 0}
                className="p-1.5 rounded-lg hover:bg-muted/50 disabled:opacity-30 transition-colors"
                id="prev-slide-btn"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <span className="text-xs font-medium text-muted-foreground tabular-nums min-w-[60px] text-center">
                {selectedIdx + 1} / {totalSlides}
              </span>
              <button
                onClick={goNext}
                disabled={selectedIdx >= totalSlides - 1}
                className="p-1.5 rounded-lg hover:bg-muted/50 disabled:opacity-30 transition-colors"
                id="next-slide-btn"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>

            {selectedSlide && (
              <span className="text-xs text-muted-foreground/60 truncate max-w-xs">
                {selectedSlide.blocks?.[0]?.text?.slice(0, 50) ?? "Untitled"}
              </span>
            )}
          </div>

          {/* Canvas — clip container: 960*0.74 × 540*0.74 */}
          <div className="flex-1 flex items-center justify-center rounded-b-xl border border-border/50 bg-muted/20 overflow-hidden p-4">
            {selectedSlide && (
              <div style={{ width: "710px", height: "400px", overflow: "hidden", flexShrink: 0, borderRadius: "8px", boxShadow: "0 4px 24px rgba(0,0,0,0.12)" }}>
                <SlideCanvas
                  slide={selectedSlide}
                  tokens={tokens}
                  scale={0.74}
                  isEditable={true}
                />
              </div>
            )}
          </div>

          {/* Regenerate panel */}
          <div className="mt-3 flex gap-2" id="regenerate-panel">
            <input
              type="text"
              id="regenerate-input"
              value={regenerateInstruction}
              onChange={(e) => setRegenerateInstruction(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleRegenerate()}
              placeholder="Regenerate this slide: e.g. 'make it more visual'"
              className="flex-1 text-sm px-3 py-2 rounded-lg border border-border/50 bg-card focus:outline-none focus:ring-1 focus:ring-primary/30 placeholder:text-muted-foreground/40"
              disabled={isRegenerating}
            />
            <button
              id="regenerate-btn"
              onClick={handleRegenerate}
              disabled={!regenerateInstruction.trim() || isRegenerating}
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-secondary/10 border border-secondary/20 text-secondary-foreground text-sm font-medium hover:bg-secondary/20 disabled:opacity-40 transition-colors"
            >
              <RefreshCw
                className={`h-3.5 w-3.5 ${isRegenerating ? "animate-spin" : ""}`}
              />
              Regenerate
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
