package pptx_test

import (
	"os"
	"strings"
	"testing"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/pptx"
)

// TestFixtureDeck verifies that a small deck can be assembled without errors
// and that the resulting OOXML bytes are non-empty.
func TestFixtureDeck(t *testing.T) {
	tokens := &barqv1.DesignTokens{
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

	asm, err := pptx.NewAssembler(tokens)
	if err != nil {
		t.Fatalf("NewAssembler: %v", err)
	}

	slides := []*barqv1.SlideNode{
		{
			Index: 1,
			Layout: &barqv1.SlideLayout{LayoutID: barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER},
			Blocks: []*barqv1.ContentBlock{
				{Role: "heading", Text: "Q4 Business Review"},
				{Role: "subtitle", Text: "Fiscal Year 2025"},
			},
		},
		{
			Index: 2,
			Layout: &barqv1.SlideLayout{LayoutID: barqv1.LayoutID_LAYOUT_ID_BULLET_LIST},
			Blocks: []*barqv1.ContentBlock{
				{Role: "heading", Text: "Key Achievements"},
				{Role: "body", Text: "Revenue grew 42% YoY"},
				{Role: "body", Text: "Customer NPS improved to 72"},
				{Role: "body", Text: "Launched 3 new product lines"},
			},
		},
		{
			Index: 3,
			Layout: &barqv1.SlideLayout{LayoutID: barqv1.LayoutID_LAYOUT_ID_KPI_CARDS},
			Blocks: []*barqv1.ContentBlock{
				{Role: "heading", Text: "Metrics Dashboard"},
				{Role: "body", Text: "Revenue | $4.2M | +42%"},
				{Role: "body", Text: "Users | 128K | +18%"},
				{Role: "body", Text: "NPS | 72 | +8pts"},
			},
		},
	}

	pr := asm.Presentation()
	for _, node := range slides {
		if err := pptx.AddSlide(pr, node, tokens); err != nil {
			t.Fatalf("AddSlide %d: %v", node.GetIndex(), err)
		}
		slides := pr.Slides()
		lastSlide := slides[len(slides)-1]
		if err := pptx.AddKPICards(lastSlide, node, tokens); err != nil {
			t.Fatalf("AddKPICards %d: %v", node.GetIndex(), err)
		}
	}

	tmp, err := os.CreateTemp("", "fixture-*.pptx")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := pr.SaveToFile(tmp.Name()); err != nil {
		if strings.Contains(err.Error(), "license") {
			t.Skipf("unioffice license not available: %v", err)
		}
		t.Fatalf("SaveToFile: %v", err)
	}

	info, err := os.Stat(tmp.Name())
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() < 1024 {
		t.Errorf("PPTX too small (%d bytes), likely empty", info.Size())
	}
	t.Logf("fixture deck written: %d bytes", info.Size())
}
