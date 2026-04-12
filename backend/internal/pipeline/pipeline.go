// Package pipeline wires all generation phases into a concurrent channel-based
// pipeline that emits progress events suitable for server-streaming.
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/intent"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/YASSERRMD/barq-slides/internal/planner"
	"github.com/YASSERRMD/barq-slides/internal/validate"
)

// Event is a progress update emitted by the pipeline.
type Event struct {
	RequestID   string
	Event       barqv1.GenerateStreamEvent
	Message     string
	ProgressPct int32
	Slide       *barqv1.SlideNode  // non-nil for SLIDE_READY events
	Plan        *barqv1.DeckPlan   // non-nil for PLAN_READY events
	Tokens      *barqv1.DesignTokens // non-nil for TOKENS_READY events
	Err         error              // non-nil on FAILED events
}

// Config bundles all pipeline dependencies.
type Config struct {
	LLMProvider llm.Provider
	Log         *slog.Logger
}

// Run executes the full deck-generation pipeline for intent and streams
// Events on the returned channel. The channel is closed when the pipeline
// finishes or encounters a fatal error.
func Run(ctx context.Context, spec *barqv1.IntentSpec, cfg Config) <-chan Event {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		run(ctx, spec, cfg, ch)
	}()
	return ch
}

func emit(ch chan<- Event, ev Event) {
	select {
	case ch <- ev:
	default:
	}
}

func run(ctx context.Context, spec *barqv1.IntentSpec, cfg Config, ch chan<- Event) {
	reqID := spec.GetRequestId()
	log := cfg.Log.With(slog.String("request_id", reqID))

	emit(ch, Event{
		RequestID:   reqID,
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_STARTED,
		Message:     "Pipeline started",
		ProgressPct: 0,
	})

	// ── Phase 1: Parse intent ────────────────────────────────────────────────
	log.Info("parsing intent")
	parser := intent.NewParser(cfg.LLMProvider)
	parsed, err := parser.Parse(ctx, spec)
	if err != nil {
		emit(ch, Event{
			RequestID: reqID,
			Event:     barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_ERROR,
			Message:   fmt.Sprintf("intent parsing failed: %v", err),
			Err:       err,
		})
		return
	}
	emit(ch, Event{
		RequestID:   reqID,
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_PLAN_READY,
		Message:     "Intent parsed",
		ProgressPct: 10,
	})

	// ── Phase 2: Build narrative arc plan ───────────────────────────────────
	log.Info("building narrative arc plan")
	arcPlanner := planner.NewArcPlanner(cfg.LLMProvider)
	plan, err := arcPlanner.Plan(ctx, parsed)
	if err != nil {
		emit(ch, Event{
			RequestID: reqID,
			Event:     barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_ERROR,
			Message:   fmt.Sprintf("arc planning failed: %v", err),
			Err:       err,
		})
		return
	}
	emit(ch, Event{
		RequestID:   reqID,
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_PLAN_READY,
		Message:     fmt.Sprintf("Plan ready: %d sections", len(plan.GetSections())),
		ProgressPct: 20,
		Plan:        plan,
	})

	// ── Phase 3: Resolve design tokens ──────────────────────────────────────
	log.Info("resolving design tokens")
	tokens := defaultTokens()
	emit(ch, Event{
		RequestID:   reqID,
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_TOKENS_READY,
		Message:     "Design tokens resolved",
		ProgressPct: 25,
		Tokens:      tokens,
	})

	// ── Phase 4: Expand slides ───────────────────────────────────────────────
	log.Info("expanding slide nodes")
	expander := planner.NewExpander(cfg.LLMProvider)
	nodes, err := expander.Expand(ctx, plan, parsed)
	if err != nil {
		emit(ch, Event{
			RequestID: reqID,
			Event:     barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_ERROR,
			Message:   fmt.Sprintf("slide expansion failed: %v", err),
			Err:       err,
		})
		return
	}

	total := int32(len(nodes))
	if total == 0 {
		total = 1
	}

	// ── Phase 5: QA + emit per-slide events ─────────────────────────────────
	for i, node := range nodes {
		// Apply overflow auto-fix.
		node = validate.AutoFixOverflows(node, tokens)
		nodes[i] = node

		pct := int32(20) + int32(70)*(int32(i+1)/total)
		emit(ch, Event{
			RequestID:   reqID,
			Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_SLIDE_READY,
			Message:     fmt.Sprintf("Slide %d ready", node.GetIndex()),
			ProgressPct: pct,
			Slide:       node,
		})
	}

	emit(ch, Event{
		RequestID:   reqID,
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_COMPLETED,
		Message:     fmt.Sprintf("Deck complete: %d slides", len(nodes)),
		ProgressPct: 100,
	})
}

// defaultTokens returns a reasonable set of design tokens for the pipeline
// without requiring an LLM call. Callers that need brand-matched tokens
// should replace this with theme.GenerateTokens(mood).
func defaultTokens() *barqv1.DesignTokens {
	return &barqv1.DesignTokens{
		PrimaryHex:      "#4F46E5",
		SecondaryHex:    "#10B981",
		BackgroundHex:   "#FFFFFF",
		OnBackgroundHex: "#111827",
		SurfaceHex:      "#F9FAFB",
		MutedHex:        "#6B7280",
		Heading: &barqv1.Typography{
			FontFamily: "Inter",
			SizePt:     36,
		},
		Subheading: &barqv1.Typography{
			FontFamily: "Inter",
			SizePt:     20,
		},
		Body: &barqv1.Typography{
			FontFamily: "Inter",
			SizePt:     14,
		},
	}
}
