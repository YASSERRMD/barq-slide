# barq-slides

**Next-Gen AI Presentation Engine** — Dual-Render Architecture

> Treats every deck as a structured visual argument. Plans it as a graph, assigns deterministic design tokens, generates SVG assets and Mermaid diagrams, then splits into two rendering paths:
>
> 1. **HTML/CSS engine** — real-time web preview (Tailwind / CSS Grid)
> 2. **Native OOXML engine** — 100% editable `.pptx` export via `unioffice`

## Monorepo Structure

```
barq-slides/
├── backend/          # Go 1.23+ — AI pipeline, PPTX assembler, ConnectRPC server
├── web/              # Next.js 14+ — App Router, Tailwind, shadcn/ui
├── proto/            # Protobuf definitions (buf)
└── .github/          # CI workflows
```

## Quick Start

```bash
# Backend
cd backend && make dev

# Frontend
cd web && npm run dev
```

## Architecture

See [BARQ_SLIDES_SPEC.md](./BARQ_SLIDES_SPEC.md) for the full specification.

## Tech Stack

| | Technology |
|---|---|
| Backend | Go 1.23+, ConnectRPC, unioffice, chromedp |
| LLM | Anthropic Claude, Google Gemini, OpenAI-compat (xAI / Minimax) |
| Frontend | Next.js 14, TypeScript, Tailwind CSS, shadcn/ui, Zustand |
| Contracts | Protobuf + buf |
