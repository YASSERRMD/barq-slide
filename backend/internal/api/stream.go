// Package api provides streaming helpers for the ConnectRPC HeliosService.
package api

import (
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/pipeline"
)

// streamEvent converts a pipeline.Event into a GenerateStreamResponse and
// sends it on the provided ConnectRPC server stream. Each response is
// stamped with the current time.
func streamEvent(
	stream *connect.ServerStream[barqv1.GenerateStreamResponse],
	ev pipeline.Event,
) error {
	resp := &barqv1.GenerateStreamResponse{
		RequestId:   ev.RequestID,
		Event:       ev.Event,
		Message:     ev.Message,
		ProgressPct: ev.ProgressPct,
		Timestamp:   timestamppb.New(time.Now()),
	}

	if ev.Slide != nil {
		resp.Slide = ev.Slide
	}

	if len(ev.Slides) > 0 {
		resp.Slides = ev.Slides
	}

	if ev.Plan != nil {
		resp.Plan = ev.Plan
	}

	if ev.Tokens != nil {
		resp.Tokens = ev.Tokens
	}

	if ev.Err != nil {
		resp.ErrorMessage = ev.Err.Error()
		resp.ErrorCode = mapErrorCode(ev.Event)
	}

	return stream.Send(resp)
}

// mapErrorCode returns a machine-readable error code string for the given
// event type. Only ERROR events produce a non-empty code.
func mapErrorCode(ev barqv1.GenerateStreamEvent) string {
	switch ev {
	case barqv1.GenerateStreamEvent_GENERATE_STREAM_EVENT_ERROR:
		return "GENERATION_FAILED"
	default:
		return ""
	}
}
