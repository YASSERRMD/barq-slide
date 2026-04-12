"use client";

import { useState, useRef, useEffect, type KeyboardEvent } from "react";
import { Send, Sparkles, ChevronDown } from "lucide-react";
import { useGenerateDeck } from "@/lib/store/use-generate-deck";
import { useDeckStore } from "@/lib/store/deck-store";

const TONE_OPTIONS = ["professional", "conversational", "academic", "bold", "minimal"] as const;
const AUDIENCE_OPTIONS = ["executives", "team", "investors", "students", "general"] as const;

export function PromptInput() {
  const [prompt, setPrompt] = useState("");
  const [audience, setAudience] = useState("");
  const [tone, setTone] = useState("");
  const [showOptions, setShowOptions] = useState(false);
  const [slideCount, setSlideCount] = useState(10);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const generate = useGenerateDeck();
  const status = useDeckStore((s) => s.progress.status);
  const isGenerating = status !== "idle" && status !== "completed" && status !== "error";

  // Auto-resize textarea.
  useEffect(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 180)}px`;
  }, [prompt]);

  const handleSubmit = () => {
    if (!prompt.trim() || isGenerating) return;
    generate({
      prompt: prompt.trim(),
      audience,
      tone,
      totalSlides: slideCount,
    });
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto" id="prompt-input-container">
      <div className="relative rounded-2xl border border-border/50 bg-card shadow-lg shadow-black/5 transition-all focus-within:border-primary/30 focus-within:shadow-primary/5 focus-within:shadow-xl">
        {/* Main input area */}
        <div className="flex items-end gap-2 p-3">
          <div className="flex-1 min-w-0">
            <textarea
              ref={textareaRef}
              id="prompt-textarea"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Describe your presentation... (e.g. 'Quarterly revenue analysis for the board')"
              rows={1}
              disabled={isGenerating}
              className="w-full resize-none bg-transparent text-sm leading-relaxed placeholder:text-muted-foreground/50 focus:outline-none disabled:opacity-50 py-1.5"
            />
          </div>

          {/* Options toggle */}
          <button
            id="options-toggle"
            type="button"
            onClick={() => setShowOptions(!showOptions)}
            className="flex-shrink-0 p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
            title="Generation options"
          >
            <ChevronDown
              className={`h-4 w-4 transition-transform duration-200 ${
                showOptions ? "rotate-180" : ""
              }`}
            />
          </button>

          {/* Send button */}
          <button
            id="send-button"
            type="button"
            onClick={handleSubmit}
            disabled={!prompt.trim() || isGenerating}
            className="flex-shrink-0 p-2.5 rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-200 active:scale-95"
            title="Generate deck"
          >
            {isGenerating ? (
              <Sparkles className="h-4 w-4 animate-pulse" />
            ) : (
              <Send className="h-4 w-4" />
            )}
          </button>
        </div>

        {/* Expanded options panel */}
        {showOptions && (
          <div
            className="border-t border-border/30 px-4 py-3 space-y-3 animate-fade-in"
            id="options-panel"
          >
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                  Audience
                </label>
                <select
                  id="audience-select"
                  value={audience}
                  onChange={(e) => setAudience(e.target.value)}
                  className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-primary/30"
                >
                  <option value="">Auto-detect</option>
                  {AUDIENCE_OPTIONS.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt.charAt(0).toUpperCase() + opt.slice(1)}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                  Tone
                </label>
                <select
                  id="tone-select"
                  value={tone}
                  onChange={(e) => setTone(e.target.value)}
                  className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-primary/30"
                >
                  <option value="">Auto-detect</option>
                  {TONE_OPTIONS.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt.charAt(0).toUpperCase() + opt.slice(1)}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                Slides: {slideCount}
              </label>
              <input
                id="slide-count-range"
                type="range"
                min={5}
                max={30}
                value={slideCount}
                onChange={(e) => setSlideCount(Number(e.target.value))}
                className="w-full h-1.5 rounded-full appearance-none bg-muted cursor-pointer accent-primary"
              />
              <div className="flex justify-between text-[10px] text-muted-foreground/60 mt-0.5">
                <span>5</span>
                <span>30</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
