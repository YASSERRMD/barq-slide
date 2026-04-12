// Package charts routes between native PPTX chart objects and SVG chart rendering.
package charts

import (
	"encoding/json"
	"fmt"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// RenderPath determines which rendering path to use for a chart.
type RenderPath string

const (
	// RenderPathNative uses unioffice native chart objects (bar, line, pie).
	// Best for charts that must be editable in PowerPoint.
	RenderPathNative RenderPath = "native"

	// RenderPathSVG uses go-echarts to render an SVG.
	// Used for complex charts (scatter, combined) that unioffice doesn't support natively.
	RenderPathSVG RenderPath = "svg"
)

// ChartData is the parsed chart data payload from an AssetSlot.
type ChartData struct {
	// Type is the chart type hint (bar, line, pie, scatter, mixed).
	Type string `json:"type"`

	// Title is the chart title.
	Title string `json:"title"`

	// Labels are the X-axis category labels.
	Labels []string `json:"labels"`

	// Series holds one or more data series.
	Series []DataSeries `json:"series"`

	// XAxisLabel is the X-axis title.
	XAxisLabel string `json:"x_axis_label"`

	// YAxisLabel is the Y-axis title.
	YAxisLabel string `json:"y_axis_label"`
}

// DataSeries is one series in a chart.
type DataSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

// Route determines the render path for a chart asset slot.
// Native path: bar, line, pie with ≤3 series.
// SVG path: scatter, mixed, or >3 series.
func Route(slot *barqv1.AssetSlot) RenderPath {
	assetType := slot.GetType()

	switch assetType {
	case barqv1.AssetType_ASSET_TYPE_CHART_BAR,
		barqv1.AssetType_ASSET_TYPE_CHART_LINE,
		barqv1.AssetType_ASSET_TYPE_CHART_PIE:
		// Check series count from chart data.
		if data, err := ParseChartData(slot.GetChartDataJson()); err == nil {
			if len(data.Series) <= 3 {
				return RenderPathNative
			}
		}
		return RenderPathSVG

	case barqv1.AssetType_ASSET_TYPE_CHART_SCATTER:
		return RenderPathSVG

	default:
		return RenderPathSVG
	}
}

// ParseChartData parses the JSON payload from an AssetSlot's ChartDataJson field.
func ParseChartData(jsonStr string) (*ChartData, error) {
	if jsonStr == "" {
		return &ChartData{}, nil
	}

	var data ChartData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("parsing chart data: %w", err)
	}
	return &data, nil
}

// SampleChartData generates plausible sample data from a natural language query.
// Used as a fallback when the LLM didn't provide structured chart data.
func SampleChartData(query string, assetType barqv1.AssetType) *ChartData {
	switch assetType {
	case barqv1.AssetType_ASSET_TYPE_CHART_BAR:
		return &ChartData{
			Type:       "bar",
			Title:      query,
			Labels:     []string{"2021", "2022", "2023", "2024"},
			YAxisLabel: "Value",
			Series: []DataSeries{
				{Name: "Series A", Values: []float64{42, 58, 71, 89}},
			},
		}
	case barqv1.AssetType_ASSET_TYPE_CHART_LINE:
		return &ChartData{
			Type:       "line",
			Title:      query,
			Labels:     []string{"Q1", "Q2", "Q3", "Q4"},
			YAxisLabel: "Value",
			Series: []DataSeries{
				{Name: "Trend", Values: []float64{30, 45, 60, 80}},
			},
		}
	case barqv1.AssetType_ASSET_TYPE_CHART_PIE:
		return &ChartData{
			Type:   "pie",
			Title:  query,
			Labels: []string{"Category A", "Category B", "Category C", "Other"},
			Series: []DataSeries{
				{Name: "Share", Values: []float64{40, 30, 20, 10}},
			},
		}
	default:
		return &ChartData{
			Type:   "bar",
			Title:  query,
			Labels: []string{"A", "B", "C", "D"},
			Series: []DataSeries{
				{Name: "Data", Values: []float64{25, 50, 75, 100}},
			},
		}
	}
}
