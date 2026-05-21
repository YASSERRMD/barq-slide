package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/pptx"
)

// exportHTTPRequest is the JSON body sent by the browser.
type exportHTTPRequest struct {
	RequestID string            `json:"requestId"`
	Slides    []json.RawMessage `json:"slides"`
	Tokens    json.RawMessage   `json:"tokens"`
}

// HandleExportHTTP accepts POST /api/export with JSON slides + tokens and
// returns the .pptx file as a raw binary download.
// This sidesteps all ConnectRPC binary-encoding issues.
func HandleExportHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50 MB max

	var body exportHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal design tokens (proto JSON — camelCase field names).
	var tokens barqv1.DesignTokens
	if len(body.Tokens) > 0 {
		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		if err := opts.Unmarshal(body.Tokens, &tokens); err != nil {
			// Not fatal — use zero tokens with default colours.
			_ = err
		}
	}

	// Unmarshal slides.
	var slides []*barqv1.SlideNode
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for _, raw := range body.Slides {
		var node barqv1.SlideNode
		if err := opts.Unmarshal(raw, &node); err == nil {
			slides = append(slides, &node)
		}
	}

	if len(slides) == 0 {
		http.Error(w, "no slides", http.StatusBadRequest)
		return
	}

	// Build PPTX.
	asm, err := pptx.NewAssembler(&tokens)
	if err != nil {
		http.Error(w, "assembler error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, node := range slides {
		if err := asm.AddSlide(node, &tokens); err != nil {
			http.Error(w, "slide error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var buf bytes.Buffer
	if err := asm.Save(&buf); err != nil {
		http.Error(w, "save error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := "barq-slides.pptx"
	if body.RequestID != "" {
		// Sanitize: keep only alphanumeric and hyphen chars (UUID-safe characters)
		// to prevent Content-Disposition header injection.
		safeID := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, body.RequestID)
		if safeID != "" {
			filename = fmt.Sprintf("barq-slides-%s.pptx", safeID)
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}
