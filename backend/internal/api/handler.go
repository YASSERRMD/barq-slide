// Package api implements the ConnectRPC HeliosService handler.
package api

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/YASSERRMD/barq-slides/internal/logger"
	"github.com/YASSERRMD/barq-slides/internal/pipeline"
)

// Handler implements barqv1.HeliosServiceHandler.
type Handler struct {
	barqv1.UnimplementedHeliosServiceHandler
	log      *slog.Logger
	provider llm.Provider
}

// New creates a new Handler.
func New(log *slog.Logger, provider llm.Provider) *Handler {
	return &Handler{log: log, provider: provider}
}

// GenerateDeck streams the full deck generation pipeline.
func (h *Handler) GenerateDeck(
	ctx context.Context,
	req *connect.Request[barqv1.GenerateDeckRequest],
	stream *connect.ServerStream[barqv1.GenerateStreamResponse],
) error {
	spec := req.Msg.GetIntent()
	log := logger.FromContext(ctx).With(
		slog.String("request_id", spec.GetRequestID()),
		slog.String("method", "GenerateDeck"),
	)
	log.Info("deck generation started")

	cfg := pipeline.Config{
		LLMProvider: h.provider,
		Log:         log,
	}

	for ev := range pipeline.Run(ctx, spec, cfg) {
		if err := streamEvent(stream, ev); err != nil {
			return fmt.Errorf("sending stream event: %w", err)
		}
	}

	return nil
}

// RegenerateSlide regenerates a single slide.
func (h *Handler) RegenerateSlide(
	ctx context.Context,
	req *connect.Request[barqv1.RegenerateSlideRequest],
	stream *connect.ServerStream[barqv1.GenerateStreamResponse],
) error {
	logger.FromContext(ctx).Info("slide regeneration started",
		slog.String("request_id", req.Msg.GetRequestID()),
		slog.String("slide_id", req.Msg.GetSlideID()),
	)
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("slide regeneration coming soon"))
}

// ExportDeck exports a completed deck as a chunked PPTX/PDF stream.
func (h *Handler) ExportDeck(
	ctx context.Context,
	req *connect.Request[barqv1.ExportDeckRequest],
	stream *connect.ServerStream[barqv1.ExportChunkResponse],
) error {
	logger.FromContext(ctx).Info("deck export started",
		slog.String("request_id", req.Msg.GetRequestID()),
		slog.String("format", req.Msg.GetFormat()),
	)
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("export coming soon"))
}

// GetDeckPlan returns the deck plan for a completed or in-progress request.
func (h *Handler) GetDeckPlan(
	ctx context.Context,
	req *connect.Request[barqv1.GetDeckPlanRequest],
) (*connect.Response[barqv1.GetDeckPlanResponse], error) {
	logger.FromContext(ctx).Info("get deck plan",
		slog.String("request_id", req.Msg.GetRequestID()),
	)
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("deck plan lookup coming soon"))
}
