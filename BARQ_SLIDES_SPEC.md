# BARQ-SLIDES — Next-Gen AI Presentation Engine

## System Role
Principal Full-Stack Architect & AI Systems Engineer

## Runtime
- **Backend:** Go 1.23+
- **Frontend:** Next.js 14+ / React (App Router)
- **Architecture:** Dual-Render (HTML Preview + Native OOXML `.pptx` via `unioffice`)

---

## The Core Mandate

Every existing AI presentation tool fails in one of two ways:
- They build pretty web canvases that export to broken, flattened PPTX files (Gamma, Beautiful.ai)
- They build native PPTX files that are ugly and rely on rigid SmartArt (Copilot)

`barq-slides` is the **fourth path**: A dual-render engine. It treats a deck as a structured visual argument. It plans the deck as a graph, assigns a deterministic design token system, generates SVG assets/Mermaid diagrams, and then splits into TWO rendering paths:

1. An **HTML/CSS engine** (Bootstrap 5/Tailwind) for real-time web previews.
2. A **native Go OOXML engine** (`unioffice`) for the final, 100% editable `.pptx` export.

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Monorepo | `/backend` (Go) + `/web` (Next.js) |
| Contracts | Protobuf (`proto/`) compiled via `buf`. RPC via `connectrpc.com/connect` |
| PPTX | `unioffice` |
| Headless QA | `chromedp` |
| Charts | `go-echarts`, native unioffice charts |
| Diagrams | `mermaid-cli` subprocess |
| Spreadsheets | `excelize` |
| Config | `koanf` |
| LLM (Anthropic) | `github.com/anthropics/anthropic-sdk-go` |
| LLM (Gemini) | `github.com/google/genai-go` |
| LLM (OpenAI compat) | `sashabaranov/go-openai` (routes OpenAI / xAI / Minimax) |
| Frontend | TypeScript, App Router, Tailwind CSS, shadcn/ui, Zustand, `@connectrpc/connect-web` |

---

## Execution Protocol

Strictly phase-by-phase, commit-by-commit using Conventional Commits. Each atomic task is committed separately. Each phase is pushed to its own branch (`phase_1`, `phase_2`, …), then PR'd and merged to `main`.

---

## PART 1: INFRASTRUCTURE & DATA CONTRACTS

### PHASE 1: Monorepo Foundation
- `chore: init barq-slides monorepo with backend and web directories`
- `chore(backend): init go module barq-slides and Makefile`
- `chore(backend): add golangci-lint config`
- `chore(web): init next.js 14 app router with typescript`
- `ci: add github actions workflow for go and next.js`

### PHASE 2: Protobuf Source of Truth
- `feat(proto): define IntentSpec and DeckPlan messages`
- `feat(proto): define SlideNode, SlideLayout, DesignTokens, AssetSlot`
- `feat(proto): define HeliosService ConnectRPC streaming methods`
- `chore: configure buf.yaml and compile go and typescript stubs`

### PHASE 3 & 4: Scaffold Backend & Frontend
- `feat(backend/config): implement koanf loader and slog wrapper`
- `feat(backend/cmd): scaffold connectrpc server entrypoint`
- `chore(web): configure tailwindcss and initialize shadcn/ui`
- `feat(web/rpc): setup @connectrpc/connect-web client and zustand store`

---

## PART 2: THE AI PIPELINE & MULTI-PROVIDER LLM

### PHASE 5: Multi-Provider LLM Abstraction Layer (CRITICAL)
- `feat(backend/llm): define Provider interface (GenerateStructured, GenerateText) and Config`
- `feat(backend/llm): implement Anthropic adapter (force output_formatter Tool Use)`
- `feat(backend/llm): implement Gemini adapter (force ResponseSchema)`
- `feat(backend/llm): implement OpenAI-Compat adapter (sashabaranov/go-openai with ResponseFormatJSONSchema)`
- `feat(backend/llm): implement Factory pattern injecting BaseURL for xAI/Minimax routing`
- `test(backend/llm): add mock validation for multi-provider routing`

### PHASE 6 & 7: Intent & Content Planner
- `feat(backend/intent): implement prompt parser for audience, tone, domain`
- `feat(backend/planner): implement narrative arc generation`
- `feat(backend/planner): implement slide expansion and data/diagram flagging`
- `test(backend/planner): add golden prompt snapshot tests`

### PHASE 8: Theme & Design Tokens
- `feat(backend/theme): implement subject classifier to mood mapping`
- `feat(backend/theme): implement token generator (hex palettes, typography)`
- `feat(backend/theme): implement WCAG AA contrast validation/adjustment`
- `chore(backend/assets): bundle 7 OSS fonts and mapping logic`

### PHASE 9 & 10: Layout Library & Physics
- `feat(backend/layout): define 30+ layout geometries in EMUs (914400 per inch)`
- `feat(backend/layout/selector): implement scoring logic by content density`
- `feat(backend/layout/overflow): implement text bounding box calculation and split policy`

---

## PART 3: ASSET GENERATION ENGINES

### PHASE 11 & 12: Icons & Diagrams
- `chore(backend/assets): bundle lucide and phosphor SVG sets`
- `feat(backend/icons): implement semantic matching and XML recoloring (NO EMOJIS)`
- `feat(backend/diagrams): implement LLM to Mermaid syntax generator`
- `feat(backend/diagrams): implement mermaid-cli subprocess with token CSS`

### PHASE 13 & 14: Charts & Images
- `feat(backend/charts): implement router for Native PPTX vs SVG charts`
- `feat(backend/charts): implement go-echarts SVG renderer for complex charts`
- `feat(backend/images): implement Unsplash adapter with attribution`
- `chore(backend/assets): bundle abstract SVG illustrations as offline fallback`

---

## PART 4: THE DUAL RENDERER

### PHASE 15 & 16: HTML Composer & QA
- `feat(backend/compose): implement html/template engine mapping layouts to CSS Grid`
- `feat(backend/compose): inject DesignTokens as inline CSS variables`
- `feat(backend/render): implement chromedp tab pool for HTML to PNG snapshot rendering`

### PHASE 17 & 18: PPTX Assembler (Native OOXML)
- `feat(backend/pptx): compile unioffice SlideMaster from DesignTokens`
- `feat(backend/pptx): embed TTF font files into PPTX archive`
- `feat(backend/pptx): map EMU regions to native p:sp shapes and p:txBody text frames`
- `feat(backend/pptx): map SVG paths to native PPTX freeform shapes`

### PHASE 19: Native Charts & Assembly
- `feat(backend/pptx): compile unioffice native chart objects (bar/line/pie)`
- `feat(backend/pptx): implement logical shape grouping (e.g., KPI cards)`
- `test(backend/pptx): generate fixture deck and validate OOXML integrity`

---

## PART 5: ORCHESTRATION & API

### PHASE 20 & 21: QA & Adapters
- `feat(backend/validate): implement overflow detector and auto-fix loop`
- `feat(backend/adapters): implement excelize xlsx ingestor and stat profiler`

### PHASE 22: Streaming Orchestrator
- `feat(backend/pipeline): wire phases into concurrent pipeline channel`
- `feat(backend/api): implement GenerateDeck Connect server-streaming method`
- `feat(backend/api): stream GenerateStreamResponse events to frontend`

---

## PART 6: FRONTEND UX & STATE (NEXT.JS)

### PHASE 23 & 24: State & Prompt UX
- `feat(web/store): map Zustand state to streaming Connect responses`
- `feat(web/components): build chat-like PromptInput and TerminalProgress visualizer`

### PHASE 25 & 26: Web Canvas & Delivery
- `feat(web/canvas): build responsive CSS Grid renderer mapping to SlideLayout JSON`
- `feat(web/editor): implement Slide Thumbnail sidebar and Regenerate Slide RPC`
- `feat(web/editor): wire Download PPTX button to ExportDeck RPC`
- `chore(docker): create multi-stage Dockerfiles and docker-compose.yml`
