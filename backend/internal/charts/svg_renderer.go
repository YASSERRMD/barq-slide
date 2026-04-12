package charts

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// SVGRenderer renders chart data to SVG using go-echarts.
type SVGRenderer struct{}

// NewSVGRenderer creates an SVGRenderer.
func NewSVGRenderer() *SVGRenderer { return &SVGRenderer{} }

// Render produces SVG bytes for the given chart data and design tokens.
func (r *SVGRenderer) Render(data *ChartData, tokens *barqv1.DesignTokens, assetType barqv1.AssetType) ([]byte, error) {
	switch assetType {
	case barqv1.AssetType_ASSET_TYPE_CHART_BAR:
		return r.renderBar(data, tokens)
	case barqv1.AssetType_ASSET_TYPE_CHART_LINE:
		return r.renderLine(data, tokens)
	case barqv1.AssetType_ASSET_TYPE_CHART_PIE:
		return r.renderPie(data, tokens)
	case barqv1.AssetType_ASSET_TYPE_CHART_SCATTER:
		return r.renderScatter(data, tokens)
	default:
		return r.renderBar(data, tokens)
	}
}

func (r *SVGRenderer) renderBar(data *ChartData, tokens *barqv1.DesignTokens) ([]byte, error) {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: data.Title}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:        "1200px",
			Height:       "600px",
			Renderer:     "svg",
			BackgroundColor: tokens.GetBackgroundHex(),
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: data.XAxisLabel}),
		charts.WithYAxisOpts(opts.YAxis{Name: data.YAxisLabel}),
		charts.WithColorsOpts(opts.Colors{tokens.GetPrimaryHex(), tokens.GetSecondaryHex()}),
	)

	bar.SetXAxis(data.Labels)

	for _, series := range data.Series {
		items := make([]opts.BarData, len(series.Values))
		for i, v := range series.Values {
			items[i] = opts.BarData{Value: v}
		}
		bar.AddSeries(series.Name, items)
	}

	return renderToSVG(bar)
}

func (r *SVGRenderer) renderLine(data *ChartData, tokens *barqv1.DesignTokens) ([]byte, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: data.Title}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:           "1200px",
			Height:          "600px",
			Renderer:        "svg",
			BackgroundColor: tokens.GetBackgroundHex(),
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: data.XAxisLabel}),
		charts.WithYAxisOpts(opts.YAxis{Name: data.YAxisLabel}),
		charts.WithColorsOpts(opts.Colors{tokens.GetPrimaryHex(), tokens.GetSecondaryHex()}),
	)

	line.SetXAxis(data.Labels)

	for _, series := range data.Series {
		items := make([]opts.LineData, len(series.Values))
		for i, v := range series.Values {
			items[i] = opts.LineData{Value: v}
		}
		line.AddSeries(series.Name, items,
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))
	}

	return renderToSVG(line)
}

func (r *SVGRenderer) renderPie(data *ChartData, tokens *barqv1.DesignTokens) ([]byte, error) {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: data.Title}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:           "1200px",
			Height:          "600px",
			Renderer:        "svg",
			BackgroundColor: tokens.GetBackgroundHex(),
		}),
	)

	items := make([]opts.PieData, 0)
	if len(data.Series) > 0 {
		for i, label := range data.Labels {
			if i < len(data.Series[0].Values) {
				items = append(items, opts.PieData{
					Name:  label,
					Value: data.Series[0].Values[i],
				})
			}
		}
	}

	pie.AddSeries("", items,
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"25%", "65%"},
		}),
	)

	return renderToSVG(pie)
}

func (r *SVGRenderer) renderScatter(data *ChartData, tokens *barqv1.DesignTokens) ([]byte, error) {
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: data.Title}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:           "1200px",
			Height:          "600px",
			Renderer:        "svg",
			BackgroundColor: tokens.GetBackgroundHex(),
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: data.XAxisLabel}),
		charts.WithYAxisOpts(opts.YAxis{Name: data.YAxisLabel}),
		charts.WithColorsOpts(opts.Colors{tokens.GetPrimaryHex(), tokens.GetSecondaryHex()}),
	)

	for i, series := range data.Series {
		items := make([]opts.ScatterData, len(series.Values))
		for j, v := range series.Values {
			x := 0.0
			if i < len(data.Labels) && j < len(data.Labels) {
				x = float64(j)
			}
			items[j] = opts.ScatterData{Value: []interface{}{x, v}}
		}
		scatter.AddSeries(series.Name, items)
	}

	return renderToSVG(scatter)
}

// renderToSVG renders an echarts chart to SVG bytes.
func renderToSVG(chart render.Renderer) ([]byte, error) {
	var buf bytes.Buffer
	if err := chart.Render(&buf); err != nil {
		return nil, fmt.Errorf("echarts render: %w", err)
	}

	// go-echarts renders an HTML page; extract the SVG element.
	html := buf.String()
	svgStart := strings.Index(html, "<svg")
	svgEnd := strings.LastIndex(html, "</svg>")

	if svgStart >= 0 && svgEnd > svgStart {
		return []byte(html[svgStart : svgEnd+6]), nil
	}

	// Fallback: return the full HTML if SVG extraction fails.
	return buf.Bytes(), nil
}
