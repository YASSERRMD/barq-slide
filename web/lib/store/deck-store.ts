"use client";

import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

// ── Minimal TS mirrors of the proto types (until buf generates TypeScript) ────

export interface Typography {
  fontFamily: string;
  sizePt: number;
  weight: number;
  lineHeight: number;
  isUppercase: boolean;
}

export interface DesignTokens {
  primaryHex: string;
  secondaryHex: string;
  backgroundHex: string;
  surfaceHex: string;
  onPrimaryHex: string;
  onBackgroundHex: string;
  mutedHex: string;
  heading?: Typography;
  subheading?: Typography;
  body?: Typography;
  caption?: Typography;
  mood: string;
}

export interface ContentBlock {
  id: string;
  role: string;
  text: string;
  emphasis: number;
}

export interface AssetSlot {
  id: string;
  type: number;
  query: string;
  iconName?: string;
  imageUrl?: string;
  mermaidSource?: string;
  chartDataJson?: string;
  renderedMime?: string;
}

export interface SlideLayout {
  layoutId: number;
  columnCount: number;
  hasMediaRegion: boolean;
  isDarkBackground: boolean;
}

export interface SlideNode {
  id: string;
  requestId: string;
  index: number;
  role: number;
  layout?: SlideLayout;
  blocks: ContentBlock[];
  assets: AssetSlot[];
  speakerNotes: string;
  bgColorHex: string;
  tokens?: DesignTokens;
}

export interface DeckSection {
  index: number;
  title: string;
  slideIds: string[];
}

export interface DeckPlan {
  requestId: string;
  title: string;
  subtitle: string;
  arc: number;
  sections: DeckSection[];
  totalSlides: number;
  topicTags: string[];
}

// ── Generation stream event types ─────────────────────────────────────────────

export type GenerationStatus =
  | "idle"
  | "started"
  | "planning"
  | "theming"
  | "generating"
  | "completed"
  | "error";

export interface GenerationProgress {
  status: GenerationStatus;
  message: string;
  progressPct: number;
  error?: string;
}

// ── Zustand store ─────────────────────────────────────────────────────────────

export interface DeckState {
  // ── Settings ─────────────────────────────────────────────────────────────────
  llmProvider: string;
  llmApiKey: string;

  // ── Generation State ─────────────────────────────────────────────────────────

  // Current active request ID.
  requestId: string | null;

  // Generation progress.
  progress: GenerationProgress;

  // Deck plan from the planner phase.
  plan: DeckPlan | null;

  // Design tokens from the theme phase.
  tokens: DesignTokens | null;

  // All slides produced so far (keyed by ID for O(1) updates).
  slidesById: Record<string, SlideNode>;

  // Ordered slide IDs (reflects final deck order).
  slideOrder: string[];

  // Index of the currently selected slide in the editor.
  selectedSlideIndex: number;

  // Snapshot PNGs keyed by slide ID (from chromedp QA renders).
  snapshotsById: Record<string, string>; // base64 data URLs

  // ── Actions ──────────────────────────────────────────────────────────────────

  setRequestId: (id: string) => void;
  setProgress: (p: Partial<GenerationProgress>) => void;
  setPlan: (plan: DeckPlan) => void;
  setTokens: (tokens: DesignTokens) => void;
  upsertSlide: (slide: SlideNode) => void;
  setSnapshot: (slideId: string, dataUrl: string) => void;
  setSelectedSlideIndex: (index: number) => void;
  setCompleted: (slides: SlideNode[]) => void;
  updateBlockText: (slideId: string, blockId: string, blockIndex: number, text: string) => void;
  setLlmConfig: (provider: string, apiKey: string) => void;
  reset: () => void;
}

const initialProgress: GenerationProgress = {
  status: "idle",
  message: "",
  progressPct: 0,
};

export const useDeckStore = create<DeckState>()(
  persist(
    devtools(
      (set) => ({
        llmProvider: "anthropic",
        llmApiKey: "",
        requestId: null,
        progress: initialProgress,
        plan: null,
        tokens: null,
        slidesById: {},
        slideOrder: [],
        selectedSlideIndex: 0,
        snapshotsById: {},

      setRequestId: (id) => set({ requestId: id }),

      setProgress: (p) =>
        set((state) => ({
          progress: { ...state.progress, ...p },
        })),

      setPlan: (plan) => set({ plan }),

      setTokens: (tokens) => set({ tokens }),

      upsertSlide: (slide) =>
        set((state) => {
          const exists = slide.id in state.slidesById;
          const slideOrder = exists
            ? state.slideOrder
            : [...state.slideOrder, slide.id];
          return {
            slidesById: { ...state.slidesById, [slide.id]: slide },
            slideOrder,
          };
        }),

      setSnapshot: (slideId, dataUrl) =>
        set((state) => ({
          snapshotsById: { ...state.snapshotsById, [slideId]: dataUrl },
        })),

      setSelectedSlideIndex: (index) => set({ selectedSlideIndex: index }),

      setCompleted: (slides) => {
        const slidesById: Record<string, SlideNode> = {};
        const slideOrder: string[] = [];
        slides.forEach((s) => {
          slidesById[s.id] = s;
          slideOrder.push(s.id);
        });
        set({
          slidesById,
          slideOrder,
          progress: { status: "completed", message: "Deck ready", progressPct: 100 },
        });
      },

      updateBlockText: (slideId, blockId, blockIndex, text) =>
        set((state) => {
          const slide = state.slidesById[slideId];
          if (!slide) return state;

          // Deep copy the blocks array and the specific block to maintain immutability.
          const newBlocks = [...slide.blocks];
          
          // Find the block either by ID or Index.
          let targetIndex = -1;
          if (blockId) {
            targetIndex = newBlocks.findIndex((b) => b.id === blockId);
          }
          if (targetIndex === -1) {
            targetIndex = blockIndex;
          }

          if (targetIndex >= 0 && targetIndex < newBlocks.length) {
            newBlocks[targetIndex] = { ...newBlocks[targetIndex], text };
            
            return {
              slidesById: {
                ...state.slidesById,
                [slideId]: { ...slide, blocks: newBlocks },
              },
            };
          }
          return state;
        }),

      setLlmConfig: (provider, apiKey) =>
        set({ llmProvider: provider, llmApiKey: apiKey }),
      reset: () =>
        set((state) => ({
          requestId: null,
          progress: initialProgress,
          plan: null,
          tokens: null,
          slidesById: {},
          slideOrder: [],
          selectedSlideIndex: 0,
          snapshotsById: {},
          // We intentionally do not reset llmProvider and llmApiKey here
        })),
    }),
    { name: "deck-store-dev" }
  ),
  {
    name: "barq-llm-config",
    partialize: (state) => ({
      llmProvider: state.llmProvider,
      llmApiKey: state.llmApiKey,
    }),
  }
)
);
