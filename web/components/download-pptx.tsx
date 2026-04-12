"use client";

import { useState } from "react";
import { Download, FileDown, Loader2 } from "lucide-react";
import { useDeckStore } from "@/lib/store/deck-store";

/**
 * DownloadPPTX button triggers the ExportDeck server-streaming RPC,
 * accumulates chunked binary data, and initiates a file download
 * when all chunks are received.
 */
export function DownloadPPTX() {
  const progress = useDeckStore((s) => s.progress);
  const requestId = useDeckStore((s) => s.requestId);
  const plan = useDeckStore((s) => s.plan);
  const slideCount = useDeckStore((s) => s.slideOrder.length);

  const [isExporting, setIsExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);

  const isReady = progress.status === "completed" && slideCount > 0;

  const handleExport = async () => {
    if (!requestId || isExporting) return;
    setIsExporting(true);
    setExportProgress(0);

    const baseUrl =
      process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

    try {
      // Build the active slide payload preserving order.
      const slidesPayload = useDeckStore.getState().slideOrder
        .map(id => useDeckStore.getState().slidesById[id])
        .filter(Boolean);
      const tokensPayload = useDeckStore.getState().tokens;

      const res = await fetch(
        `${baseUrl}/barq.v1.HeliosService/ExportDeck`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Connect-Protocol-Version": "1",
          },
          body: JSON.stringify({
            requestId,
            format: "pptx",
            slides: slidesPayload,
            tokens: tokensPayload,
          }),
        }
      );

      if (!res.ok) {
        console.error("Export failed:", res.status);
        return;
      }

      const reader = res.body?.getReader();
      if (!reader) return;

      const chunks: Uint8Array[] = [];
      const decoder = new TextDecoder();
      let buffer = "";
      let totalBytes = 0;
      let filename = "barq-slides-deck.pptx";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed) continue;

          try {
            const envelope = JSON.parse(trimmed);
            const msg = envelope.result ?? envelope;

            if (msg.totalBytes) totalBytes = msg.totalBytes;
            if (msg.filename) filename = msg.filename;

            // Data is base64-encoded in the JSON response.
            if (msg.data) {
              const binary = atob(msg.data);
              const bytes = new Uint8Array(binary.length);
              for (let i = 0; i < binary.length; i++) {
                bytes[i] = binary.charCodeAt(i);
              }
              chunks.push(bytes);

              if (totalBytes > 0) {
                const received = chunks.reduce((s, c) => s + c.length, 0);
                setExportProgress(Math.min(100, Math.round((received / totalBytes) * 100)));
              }
            }
          } catch {
            // Partial JSON.
          }
        }
      }

      // Assemble chunks into a blob and download.
      if (chunks.length > 0) {
        const blob = new Blob(chunks as BlobPart[], {
          type: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }
    } catch (err) {
      console.error("Export error:", err);
    } finally {
      setIsExporting(false);
      setExportProgress(0);
    }
  };

  if (!isReady && !isExporting) return null;

  return (
    <div className="flex items-center gap-3" id="download-pptx">
      <button
        id="download-pptx-btn"
        onClick={handleExport}
        disabled={isExporting}
        className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-primary to-primary/80 text-primary-foreground text-sm font-medium shadow-sm hover:shadow-md hover:from-primary/95 hover:to-primary/75 disabled:opacity-60 transition-all duration-200 active:scale-[0.98]"
      >
        {isExporting ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            {exportProgress > 0 ? `${exportProgress}%` : "Exporting..."}
          </>
        ) : (
          <>
            <Download className="h-4 w-4" />
            Download PPTX
          </>
        )}
      </button>

      {plan && (
        <span className="text-xs text-muted-foreground/60">
          <FileDown className="h-3 w-3 inline mr-1" />
          {slideCount} slides
        </span>
      )}
    </div>
  );
}
