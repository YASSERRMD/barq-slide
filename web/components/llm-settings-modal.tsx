"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, CheckCircle2, Eye, EyeOff, Loader2, Settings2, X } from "lucide-react";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

interface LLMConfig {
  provider: string;
  apiKey: string;
  model: string;
  baseUrl: string;
}

export function LlmSettingsModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [showKey, setShowKey] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [configured, setConfigured] = useState(false);

  const [config, setConfig] = useState<LLMConfig>({
    provider: "anthropic",
    apiKey: "",
    model: "",
    baseUrl: "",
  });

  // Check on mount whether backend already has a config saved
  useEffect(() => {
    fetch(`${API_BASE_URL}/api/llm-config`)
      .then((r) => r.json())
      .then((data) => {
        setConfigured(
          data.provider === "ollama" || (data.apiKey && data.apiKey !== "***")
        );
      })
      .catch(() => {});
  }, []);

  // Load full config from backend when modal opens
  useEffect(() => {
    if (!isOpen) return;
    fetch(`${API_BASE_URL}/api/llm-config`)
      .then((r) => r.json())
      .then((data) => {
        setConfig({
          provider: data.provider || "anthropic",
          // Backend masks the key as "***" — keep field blank so user can re-enter if needed
          apiKey: data.apiKey === "***" ? "" : (data.apiKey || ""),
          model: data.model || "",
          baseUrl: data.baseUrl || "",
        });
      })
      .catch(() => {});
  }, [isOpen]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await fetch(`${API_BASE_URL}/api/llm-config`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config),
      });
      setConfigured(true);
      setSaved(true);
      setTimeout(() => {
        setSaved(false);
        setIsOpen(false);
      }, 800);
    } catch {
      // ignore
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="relative flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground rounded-md hover:bg-muted/50 transition-colors"
      >
        <Settings2 className="h-4 w-4" />
        Settings
        {!configured && (
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

            {!configured && (
              <div className="flex items-start gap-2.5 mb-4 px-3 py-2.5 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-600 text-xs">
                <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
                <span>No API key saved on server — slides cannot be generated until you save a valid key.</span>
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Provider</label>
                <select
                  value={config.provider}
                  onChange={(e) => setConfig({ ...config, provider: e.target.value })}
                  className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                >
                  <option value="anthropic">Anthropic</option>
                  <option value="openai">OpenAI / OpenAI-compatible</option>
                  <option value="gemini">Gemini</option>
                  <option value="xai">xAI</option>
                  <option value="minimax">Minimax</option>
                  <option value="ollama">Ollama (local)</option>
                </select>
              </div>

              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1.5 block">API Key</label>
                <div className="relative">
                  <input
                    type={showKey ? "text" : "password"}
                    value={config.apiKey}
                    onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                    placeholder={config.provider === "ollama" ? "Not required for Ollama" : "Enter API key"}
                    disabled={config.provider === "ollama"}
                    className="w-full text-sm rounded-lg border border-border/50 bg-background pl-3 pr-9 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30 disabled:opacity-50"
                  />
                  {config.provider !== "ollama" && (
                    <button
                      type="button"
                      onClick={() => setShowKey(!showKey)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Base URL</label>
                  <input
                    type="text"
                    value={config.baseUrl}
                    onChange={(e) => setConfig({ ...config, baseUrl: e.target.value })}
                    placeholder="e.g. https://open.bigmodel.cn/api/paas/v4"
                    className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Model</label>
                  <input
                    type="text"
                    value={config.model}
                    onChange={(e) => setConfig({ ...config, model: e.target.value })}
                    placeholder="e.g. glm-4"
                    className="w-full text-sm rounded-lg border border-border/50 bg-background px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary/30"
                  />
                </div>
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button
                onClick={handleSave}
                disabled={saving}
                className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-60"
              >
                {saving && <Loader2 className="h-4 w-4 animate-spin" />}
                {saved && <CheckCircle2 className="h-4 w-4" />}
                {saved ? "Saved!" : saving ? "Saving…" : "Save to Server"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
