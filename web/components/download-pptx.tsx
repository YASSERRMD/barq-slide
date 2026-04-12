"use client";

import { useState } from "react";
import { Download, FileDown, Loader2 } from "lucide-react";
import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ExportDeckRequestSchema, HeliosService } from "@/gen/barq/v1/service_pb";
import { useDeckStore } from "@/lib/store/deck-store";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";
const transport = createConnectTransport({ baseUrl: API_BASE_URL });
const client = createClient(HeliosService, transport);

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

    try {
      const state = useDeckStore.getState();
      const slidesPayload = state.slideOrder
        .map((id) => state.slidesById[id])
        .filter(Boolean) as any[];
      const tokensPayload = state.tokens as any;

      const request = create(ExportDeckRequestSchema, {
        requestId,
        format: "pptx",
        slides: slidesPayload,
        tokens: tokensPayload,
      });

      const chunks: Uint8Array[] = [];
      let totalBytes = 0;
      let filename = "barq-slides-deck.pptx";

      for await (const chunk of client.exportDeck(request)) {
        if (chunk.filename) filename = chunk.filename;
        if (chunk.totalBytes) totalBytes = Number(chunk.totalBytes);

        if (chunk.data && chunk.data.length > 0) {
          chunks.push(chunk.data);
          const received = chunks.reduce((s, c) => s + c.length, 0);
          if (totalBytes > 0) {
            setExportProgress(Math.min(100, Math.round((received / totalBytes) * 100)));
          }
        }
      }

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
            {exportProgress > 0 ? `${exportProgress}%` : "Exporting…"}
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
