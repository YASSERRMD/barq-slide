"use client";

import { useDeckStore } from "./deck-store";

/**
 * Stream event constants matching the backend GenerateStreamEvent enum.
 * These mirror the protobuf values sent by the Go server.
 */
const StreamEvent = {
  UNSPECIFIED: 0,
  STARTED: 1,
  PLAN_READY: 2,
  TOKENS_READY: 3,
  SLIDE_STARTED: 4,
  SLIDE_READY: 5,
  ASSET_READY: 6,
  SNAPSHOT_READY: 7,
  COMPLETED: 8,
  ERROR: 9,
} as const;

/**
 * HeliosService path and method descriptor for server-streaming GenerateDeck.
 * This avoids needing buf-generated descriptors — we call the Connect
 * protocol directly using fetch.
 */
const GENERATE_DECK_PATH = "/barq.v1.HeliosService/GenerateDeck";

interface StreamResponseJSON {
  requestId?: string;
  event?: number;
  message?: string;
  progressPct?: number;
  plan?: Record<string, unknown>;
  tokens?: Record<string, unknown>;
  slide?: Record<string, unknown>;
  snapshotPng?: string; // base64
  slides?: Record<string, unknown>[];
  errorMessage?: string;
  errorCode?: string;
}

/**
 * useGenerateDeck returns a trigger function that kicks off deck generation
 * and maps every server-streaming event into the Zustand DeckState.
 *
 * @example
 * const generate = useGenerateDeck();
 * generate({ prompt: "Quarterly revenue analysis", audience: "board" });
 */
export function useGenerateDeck() {
  const store = useDeckStore.getState;

  return async (intent: {
    requestId?: string;
    prompt: string;
    audience?: string;
    tone?: string;
    domain?: string;
    totalSlides?: number;
  }) => {
    const requestId = intent.requestId ?? crypto.randomUUID();

    // Reset store for a fresh run.
    store().reset();
    store().setRequestId(requestId);
    store().setProgress({
      status: "started",
      message: "Connecting to server...",
      progressPct: 0,
    });

    const baseUrl =
      process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

    try {
      const res = await fetch(`${baseUrl}${GENERATE_DECK_PATH}`, {
        method: "POST",
        headers: {
          // Connect server-streaming requires application/connect+json, NOT application/json.
          // application/json is only valid for unary RPCs.
          "Content-Type": "application/connect+json",
          "Connect-Protocol-Version": "1",
          "X-Llm-Provider": store().llmProvider || "anthropic",
          "X-Llm-Api-Key": store().llmApiKey || "",
          "X-Llm-Model": store().llmModel || "",
          "X-Llm-Base-Url": store().llmBaseUrl || "",
        },
        body: JSON.stringify({
          intent: {
            requestId,
            prompt: intent.prompt,
            audience: intent.audience ?? "",
            tone: intent.tone ?? "",
            domain: intent.domain ?? "",
            totalSlides: intent.totalSlides ?? 0,
          },
        }),
      });

      if (!res.ok) {
        store().setProgress({
          status: "error",
          message: `Server error: ${res.status} ${res.statusText}`,
          error: `HTTP ${res.status}`,
        });
        return;
      }

      const reader = res.body?.getReader();
      if (!reader) {
        store().setProgress({
          status: "error",
          message: "No response stream available",
          error: "NO_STREAM",
        });
        return;
      }

      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // Connect server-streaming sends newline-delimited JSON envelopes.
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed) continue;

          try {
            const envelope = JSON.parse(trimmed);
            // Connect wraps each message in { "result": { ... } }
            const msg: StreamResponseJSON =
              envelope.result ?? envelope;
            handleStreamEvent(msg);
          } catch {
            // Partial JSON line — accumulate more data.
          }
        }
      }

      // Process any remaining buffer.
      if (buffer.trim()) {
        try {
          const envelope = JSON.parse(buffer.trim());
          const msg: StreamResponseJSON = envelope.result ?? envelope;
          handleStreamEvent(msg);
        } catch {
          // Ignore trailing partial data.
        }
      }
    } catch (err) {
      store().setProgress({
        status: "error",
        message: err instanceof Error ? err.message : "Unknown error",
        error: "NETWORK_ERROR",
      });
    }
  };
}

/**
 * Maps a single streamed JSON response into the Zustand deck store.
 */
function handleStreamEvent(msg: StreamResponseJSON) {
  const store = useDeckStore.getState();
  const event = msg.event ?? StreamEvent.UNSPECIFIED;

  switch (event) {
    case StreamEvent.STARTED:
      store.setProgress({
        status: "started",
        message: msg.message ?? "Generation started",
        progressPct: msg.progressPct ?? 0,
      });
      break;

    case StreamEvent.PLAN_READY:
      if (msg.plan) {
        store.setPlan(msg.plan as any);
      }
      store.setProgress({
        status: "planning",
        message: msg.message ?? "Plan ready",
        progressPct: msg.progressPct ?? 20,
      });
      break;

    case StreamEvent.TOKENS_READY:
      if (msg.tokens) {
        store.setTokens(msg.tokens as any);
      }
      store.setProgress({
        status: "theming",
        message: msg.message ?? "Design tokens resolved",
        progressPct: msg.progressPct ?? 25,
      });
      break;

    case StreamEvent.SLIDE_STARTED:
      store.setProgress({
        status: "generating",
        message: msg.message ?? "Generating slide...",
        progressPct: msg.progressPct ?? 30,
      });
      break;

    case StreamEvent.SLIDE_READY:
      if (msg.slide) {
        store.upsertSlide(msg.slide as any);
      }
      store.setProgress({
        status: "generating",
        message: msg.message ?? "Slide ready",
        progressPct: msg.progressPct ?? 50,
      });
      break;

    case StreamEvent.SNAPSHOT_READY:
      if (msg.slide && (msg.slide as any).id && msg.snapshotPng) {
        store.setSnapshot(
          (msg.slide as any).id,
          `data:image/png;base64,${msg.snapshotPng}`
        );
      }
      break;

    case StreamEvent.COMPLETED:
      if (msg.slides && msg.slides.length > 0) {
        store.setCompleted(msg.slides as any);
      } else {
        store.setProgress({
          status: "completed",
          message: msg.message ?? "Deck complete",
          progressPct: 100,
        });
      }
      break;

    case StreamEvent.ERROR:
      store.setProgress({
        status: "error",
        message: msg.errorMessage ?? msg.message ?? "Generation failed",
        error: msg.errorCode ?? "UNKNOWN",
      });
      break;

    default:
      // Unknown event — update progress message if present.
      if (msg.message) {
        store.setProgress({ message: msg.message });
      }
  }
}
