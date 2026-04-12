"use client";

import { useState } from "react";
import { useDeckStore } from "@/lib/store/deck-store";
import { AlertTriangle, Eye, EyeOff, Settings2, X } from "lucide-react";

export function LlmSettingsModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [showKey, setShowKey] = useState(false);

  const llmProvider = useDeckStore((s) => s.llmProvider);
  const llmApiKey = useDeckStore((s) => s.llmApiKey);
  const llmModel = useDeckStore((s) => s.llmModel);
  const llmBaseUrl = useDeckStore((s) => s.llmBaseUrl);
  const setLlmConfig = useDeckStore((s) => s.setLlmConfig);

  const needsApiKey = llmProvider !== "ollama" && !llmApiKey;

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="relative flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground rounded-md hover:bg-muted/50 transition-colors"
      >
        <Settings2 className="h-4 w-4" />
        Settings
        {needsApiKey && (
          <span className="absolute -top-1 -right-1 flex h-2.5 w-2.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-amber-500" />
          </span>
        )}
      </button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4 animate-fade-in">
          <div className="bg-card w-full max-w-md rounded-xl border border-border/50 shadow-xl overflow-hidden p-6 relative">
            <button
              onClick={() => setIsOpen(false)}
              className="absolute top-4 right-4 p-1.5 text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-md transition-colors"
            >
              <X className="h-4 w-4" />
            </button>

            <h2 className="text-lg font-semibold text-foreground mb-4">LLM Configuration</h2>

            {needsApiKey && (
              <div className="flex items-start gap-2.5 mb-4 px-3 py-2.5 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-600 text-xs">
                <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
                <span>No API key set — slides cannot be generated until you enter a valid key below.</span>
              </div>
            )}
            
            <div className="space-y-4">
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                  Provider
                </label>
                <select
                  value={llmProvider}
                  onChange={(e) => setLlmConfig(e.target.value, llmApiKey, llmModel, llmBaseUrl)}
                  className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                >
                  <option value="anthropic">Anthropic</option>
                  <option value="openai">OpenAI</option>
                  <option value="gemini">Gemini</option>
                  <option value="xai">xAI</option>
                  <option value="minimax">Minimax</option>
                  <option value="ollama">Ollama</option>
                </select>
              </div>

              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                  API Key
                </label>
                <div className="relative">
                  <input
                    type={showKey ? "text" : "password"}
                    value={llmApiKey}
                    onChange={(e) => setLlmConfig(llmProvider, e.target.value, llmModel, llmBaseUrl)}
                    placeholder="Enter API Key (saved locally)"
                    className="w-full text-sm rounded-lg border border-border/50 bg-background pl-3 pr-9 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                  />
                  <button
                    type="button"
                    onClick={() => setShowKey(!showKey)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    title={showKey ? "Hide key" : "Show key"}
                  >
                    {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 pt-2">
                 <div>
                    <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                      Custom Base URL
                    </label>
                    <input
                      type="text"
                      value={llmBaseUrl}
                      onChange={(e) => setLlmConfig(llmProvider, llmApiKey, llmModel, e.target.value)}
                      placeholder="e.g. http://localhost:11434/v1"
                      className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                    />
                 </div>
                 <div>
                    <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                      Model Override
                    </label>
                    <input
                      type="text"
                      value={llmModel}
                      onChange={(e) => setLlmConfig(llmProvider, llmApiKey, e.target.value, llmBaseUrl)}
                      placeholder="e.g. llama3"
                      className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                    />
                 </div>
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button
                onClick={() => setIsOpen(false)}
                className="px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
