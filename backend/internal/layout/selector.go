package layout

import (
	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// Score represents a layout candidate with its computed fit score.
type Score struct {
	Geometry Geometry
	Score    float64
}

// Select picks the best layout for a SlideNode based on content density
// and asset requirements. Higher score = better fit.
func Select(node *barqv1.SlideNode) Geometry {
	scores := scoreAll(node)
	if len(scores) == 0 {
		return Find(barqv1.LayoutID_LAYOUT_ID_BULLET_LIST)
	}

	best := scores[0]
	for _, s := range scores[1:] {
		if s.Score > best.Score {
			best = s
		}
	}
	return best.Geometry
}

// scoreAll computes fitness scores for every geometry given the slide content.
func scoreAll(node *barqv1.SlideNode) []Score {
	role := node.GetRole()
	blocks := node.GetBlocks()
	assets := node.GetAssets()

	bodyBlockCount := countBodyBlocks(blocks)
	hasChart := hasAssetType(assets, barqv1.AssetType_ASSET_TYPE_CHART_BAR,
		barqv1.AssetType_ASSET_TYPE_CHART_LINE, barqv1.AssetType_ASSET_TYPE_CHART_PIE,
		barqv1.AssetType_ASSET_TYPE_CHART_SCATTER)
	hasDiagram := hasAssetType(assets, barqv1.AssetType_ASSET_TYPE_DIAGRAM_FLOWCHART,
		barqv1.AssetType_ASSET_TYPE_DIAGRAM_SEQUENCE, barqv1.AssetType_ASSET_TYPE_DIAGRAM_MINDMAP,
		barqv1.AssetType_ASSET_TYPE_DIAGRAM_GANTT)
	hasPhoto := hasAssetType(assets, barqv1.AssetType_ASSET_TYPE_PHOTO,
		barqv1.AssetType_ASSET_TYPE_ILLUSTRATION)
	iconCount := countAssetType(assets, barqv1.AssetType_ASSET_TYPE_ICON)
	kpiCount := countKPIBlocks(blocks)

	var scores []Score

	for _, g := range allGeometries {
		s := computeScore(g, role, bodyBlockCount, kpiCount, iconCount,
			hasChart, hasDiagram, hasPhoto)
		scores = append(scores, Score{Geometry: g, Score: s})
	}

	return scores
}

// computeScore returns a fitness value for a geometry given slide characteristics.
// Scoring uses additive bonuses and penalties on a 0–100 scale.
func computeScore(
	g Geometry,
	role barqv1.SlideRole,
	bodyBlocks, kpiCount, iconCount int,
	hasChart, hasDiagram, hasPhoto bool,
) float64 {
	score := 50.0

	switch role {
	case barqv1.SlideRole_SLIDE_ROLE_TITLE:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER ||
			g.LayoutID == barqv1.LayoutID_LAYOUT_ID_TITLE_LEFT {
			score += 50
		} else {
			score -= 40
		}

	case barqv1.SlideRole_SLIDE_ROLE_SECTION_DIVIDER:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_SECTION_DIVIDER {
			score += 50
		} else {
			score -= 40
		}

	case barqv1.SlideRole_SLIDE_ROLE_CLOSING:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_CLOSING {
			score += 50
		} else {
			score -= 30
		}

	case barqv1.SlideRole_SLIDE_ROLE_IMAGE_HERO:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_HERO_IMAGE {
			score += 50
		} else {
			score -= 30
		}

	case barqv1.SlideRole_SLIDE_ROLE_QUOTE:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL {
			score += 50
		} else {
			score -= 30
		}

	case barqv1.SlideRole_SLIDE_ROLE_KPI:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_KPI_CARDS {
			score += 50
		} else {
			score -= 20
		}

	case barqv1.SlideRole_SLIDE_ROLE_TIMELINE:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_TIMELINE_HORIZONTAL ||
			g.LayoutID == barqv1.LayoutID_LAYOUT_ID_TIMELINE_VERTICAL {
			score += 40
		} else {
			score -= 20
		}

	case barqv1.SlideRole_SLIDE_ROLE_COMPARISON:
		if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_COMPARISON_TABLE ||
			g.LayoutID == barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN {
			score += 40
		} else {
			score -= 15
		}

	case barqv1.SlideRole_SLIDE_ROLE_DATA:
		if hasChart {
			if g.HasMediaRegion && (g.LayoutID == barqv1.LayoutID_LAYOUT_ID_CHART_FULL ||
				g.LayoutID == barqv1.LayoutID_LAYOUT_ID_CHART_WITH_TEXT) {
				score += 40
			}
		}

	case barqv1.SlideRole_SLIDE_ROLE_DIAGRAM:
		if hasDiagram {
			if g.LayoutID == barqv1.LayoutID_LAYOUT_ID_DIAGRAM_FULL {
				score += 40
			}
		}
	}

	// Content density bonuses.
	if hasChart && g.HasMediaRegion {
		score += 15
	}
	if hasDiagram && g.HasMediaRegion {
		score += 15
	}
	if hasPhoto && g.HasMediaRegion {
		score += 10
	}

	// Icon grid.
	if iconCount >= 3 && g.LayoutID == barqv1.LayoutID_LAYOUT_ID_ICON_GRID {
		score += 20
	}
	if iconCount >= 3 && g.ColumnCount == 3 {
		score += 10
	}

	// KPI density.
	if kpiCount >= 2 && g.ColumnCount >= 2 {
		score += 10
	}

	// Body block density → column count match.
	switch {
	case bodyBlocks <= 3:
		if g.ColumnCount == 1 {
			score += 10
		}
	case bodyBlocks <= 5:
		if g.ColumnCount == 1 || g.ColumnCount == 2 {
			score += 5
		}
	default:
		if g.ColumnCount >= 2 {
			score += 5
		}
	}

	return score
}

// ── helpers ────────────────────────────────────────────────────────────────────

func countBodyBlocks(blocks []*barqv1.ContentBlock) int {
	count := 0
	for _, b := range blocks {
		if b.GetRole() == "body" {
			count++
		}
	}
	return count
}

func countKPIBlocks(blocks []*barqv1.ContentBlock) int {
	count := 0
	for _, b := range blocks {
		if b.GetRole() == "kpi_value" {
			count++
		}
	}
	return count
}

func hasAssetType(assets []*barqv1.AssetSlot, types ...barqv1.AssetType) bool {
	for _, a := range assets {
		for _, t := range types {
			if a.GetType() == t {
				return true
			}
		}
	}
	return false
}

func countAssetType(assets []*barqv1.AssetSlot, t barqv1.AssetType) int {
	count := 0
	for _, a := range assets {
		if a.GetType() == t {
			count++
		}
	}
	return count
}
