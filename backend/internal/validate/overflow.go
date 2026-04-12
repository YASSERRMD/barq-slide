// Package validate provides slide quality-assurance utilities for the
// barq-slides engine, including overflow detection and auto-fix loops.
package validate

import (
	"strings"
	"unicode/utf8"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/layout"
)

// OverflowSeverity indicates how critical an overflow issue is.
type OverflowSeverity int

const (
	SeverityInfo    OverflowSeverity = iota // minor — fits but is tight
	SeverityWarning                         // truncation likely in some renderers
	SeverityCritical                        // hard overflow, must be fixed
)

// OverflowIssue describes a single overflow or content-fit problem on a slide.
type OverflowIssue struct {
	SlideIndex int32
	BlockRole  string
	Message    string
	Severity   OverflowSeverity
}

// maxRunesByRegion returns the approximate maximum character count that fits
// in a Region at the given font size (points) without overflow.
// This is a heuristic: assumes average glyph width of ~0.55 × font size
// in EMU and line height of ~1.4 × font size.
func maxRunesByRegion(r layout.Region, fontSizePt float64) int {
	if fontSizePt <= 0 || r.Width <= 0 || r.Height <= 0 {
		return 0
	}
	// 1 pt = 12700 EMU
	const emuPerPt = 12700.0
	glyphWidthEMU := fontSizePt * 0.55 * emuPerPt
	lineHeightEMU := fontSizePt * 1.4 * emuPerPt

	charsPerLine := float64(r.Width) / glyphWidthEMU
	lines := float64(r.Height) / lineHeightEMU
	return int(charsPerLine * lines)
}

// DetectOverflows inspects a SlideNode for content that exceeds its region
// boundaries and returns a list of issues.
func DetectOverflows(node *barqv1.SlideNode, tokens *barqv1.DesignTokens) []OverflowIssue {
	geo := layout.Find(node.GetLayout().GetLayoutID())
	var issues []OverflowIssue

	headingSizePt := 32.0
	if tokens.GetHeading().GetSizePt() > 0 {
		headingSizePt = float64(tokens.GetHeading().GetSizePt())
	}
	bodySizePt := 14.0
	if tokens.GetBody().GetSizePt() > 0 {
		bodySizePt = float64(tokens.GetBody().GetSizePt())
	}

	for _, block := range node.GetBlocks() {
		text := block.GetText()
		runeCount := utf8.RuneCountInString(text)

		switch block.GetRole() {
		case "heading":
			max := maxRunesByRegion(geo.Title, headingSizePt)
			if max > 0 && runeCount > max {
				sev := SeverityWarning
				if runeCount > max*2 {
					sev = SeverityCritical
				}
				issues = append(issues, OverflowIssue{
					SlideIndex: node.GetIndex(),
					BlockRole:  "heading",
					Message:    "heading text may overflow title region",
					Severity:   sev,
				})
			}
		case "body":
			if len(geo.Body) == 0 {
				continue
			}
			max := maxRunesByRegion(geo.Body[0], bodySizePt)
			if max > 0 && runeCount > max {
				sev := SeverityWarning
				if runeCount > max*2 {
					sev = SeverityCritical
				}
				issues = append(issues, OverflowIssue{
					SlideIndex: node.GetIndex(),
					BlockRole:  "body",
					Message:    "body text may overflow content region",
					Severity:   sev,
				})
			}
		}
	}

	return issues
}

// AutoFixOverflows attempts to trim content blocks that exceed region limits.
// It returns a copy of the node with truncated text; the original is unchanged.
func AutoFixOverflows(node *barqv1.SlideNode, tokens *barqv1.DesignTokens) *barqv1.SlideNode {
	issues := DetectOverflows(node, tokens)
	if len(issues) == 0 {
		return node
	}

	geo := layout.Find(node.GetLayout().GetLayoutID())
	headingSizePt := 32.0
	if tokens.GetHeading().GetSizePt() > 0 {
		headingSizePt = float64(tokens.GetHeading().GetSizePt())
	}
	bodySizePt := 14.0
	if tokens.GetBody().GetSizePt() > 0 {
		bodySizePt = float64(tokens.GetBody().GetSizePt())
	}

	// Build role → severity map.
	criticalRoles := map[string]bool{}
	for _, iss := range issues {
		if iss.Severity >= SeverityWarning {
			criticalRoles[iss.BlockRole] = true
		}
	}

	// Shallow-copy blocks; only truncate where necessary.
	newBlocks := make([]*barqv1.ContentBlock, 0, len(node.GetBlocks()))
	for _, b := range node.GetBlocks() {
		if !criticalRoles[b.GetRole()] {
			newBlocks = append(newBlocks, b)
			continue
		}

		maxRunes := 0
		switch b.GetRole() {
		case "heading":
			maxRunes = maxRunesByRegion(geo.Title, headingSizePt)
		case "body":
			if len(geo.Body) > 0 {
				maxRunes = maxRunesByRegion(geo.Body[0], bodySizePt)
			}
		}

		text := b.GetText()
		if maxRunes > 0 && utf8.RuneCountInString(text) > maxRunes {
			runes := []rune(text)
			// Leave room for the ellipsis.
			cutAt := maxRunes - 1
			if cutAt < 0 {
				cutAt = 0
			}
			text = string(runes[:cutAt]) + "…"
			// Trim at the last word boundary if possible.
			if idx := strings.LastIndexByte(text, ' '); idx > cutAt/2 {
				text = strings.TrimRight(text[:idx], " ") + "…"
			}
		}

		newBlocks = append(newBlocks, &barqv1.ContentBlock{
			Role: b.GetRole(),
			Text: text,
		})
	}

	return &barqv1.SlideNode{
		ID:           node.GetID(),
		RequestID:    node.GetRequestID(),
		Index:        node.GetIndex(),
		Role:         node.GetRole(),
		Layout:       node.GetLayout(),
		Blocks:       newBlocks,
		Assets:       node.GetAssets(),
		SpeakerNotes: node.GetSpeakerNotes(),
		BgColorHex:   node.GetBgColorHex(),
		Tokens:       node.GetTokens(),
	}
}
