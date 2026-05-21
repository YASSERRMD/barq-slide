"use client";

import { useRef, useState } from "react";
import { Download, FileDown, Loader2 } from "lucide-react";
import { useDeckStore } from "@/lib/store/deck-store";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

export function DownloadPPTX() {
  const progress    = useDeckStore((s) => s.progress);
  const requestId   = useDeckStore((s) => s.requestId);
  const plan        = useDeckStore((s) => s.plan);
  const slideCount  = useDeckStore((s) => s.slideOrder.length);

  const [isExporting, setIsExporting]       = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const [error, setError]                   = useState<string | null>(null);
  const abortRef                            = useRef<AbortController | null>(null);

  const isReady = progress.status === "completed" && slideCount > 0;

  const handleExport = async () => {
    if (!requestId) return;

    // Cancel any in-flight export before starting a new one.
    abortRef.current?.abort();
    abortRef.current = new AbortController();

    setIsExporting(true);
    setExportProgress(0);
    setError(null);

    try {
      const state = useDeckStore.getState();

      // Build slides array in the order they appear.
      const slides = state.slideOrder
        .map((id) => state.slidesById[id])
        .filter(Boolean);

      const tokens = state.tokens;

      setExportProgress(10);

      // POST raw JSON — backend returns binary PPTX directly.
      // BigInt replacer: protobuf int64 fields come as BigInt in JS.
      const bigintReplacer = (_: string, v: unknown) =>
        typeof v === "bigint" ? Number(v) : v;

      const resp = await fetch(`${API_BASE_URL}/api/export`, {
        method:  "POST",
        headers: { "Content-Type": "application/json" },
        body:    JSON.stringify({ requestId, slides, tokens }, bigintReplacer),
        signal:  abortRef.current.signal,
      });

      if (!resp.ok) {
        const text = await resp.text();
        throw new Error(`Export failed (${resp.status}): ${text}`);
      }

      setExportProgress(60);

      // Stream the binary response into a Blob.
      const blob = await resp.blob();

      setExportProgress(95);

      // Trigger browser download.
      const url = URL.createObjectURL(blob);
      const a   = document.createElement("a");
      a.href    = url;
      a.download = `barq-slides-${requestId}.pptx`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      setExportProgress(100);
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "AbortError") return;
      console.error("Export error:", err);
      setError(err instanceof Error ? err.message : "Export failed");
    } finally {
      setTimeout(() => {
        setIsExporting(false);
        setExportProgress(0);
        setError(null);
      }, 1500);
    }
  };

  if (!isReady && !isExporting) return null;

  return (
    <div className="flex flex-col gap-1" id="download-pptx">
      <div className="flex items-center gap-3">
        <button
          id="download-pptx-btn"
          onClick={handleExport}
          disabled={isExporting}
          className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-primary to-primary/80 text-primary-foreground text-sm font-medium shadow-sm hover:shadow-md hover:from-primary/95 hover:to-primary/75 disabled:opacity-60 transition-all duration-200 active:scale-[0.98]"
        >
          {isExporting ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              {exportProgress > 0 ? `${exportProgress}%` : "Preparing…"}
            </>
          ) : (
            <>
              <Download className="h-4 w-4" />
              Download PPTX
            </>
          )}
        </button>

        {plan && !isExporting && (
          <span className="text-xs text-muted-foreground/60">
            <FileDown className="h-3 w-3 inline mr-1" />
            {slideCount} slides
          </span>
        )}
      </div>

      {error && (
        <p className="text-xs text-red-500 max-w-xs truncate">{error}</p>
      )}
    </div>
  );
}
