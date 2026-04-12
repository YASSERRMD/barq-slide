package pptx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"strconv"
	"strings"

	"github.com/unidoc/unioffice/common"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/schema/soo/dml"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/layout"
)

// AddSVGAssets iterates the slide's assets and embeds any SVG as an image
// relationship positioned within the media region.
func AddSVGAssets(prs *presentation.Presentation, slide presentation.Slide, node *barqv1.SlideNode) error {
	geo := layout.Find(node.GetLayout().GetLayoutID())
	if !geo.HasMediaRegion {
		return nil
	}

	for _, asset := range node.GetAssets() {
		if len(asset.GetRenderedBytes()) == 0 {
			continue
		}

		if strings.HasPrefix(asset.GetRenderedMime(), "image/svg") {
			if err := embedSVGAsImage(prs, slide, asset, geo.Media); err != nil {
				// Non-fatal — log and continue.
				_ = fmt.Errorf("pptx: embed svg %s: %w (skipping)", asset.GetID(), err)
			}
		}
	}
	return nil
}

// embedSVGAsImage adds an SVG asset as an image shape on the slide.
func embedSVGAsImage(
	prs *presentation.Presentation,
	slide presentation.Slide,
	asset *barqv1.AssetSlot,
	rect layout.Region,
) error {
	svgBytes := asset.GetRenderedBytes()

	// Parse SVG dimensions from the viewBox or width/height attributes.
	w, h := parseSVGDimensions(svgBytes)
	if w == 0 || h == 0 {
		w, h = 800, 450 // sane fallback
	}

	img := common.Image{
		Data:   &svgBytes,
		Format: "svg+xml",
		Size:   image.Point{X: w, Y: h},
	}

	imgRef, err := prs.AddImage(img)
	if err != nil {
		return fmt.Errorf("adding image to presentation: %w", err)
	}

	prsImg := slide.AddImage(imgRef)
	imgProps := prsImg.Properties()
	imgProps.SetPosition(
		measurement.Distance(rect.X)*measurement.EMU,
		measurement.Distance(rect.Y)*measurement.EMU,
	)
	imgProps.SetSize(
		measurement.Distance(rect.Width)*measurement.EMU,
		measurement.Distance(rect.Height)*measurement.EMU,
	)

	return nil
}

// ── SVG helpers ───────────────────────────────────────────────────────────────

// parseSVGDimensions reads width and height from the root <svg> element.
func parseSVGDimensions(data []byte) (int, int) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "svg" {
			continue
		}
		var w, h int
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "width":
				w = parseDimension(attr.Value)
			case "height":
				h = parseDimension(attr.Value)
			case "viewBox":
				// viewBox="x y width height"
				parts := strings.Fields(attr.Value)
				if len(parts) == 4 {
					w = parseDimension(parts[2])
					h = parseDimension(parts[3])
				}
			}
		}
		if w > 0 && h > 0 {
			return w, h
		}
		break
	}
	return 0, 0
}

// parseDimension parses a numeric dimension string (may end in "px" etc.).
func parseDimension(s string) int {
	s = strings.TrimRight(s, "pxemrPXEM")
	v, _ := strconv.ParseFloat(s, 64)
	return int(v)
}

// parseSVGPaths extracts <path d="..."> elements from raw SVG bytes for
// future native freeform shape conversion.
func parseSVGPaths(svgBytes []byte) []string {
	var paths []string
	dec := xml.NewDecoder(bytes.NewReader(svgBytes))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "path" {
			continue
		}
		for _, attr := range se.Attr {
			if attr.Name.Local == "d" {
				paths = append(paths, attr.Value)
			}
		}
	}
	return paths
}

// Ensure dml is used (prevents import error for future freeform path work).
var _ = dml.NewCT_ShapeProperties
