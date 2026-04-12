// Package api implements the ConnectRPC HeliosService handler.
package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/YASSERRMD/barq-slides/internal/logger"
	"github.com/YASSERRMD/barq-slides/internal/pipeline"
	"github.com/YASSERRMD/barq-slides/internal/pptx"
)

// Handler implements barqv1.HeliosServiceHandler.
type Handler struct {
	barqv1.UnimplementedHeliosServiceHandler
	log     *slog.Logger
	factory *llm.Factory
}

// New creates a new Handler.
func New(log *slog.Logger, factory *llm.Factory) *Handler {
	return &Handler{log: log, factory: factory}
}

// GenerateDeck streams the full deck generation pipeline.
func (h *Handler) GenerateDeck(
	ctx context.Context,
	req *connect.Request[barqv1.GenerateDeckRequest],
	stream *connect.ServerStream[barqv1.GenerateStreamResponse],
) error {
	spec := req.Msg.GetIntent()
	log := logger.FromContext(ctx).With(
		slog.String("request_id", spec.GetRequestId()),
		slog.String("method", "GenerateDeck"),
	)
	log.Info("deck generation started")

	// Prefer headers sent by the client; fall back to the backend-stored config.
	dynProvider := req.Header().Get("X-Llm-Provider")
	dynKey := req.Header().Get("X-Llm-Api-Key")
	dynModel := req.Header().Get("X-Llm-Model")
	dynBaseUrl := req.Header().Get("X-Llm-Base-Url")

	if dynProvider == "" {
		stored := globalLLMConfig.Get()
		dynProvider = stored.Provider
		dynKey = stored.APIKey
		dynModel = stored.Model
		dynBaseUrl = stored.BaseURL
	}

	log.Info("llm config resolved",
		slog.String("provider", dynProvider),
		slog.Bool("has_key", dynKey != ""),
		slog.String("model", dynModel),
		slog.String("base_url", dynBaseUrl),
	)

	var provider llm.Provider
	var err error
	if dynProvider != "" {
		provider, err = h.factory.GetDynamic(llm.ProviderName(dynProvider), dynKey, dynModel, dynBaseUrl)
	} else {
		provider, err = h.factory.Default()
	}
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid llm configuration: %w", err))
	}

	cfg := pipeline.Config{
		LLMProvider: provider,
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
		slog.String("request_id", req.Msg.GetRequestId()),
		slog.String("slide_id", req.Msg.GetSlideId()),
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
		slog.String("request_id", req.Msg.GetRequestId()),
		slog.String("format", req.Msg.GetFormat()),
	)

	slides := req.Msg.GetSlides()
	tokens := req.Msg.GetTokens()
	if len(slides) == 0 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no slides provided for export"))
	}

	// 1. Build the PPTX Presentation
	asm, err := pptx.NewAssembler(tokens)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to initialize assembler: %w", err))
	}

	pr := asm.Presentation()
	for _, node := range slides {
		if err := pptx.AddSlide(pr, node, tokens); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add slide %d: %w", node.GetIndex(), err))
		}
		
		// If KPI Cards layout, process charts
		if node.GetLayout() != nil && node.GetLayout().GetLayoutId() == barqv1.LayoutID_LAYOUT_ID_KPI_CARDS {
			addedSlides := pr.Slides()
			lastSlide := addedSlides[len(addedSlides)-1]
			if err := pptx.AddKPICards(lastSlide, node, tokens); err != nil {
				h.log.Warn("failed to add KPI cards", slog.Any("err", err))
			}
		}
	}

	// 2. Save directly to an in-memory buffer
	var buf bytes.Buffer
	if err := pr.Save(&buf); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save presentation: %w", err))
	}

	// 3. Stream back to client in chunks
	data := buf.Bytes()
	totalBytes := int64(len(data))
	const chunkSize = 64 * 1024 // 64KB per chunk
	filename := fmt.Sprintf("barq-slides-%s.pptx", req.Msg.GetRequestId())

	for offset := int64(0); offset < totalBytes; offset += chunkSize {
		end := offset + chunkSize
		if end > totalBytes {
			end = totalBytes
		}

		chunk := data[offset:end]
		isLast := end == totalBytes

		if err := stream.Send(&barqv1.ExportChunkResponse{
			TotalBytes: totalBytes,
			Offset:     offset,
			Data:       chunk,
			IsLast:     isLast,
			Filename:   filename,
		}); err != nil {
			return fmt.Errorf("streaming chunk offset %d: %w", offset, err)
		}
	}

	return nil
}

// GetDeckPlan returns the deck plan for a completed or in-progress request.
func (h *Handler) GetDeckPlan(
	ctx context.Context,
	req *connect.Request[barqv1.GetDeckPlanRequest],
) (*connect.Response[barqv1.GetDeckPlanResponse], error) {
	logger.FromContext(ctx).Info("get deck plan",
		slog.String("request_id", req.Msg.GetRequestId()),
	)
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("deck plan lookup coming soon"))
}
