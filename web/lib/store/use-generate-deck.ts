"use client";

import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { GenerateDeckRequestSchema, HeliosService } from "@/gen/barq/v1/service_pb";
import { IntentSpecSchema } from "@/gen/barq/v1/intent_pb";
import { useDeckStore } from "./deck-store";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8090";

const transport = createConnectTransport({ baseUrl: API_BASE_URL });

/**
 * useGenerateDeck returns a trigger function that streams a deck-generation
 * pipeline and maps every server-streaming event into the Zustand DeckState.
 */
export function useGenerateDeck() {
  return async (intent: {
    requestId?: string;
    prompt: string;
    audience?: string;
    tone?: string;
    domain?: string;
    totalSlides?: number;
  }) => {
    const state = useDeckStore.getState();
    const requestId = intent.requestId ?? crypto.randomUUID();

    state.reset();
    state.setRequestId(requestId);
    state.setProgress({
      status: "started",
      message: "Connecting to server...",
      progressPct: 0,
    });

    const client = createClient(HeliosService, transport);

    const request = create(GenerateDeckRequestSchema, {
      intent: create(IntentSpecSchema, {
        requestId,
        prompt: intent.prompt,
        audience: intent.audience ?? "",
        tone: intent.tone ?? "",
        domain: intent.domain ?? "",
        slideCount: intent.totalSlides ?? 10,
      }),
    });

    try {
      for await (const msg of client.generateDeck(request)) {
        const store = useDeckStore.getState();
        const event = Number(msg.event);

        switch (event) {
          case 1: // STARTED
            store.setProgress({
              status: "started",
              message: msg.message || "Generation started",
              progressPct: msg.progressPct ?? 0,
            });
            break;

          case 2: // PLAN_READY
            if (msg.plan) store.setPlan(msg.plan as any);
            store.setProgress({
              status: "planning",
              message: msg.message || "Plan ready",
              progressPct: msg.progressPct ?? 20,
            });
            break;

          case 3: // TOKENS_READY
            if (msg.tokens) store.setTokens(msg.tokens as any);
            store.setProgress({
              status: "theming",
              message: msg.message || "Tokens ready",
              progressPct: msg.progressPct ?? 25,
            });
            break;

          case 4: // SLIDE_STARTED
            store.setProgress({
              status: "generating",
              message: msg.message || "Generating slide...",
              progressPct: msg.progressPct ?? 30,
            });
            break;

          case 5: // SLIDE_READY
            if (msg.slide) store.upsertSlide(msg.slide as any);
            store.setProgress({
              status: "generating",
              message: msg.message || "Slide ready",
              progressPct: msg.progressPct ?? 50,
            });
            break;

          case 7: // SNAPSHOT_READY
            if (msg.slide && (msg.slide as any).id && msg.snapshotPng) {
              const raw = msg.snapshotPng;
              const b64 = typeof raw === "string" ? raw : btoa(Array.from(raw as Uint8Array, (b) => String.fromCharCode(b)).join(""));
              store.setSnapshot((msg.slide as any).id, `data:image/png;base64,${b64}`);
            }
            break;

          case 8: // COMPLETED
            if (msg.slides && msg.slides.length > 0) {
              store.setCompleted(msg.slides as any);
            } else {
              store.setProgress({
                status: "completed",
                message: msg.message || "Deck complete",
                progressPct: 100,
              });
            }
            break;

          case 9: // ERROR
            store.setProgress({
              status: "error",
              message: msg.errorMessage || msg.message || "Generation failed",
              error: msg.errorCode || "UNKNOWN",
            });
            break;

          default:
            if (msg.message) store.setProgress({ message: msg.message });
        }
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Network error";
      const code = (err as Record<string, unknown>)?.code;
      useDeckStore.getState().setProgress({
        status: "error",
        message,
        error: typeof code === "string" ? code : "NETWORK_ERROR",
      });
    }
  };
}
