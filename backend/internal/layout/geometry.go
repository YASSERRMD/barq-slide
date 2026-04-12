// Package layout defines the complete set of slide layout geometries
// for the barq-slides dual-render engine. All measurements are in EMU
// (English Metric Units: 914400 per inch) for pixel-perfect PPTX parity.
package layout

import (
	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// EMU constants for a 10" × 5.625" (16:9) slide canvas.
const (
	SlideWidthEMU  int64 = 9144000 // 10 inches
	SlideHeightEMU int64 = 5143500 // 5.625 inches

	// Standard margins.
	MarginTopEMU    int64 = 457200  // 0.5"
	MarginBottomEMU int64 = 457200  // 0.5"
	MarginLeftEMU   int64 = 457200  // 0.5"
	MarginRightEMU  int64 = 457200  // 0.5"

	// Safe content area.
	ContentWidthEMU  int64 = SlideWidthEMU - MarginLeftEMU - MarginRightEMU   // 9"
	ContentHeightEMU int64 = SlideHeightEMU - MarginTopEMU - MarginBottomEMU  // 4.625"

	// Column/gutter helpers.
	GutterEMU    int64 = 228600  // 0.25"
	ColWidth1    int64 = ContentWidthEMU
	ColWidth2    int64 = (ContentWidthEMU - GutterEMU) / 2
	ColWidth3    int64 = (ContentWidthEMU - 2*GutterEMU) / 3

	// Title heights.
	TitleHeightEMU    int64 = 800000
	SubtitleHeightEMU int64 = 457200

	// Header band (slide title bar).
	HeaderHeightEMU int64 = 685800  // 0.75"
	BodyTopEMU      int64 = MarginTopEMU + HeaderHeightEMU + GutterEMU
	BodyHeightEMU   int64 = SlideHeightEMU - BodyTopEMU - MarginBottomEMU
)

// Region is a 2D bounding box in EMU coordinates.
type Region struct {
	X      int64
	Y      int64
	Width  int64
	Height int64
}

// Geometry describes the regions of a slide layout.
type Geometry struct {
	LayoutID barqv1.LayoutID

	// Title region (always present).
	Title Region

	// Subtitle / tagline (optional; used for title-slide layouts).
	Subtitle Region

	// Body regions for content (1–3 columns).
	Body []Region

	// Media region (for layouts with a full image/chart zone).
	Media Region

	// Note: ColumnCount and HasMediaRegion are derived from the layout.
	ColumnCount    int
	HasMediaRegion bool
}

// All returns all defined layout geometries.
func All() []Geometry {
	return allGeometries
}

// Find returns the Geometry for a LayoutID, or the TITLE_CENTER fallback.
func Find(id barqv1.LayoutID) Geometry {
	for _, g := range allGeometries {
		if g.LayoutID == id {
			return g
		}
	}
	return allGeometries[0] // fallback: TITLE_CENTER
}

// allGeometries defines the 20 named layout geometries.
var allGeometries = []Geometry{

	// 1. TITLE_CENTER — full-bleed centered title slide.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER,
		Title: Region{
			X: MarginLeftEMU,
			Y: SlideHeightEMU/2 - TitleHeightEMU - GutterEMU,
			Width: ContentWidthEMU, Height: TitleHeightEMU,
		},
		Subtitle: Region{
			X: MarginLeftEMU,
			Y: SlideHeightEMU/2 + GutterEMU,
			Width: ContentWidthEMU, Height: SubtitleHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: false,
	},

	// 2. TITLE_LEFT — title left-aligned with accent bar.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_TITLE_LEFT,
		Title: Region{
			X: MarginLeftEMU,
			Y: SlideHeightEMU/3,
			Width: ContentWidthEMU, Height: TitleHeightEMU,
		},
		Subtitle: Region{
			X: MarginLeftEMU,
			Y: SlideHeightEMU/3 + TitleHeightEMU + GutterEMU,
			Width: ContentWidthEMU * 2 / 3, Height: SubtitleHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: false,
	},

	// 3. HERO_IMAGE — full-bleed image with overlaid title.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_HERO_IMAGE,
		Title: Region{
			X: MarginLeftEMU,
			Y: SlideHeightEMU*2/3,
			Width: ContentWidthEMU, Height: TitleHeightEMU,
		},
		Media: Region{X: 0, Y: 0, Width: SlideWidthEMU, Height: SlideHeightEMU},
		ColumnCount: 0, HasMediaRegion: true,
	},

	// 4. TWO_COLUMN — equal left/right content columns.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU},
			{X: MarginLeftEMU + ColWidth2 + GutterEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU},
		},
		ColumnCount: 2, HasMediaRegion: false,
	},

	// 5. THREE_COLUMN — equal three-column layout.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_THREE_COLUMN,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + ColWidth3 + GutterEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + 2*(ColWidth3+GutterEMU), Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
		},
		ColumnCount: 3, HasMediaRegion: false,
	},

	// 6. BULLET_LIST — classic title + bullets.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_BULLET_LIST,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ContentWidthEMU, Height: BodyHeightEMU},
		},
		ColumnCount: 1, HasMediaRegion: false,
	},

	// 7. NUMBERED_LIST — same geometry as BULLET_LIST, different render style.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_NUMBERED_LIST,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ContentWidthEMU, Height: BodyHeightEMU},
		},
		ColumnCount: 1, HasMediaRegion: false,
	},

	// 8. QUOTE_FULL — large centered pull quote.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL,
		Title: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU*2/5,
			Width: ContentWidthEMU, Height: TitleHeightEMU * 2,
		},
		Subtitle: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU*2/5 + TitleHeightEMU*2 + GutterEMU,
			Width: ContentWidthEMU, Height: SubtitleHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: false,
	},

	// 9. KPI_CARDS — row of 3–4 metric cards.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_KPI_CARDS,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + ColWidth3 + GutterEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + 2*(ColWidth3+GutterEMU), Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
		},
		ColumnCount: 3, HasMediaRegion: false,
	},

	// 10. CHART_FULL — full-width chart with minimal title.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_CHART_FULL,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Media: Region{
			X: MarginLeftEMU, Y: BodyTopEMU,
			Width: ContentWidthEMU, Height: BodyHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: true,
	},

	// 11. CHART_WITH_TEXT — 60% chart + 40% text.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_CHART_WITH_TEXT,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: ContentWidthEMU*3/5 + MarginLeftEMU + GutterEMU, Y: BodyTopEMU,
				Width: ContentWidthEMU * 2 / 5, Height: BodyHeightEMU},
		},
		Media: Region{
			X: MarginLeftEMU, Y: BodyTopEMU,
			Width: ContentWidthEMU * 3 / 5, Height: BodyHeightEMU,
		},
		ColumnCount: 1, HasMediaRegion: true,
	},

	// 12. DIAGRAM_FULL — full-canvas diagram.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_DIAGRAM_FULL,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Media: Region{
			X: MarginLeftEMU, Y: BodyTopEMU,
			Width: ContentWidthEMU, Height: BodyHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: true,
	},

	// 13. TIMELINE_HORIZONTAL — horizontal timeline with node labels.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_TIMELINE_HORIZONTAL,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: SlideHeightEMU/2 - BodyHeightEMU/4,
				Width: ContentWidthEMU, Height: BodyHeightEMU / 2},
		},
		ColumnCount: 1, HasMediaRegion: false,
	},

	// 14. TIMELINE_VERTICAL — vertical stepped timeline.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_TIMELINE_VERTICAL,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ContentWidthEMU, Height: BodyHeightEMU},
		},
		ColumnCount: 1, HasMediaRegion: false,
	},

	// 15. SECTION_DIVIDER — full-bleed section transition.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_SECTION_DIVIDER,
		Title: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU*2/5,
			Width: ContentWidthEMU, Height: TitleHeightEMU,
		},
		Subtitle: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU*2/5 + TitleHeightEMU + GutterEMU,
			Width: ContentWidthEMU / 2, Height: SubtitleHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: false,
	},

	// 16. COMPARISON_TABLE — side-by-side comparison.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_COMPARISON_TABLE,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU},
			{X: MarginLeftEMU + ColWidth2 + GutterEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU},
		},
		ColumnCount: 2, HasMediaRegion: false,
	},

	// 17. IMAGE_GRID — 2×2 image grid with caption.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_IMAGE_GRID,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU / 2},
			{X: MarginLeftEMU + ColWidth2 + GutterEMU, Y: BodyTopEMU, Width: ColWidth2, Height: BodyHeightEMU / 2},
			{X: MarginLeftEMU, Y: BodyTopEMU + BodyHeightEMU/2 + GutterEMU, Width: ColWidth2, Height: BodyHeightEMU / 2},
			{X: MarginLeftEMU + ColWidth2 + GutterEMU, Y: BodyTopEMU + BodyHeightEMU/2 + GutterEMU, Width: ColWidth2, Height: BodyHeightEMU / 2},
		},
		ColumnCount: 2, HasMediaRegion: true,
	},

	// 18. ICON_GRID — 3-column icon + label + description grid.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_ICON_GRID,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + ColWidth3 + GutterEMU, Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
			{X: MarginLeftEMU + 2*(ColWidth3+GutterEMU), Y: BodyTopEMU, Width: ColWidth3, Height: BodyHeightEMU},
		},
		ColumnCount: 3, HasMediaRegion: false,
	},

	// 19. PROCESS_FLOW — horizontal left→right flow steps.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_PROCESS_FLOW,
		Title: Region{
			X: MarginLeftEMU, Y: MarginTopEMU,
			Width: ContentWidthEMU, Height: HeaderHeightEMU,
		},
		Body: []Region{
			{X: MarginLeftEMU, Y: BodyTopEMU, Width: ContentWidthEMU, Height: BodyHeightEMU},
		},
		ColumnCount: 1, HasMediaRegion: false,
	},

	// 20. CLOSING — CTA + contact info.
	{
		LayoutID: barqv1.LayoutID_LAYOUT_ID_CLOSING,
		Title: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU/3,
			Width: ContentWidthEMU, Height: TitleHeightEMU,
		},
		Subtitle: Region{
			X: MarginLeftEMU, Y: SlideHeightEMU/3 + TitleHeightEMU + GutterEMU,
			Width: ContentWidthEMU, Height: SubtitleHeightEMU,
		},
		ColumnCount: 0, HasMediaRegion: false,
	},
}
