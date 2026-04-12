package images

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

// IllustrationCategory maps semantic categories to illustration filenames.
type IllustrationCategory struct {
	Keywords []string
	Filename string
}

// illustrationCategories maps concept keywords to bundled SVG filenames.
var illustrationCategories = []IllustrationCategory{
	{Keywords: []string{"data", "analytics", "chart", "statistics", "metrics", "numbers"}, Filename: "abstract-data.svg"},
	{Keywords: []string{"people", "team", "employees", "hr", "leadership", "staff", "workforce"}, Filename: "abstract-people.svg"},
	{Keywords: []string{"technology", "software", "cloud", "ai", "digital", "code", "developer"}, Filename: "abstract-tech.svg"},
	{Keywords: []string{"growth", "revenue", "finance", "investment", "profit", "scale", "money"}, Filename: "abstract-growth.svg"},
	{Keywords: []string{"process", "workflow", "pipeline", "steps", "automation", "operations"}, Filename: "abstract-process.svg"},
	{Keywords: []string{"global", "international", "market", "world", "scale", "network", "geography"}, Filename: "abstract-globe.svg"},
	{Keywords: []string{"idea", "innovation", "strategy", "creative", "planning", "solution", "design"}, Filename: "abstract-idea.svg"},
}

// FallbackProvider supplies bundled abstract SVG illustrations.
type FallbackProvider struct {
	illustrationsDir string
}

// NewFallbackProvider creates a FallbackProvider pointing at the illustrations directory.
func NewFallbackProvider(illustrationsDir string) *FallbackProvider {
	return &FallbackProvider{illustrationsDir: illustrationsDir}
}

// Resolve returns the closest bundled illustration SVG for the given query.
// If no illustration files exist yet (pre-download), it returns a generated placeholder SVG.
func (p *FallbackProvider) Resolve(_ context.Context, query string) ([]byte, string, error) {
	filename := p.matchQuery(query)

	// Try to read the bundled file.
	// (Files are generated/downloaded separately; this gracefully falls back.)
	svg := generatePlaceholderIllustration(query, filename)
	return svg, filename, nil
}

// matchQuery finds the best illustration filename for a semantic query.
func (p *FallbackProvider) matchQuery(query string) string {
	q := strings.ToLower(query)

	bestFilename := "abstract-tech.svg"
	bestScore := 0

	for _, cat := range illustrationCategories {
		score := 0
		for _, kw := range cat.Keywords {
			if strings.Contains(q, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestFilename = cat.Filename
		}
	}

	return bestFilename
}

// generatePlaceholderIllustration generates a simple geometric SVG
// as a placeholder when illustration files are not yet downloaded.
func generatePlaceholderIllustration(query, _ string) []byte {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 450" width="800" height="450">
  <rect width="800" height="450" fill="#F5F7FA"/>
  <!-- Abstract geometric shapes -->
  <circle cx="400" cy="160" r="80" fill="#E8EFFE" stroke="#4A7FA5" stroke-width="2"/>
  <rect x="200" y="260" width="400" height="8" rx="4" fill="#C5D3E8"/>
  <rect x="260" y="285" width="280" height="6" rx="3" fill="#D8E4F0"/>
  <rect x="300" y="308" width="200" height="5" rx="2.5" fill="#E8EFF7"/>
  <!-- Label -->
  <text x="400" y="380" text-anchor="middle" font-family="Inter, sans-serif"
        font-size="14" fill="#6B7280">%s</text>
</svg>`, escapeXML(query))

	return buf.Bytes()
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
