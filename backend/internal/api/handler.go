// Package api implements the ConnectRPC HeliosService handler.
package api

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/logger"
)

// Handler implements barqv1.HeliosServiceHandler.
type Handler struct {
	barqv1.UnimplementedHeliosServiceHandler
	log *slog.Logger
}

// New creates a new Handler.
func New(log *slog.Logger) *Handler {
	return &Handler{log: log}
}

// GenerateDeck streams the full deck generation pipeline.
func (h *Handler) GenerateDeck(
	ctx context.Context,
	req *connect.Request[barqv1.GenerateDeckRequest],
	stream *connect.ServerStream[barqv1.GenerateStreamResponse],
) error {
	log := logger.FromContext(ctx).With(
		slog.String("request_id", req.Msg.Intent.GetRequestID()),
		slog.String("method", "GenerateDeck"),
	)
	log.Info("deck generation started")

	if err := stream.Send(&barqv1.GenerateStreamResponse{
		RequestID:   req.Msg.Intent.GetRequestID(),
		Event:       barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_STARTED,
		Message:     "Generation started",
		ProgressPct: 0,
	}); err != nil {
		return fmt.Errorf("sending start event: %w", err)
	}

	// Pipeline execution wired in Phase 22.
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
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("wired in phase 22"))
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
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("wired in phase 22"))
}

// GetDeckPlan returns the deck plan for a completed or in-progress request.
func (h *Handler) GetDeckPlan(
	ctx context.Context,
	req *connect.Request[barqv1.GetDeckPlanRequest],
) (*connect.Response[barqv1.GetDeckPlanResponse], error) {
	logger.FromContext(ctx).Info("get deck plan",
		slog.String("request_id", req.Msg.GetRequestID()),
	)
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("wired in phase 22"))
}
