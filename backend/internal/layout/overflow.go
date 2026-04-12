package layout

import (
	"math"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

const (
	// approxCharsPerEMUWidth estimates characters per EMU at body font size.
	// Derived from: (EMU width / 12700 pts_per_emu) / (font_size_pt * 0.6 avg_char_width)
	charsPerPtPerEMU = 1.0 / (12700 * 0.6)

	// linesPerEMUHeight estimates lines per EMU at body line height.
	linesPerBodyEMU = 1.0 / (16 * 1.6 * 12700) // font 16pt, line-height 1.6

	// maxBulletPoints is the maximum body blocks before splitting the slide.
	maxBulletPoints = 6

	// maxCharsPerBullet truncates individual bullet points to prevent overflow.
	maxCharsPerBullet = 200
)

// BoundingBox estimates whether content fits within a region at the given font size.
type BoundingBox struct {
	// Region is the target area in EMU.
	Region Region

	// FontSizePt is the rendered font size in points.
	FontSizePt float64

	// LineHeightMultiplier is the CSS line-height multiplier.
	LineHeightMultiplier float64
}

// OverflowResult describes how many lines fit and how many overflow.
type OverflowResult struct {
	// LinesCapacity is the maximum lines that fit in the region.
	LinesCapacity int

	// LinesRequired is the lines needed to render all content.
	LinesRequired int

	// Overflows is true when LinesRequired > LinesCapacity.
	Overflows bool

	// OverflowLineCount is the number of lines that would be cut off.
	OverflowLineCount int
}

// Estimate calculates whether text content fits in the bounding box.
func (bb BoundingBox) Estimate(text string) OverflowResult {
	heightPt := float64(bb.Region.Height) / 12700
	widthPt := float64(bb.Region.Width) / 12700

	lineHeightPt := bb.FontSizePt * bb.LineHeightMultiplier
	linesCapacity := int(math.Floor(heightPt / lineHeightPt))

	// Estimate chars per line.
	avgCharWidthPt := bb.FontSizePt * 0.52
	charsPerLine := int(math.Floor(widthPt / avgCharWidthPt))
	if charsPerLine < 1 {
		charsPerLine = 1
	}

	// Count lines needed.
	words := strings.Fields(text)
	lineCount := 1
	lineLen := 0
	for _, word := range words {
		wl := len(word) + 1
		if lineLen+wl > charsPerLine {
			lineCount++
			lineLen = wl
		} else {
			lineLen += wl
		}
	}

	overflow := lineCount > linesCapacity
	overflowLines := 0
	if overflow {
		overflowLines = lineCount - linesCapacity
	}

	return OverflowResult{
		LinesCapacity:     linesCapacity,
		LinesRequired:     lineCount,
		Overflows:         overflow,
		OverflowLineCount: overflowLines,
	}
}

// TruncateBlocks truncates bullet point text to prevent hard overflow.
// Returns modified blocks (does not mutate the originals).
func TruncateBlocks(blocks []*barqv1.ContentBlock) []*barqv1.ContentBlock {
	out := make([]*barqv1.ContentBlock, 0, len(blocks))
	for _, b := range blocks {
		if b.GetRole() != "body" {
			out = append(out, b)
			continue
		}
		text := b.GetText()
		if len(text) > maxCharsPerBullet {
			// Truncate at word boundary.
			text = truncateAtWord(text, maxCharsPerBullet)
		}
		out = append(out, &barqv1.ContentBlock{
			ID:       b.GetID(),
			Role:     b.GetRole(),
			Text:     text,
			Emphasis: b.GetEmphasis(),
		})
	}
	return out
}

// SplitOverflowSlide splits a SlideNode that has too many body bullets
// into two slides. Returns the original (trimmed) slide and an overflow slide.
// If no split is needed, returns only the original with ok=false.
func SplitOverflowSlide(node *barqv1.SlideNode) (primary, overflow *barqv1.SlideNode, split bool) {
	bodyBlocks := filterBlocksByRole(node.GetBlocks(), "body")
	if len(bodyBlocks) <= maxBulletPoints {
		return node, nil, false
	}

	// Split body blocks.
	primaryBlocks := bodyBlocks[:maxBulletPoints]
	overflowBlocks := bodyBlocks[maxBulletPoints:]

	// Non-body blocks (heading, kpi_value) go to primary only.
	nonBodyBlocks := filterBlocksExcludeRole(node.GetBlocks(), "body")

	primaryNode := &barqv1.SlideNode{
		ID:           node.GetID(),
		RequestID:    node.GetRequestID(),
		Index:        node.GetIndex(),
		Role:         node.GetRole(),
		Layout:       node.GetLayout(),
		Blocks:       append(nonBodyBlocks, primaryBlocks...),
		Assets:       node.GetAssets(),
		SpeakerNotes: node.GetSpeakerNotes(),
		BgColorHex:   node.GetBgColorHex(),
		Tokens:       node.GetTokens(),
	}

	// Overflow node: same heading, suffix "(cont'd)", excess bullets only.
	headingBlock := findFirstBlockByRole(node.GetBlocks(), "heading")
	overflowHeading := &barqv1.ContentBlock{
		ID:       headingBlock.GetID() + "_cont",
		Role:     "heading",
		Text:     headingBlock.GetText() + " (cont'd)",
		Emphasis: headingBlock.GetEmphasis(),
	}

	overflowNode := &barqv1.SlideNode{
		ID:           node.GetID() + "_overflow",
		RequestID:    node.GetRequestID(),
		Index:        node.GetIndex() + 1,
		Role:         node.GetRole(),
		Layout:       node.GetLayout(),
		Blocks:       append([]*barqv1.ContentBlock{overflowHeading}, overflowBlocks...),
		SpeakerNotes: "(continued)",
		BgColorHex:   node.GetBgColorHex(),
		Tokens:       node.GetTokens(),
	}

	return primaryNode, overflowNode, true
}

// ── helpers ────────────────────────────────────────────────────────────────────

func filterBlocksByRole(blocks []*barqv1.ContentBlock, role string) []*barqv1.ContentBlock {
	var out []*barqv1.ContentBlock
	for _, b := range blocks {
		if b.GetRole() == role {
			out = append(out, b)
		}
	}
	return out
}

func filterBlocksExcludeRole(blocks []*barqv1.ContentBlock, role string) []*barqv1.ContentBlock {
	var out []*barqv1.ContentBlock
	for _, b := range blocks {
		if b.GetRole() != role {
			out = append(out, b)
		}
	}
	return out
}

func findFirstBlockByRole(blocks []*barqv1.ContentBlock, role string) *barqv1.ContentBlock {
	for _, b := range blocks {
		if b.GetRole() == role {
			return b
		}
	}
	return &barqv1.ContentBlock{}
}

func truncateAtWord(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	// Find last space before maxChars.
	idx := strings.LastIndex(s[:maxChars], " ")
	if idx < 0 {
		idx = maxChars
	}
	return s[:idx] + "..."
}
