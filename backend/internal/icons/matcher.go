// Package icons provides semantic icon matching and SVG recoloring.
// All icon rendering is text/SVG-based — NO emoji are ever used.
package icons

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Set identifies an icon library.
type Set string

const (
	SetLucide   Set = "lucide"
	SetPhosphor Set = "phosphor"
)

// IconRef resolves to a specific icon in a set.
type IconRef struct {
	Set      Set
	Name     string
	FilePath string
}

// semanticMap maps concept keywords to Lucide icon names.
// This is the primary lookup for semantic matching.
var semanticMap = map[string]string{
	// Business & Finance
	"money":       "dollar-sign",
	"revenue":     "trending-up",
	"cost":        "trending-down",
	"budget":      "wallet",
	"investment":  "piggy-bank",
	"profit":      "bar-chart-2",
	"growth":      "trending-up",
	"decline":     "trending-down",
	"bank":        "landmark",

	// Technology
	"api":        "code",
	"code":       "code-2",
	"database":   "database",
	"cloud":      "cloud",
	"server":     "server",
	"security":   "shield",
	"lock":       "lock",
	"key":        "key",
	"ai":         "cpu",
	"machine learning": "brain",
	"ml":         "cpu",
	"data":       "bar-chart",
	"analytics":  "line-chart",
	"chart":      "bar-chart-2",
	"network":    "network",
	"mobile":     "smartphone",
	"desktop":    "monitor",
	"email":      "mail",
	"message":    "message-square",
	"bug":        "bug",
	"deploy":     "rocket",
	"docker":     "container",

	// People & Organization
	"team":        "users",
	"user":        "user",
	"person":      "user",
	"people":      "users",
	"leader":      "user-check",
	"ceo":         "briefcase",
	"hr":          "heart-handshake",
	"hire":        "user-plus",
	"organization": "building",
	"company":     "building-2",

	// Process & Flow
	"process":     "workflow",
	"flow":        "git-branch",
	"pipeline":    "git-merge",
	"step":        "list-ordered",
	"checklist":   "check-square",
	"task":        "check-circle",
	"complete":    "check",
	"plan":        "clipboard",
	"roadmap":     "map",
	"timeline":    "calendar-range",

	// Communication
	"presentation": "presentation",
	"meeting":     "calendar",
	"call":        "phone",
	"video":       "video",
	"share":       "share-2",
	"feedback":    "message-circle",
	"review":      "eye",

	// Healthcare
	"health":      "heart-pulse",
	"medical":     "stethoscope",
	"hospital":    "hospital",
	"patient":     "user",
	"medicine":    "pill",
	"research":    "flask-conical",

	// Environment
	"sustainability": "leaf",
	"environment":  "tree-pine",
	"green":        "sprout",
	"energy":       "zap",
	"solar":        "sun",
	"recycle":      "recycle",

	// Generic
	"star":         "star",
	"award":        "award",
	"warning":      "alert-triangle",
	"error":        "x-circle",
	"success":      "check-circle-2",
	"info":         "info",
	"help":         "help-circle",
	"settings":     "settings",
	"home":         "home",
	"search":       "search",
	"download":     "download",
	"upload":       "upload",
	"link":         "link",
	"globe":        "globe",
	"location":     "map-pin",
	"time":         "clock",
	"calendar":     "calendar",
	"image":        "image",
	"file":         "file",
	"folder":       "folder",
	"arrow":        "arrow-right",
	"idea":         "lightbulb",
	"innovation":   "sparkles",
	"target":       "target",
	"goal":         "flag",
}

// Matcher resolves semantic queries to icon file paths.
type Matcher struct {
	iconsDir string
}

// NewMatcher creates a Matcher with the given icons base directory.
// Expected structure: {iconsDir}/lucide/*.svg  {iconsDir}/phosphor/*.svg
func NewMatcher(iconsDir string) *Matcher {
	return &Matcher{iconsDir: iconsDir}
}

// Resolve finds the best icon for a semantic query string.
// Returns the IconRef and the raw SVG bytes recolored with the given hex colour.
func (m *Matcher) Resolve(query, colorHex string) (*IconRef, []byte, error) {
	iconName := m.matchQuery(query)

	// Try Lucide first.
	lucidePath := filepath.Join(m.iconsDir, "lucide", iconName+".svg")
	if _, err := os.Stat(lucidePath); err == nil {
		raw, err := os.ReadFile(lucidePath)
		if err != nil {
			return nil, nil, fmt.Errorf("reading lucide icon %s: %w", iconName, err)
		}
		recolored := RecolorSVG(raw, colorHex)
		return &IconRef{Set: SetLucide, Name: iconName, FilePath: lucidePath}, recolored, nil
	}

	// Fallback: Phosphor (PascalCase filename).
	phosphorName := toPascalCase(iconName)
	phosphorPath := filepath.Join(m.iconsDir, "phosphor", phosphorName+".svg")
	if _, err := os.Stat(phosphorPath); err == nil {
		raw, err := os.ReadFile(phosphorPath)
		if err != nil {
			return nil, nil, fmt.Errorf("reading phosphor icon %s: %w", phosphorName, err)
		}
		recolored := RecolorSVG(raw, colorHex)
		return &IconRef{Set: SetPhosphor, Name: phosphorName, FilePath: phosphorPath}, recolored, nil
	}

	// Return a placeholder SVG if no file found (icons not downloaded yet).
	placeholder := placeholderSVG(colorHex)
	return &IconRef{Set: SetLucide, Name: "placeholder"}, placeholder, nil
}

// matchQuery maps a free-form query to the closest icon name.
func (m *Matcher) matchQuery(query string) string {
	q := strings.ToLower(strings.TrimSpace(query))

	// Direct match.
	if name, ok := semanticMap[q]; ok {
		return name
	}

	// Word overlap scoring.
	words := strings.Fields(q)
	bestName := "circle"
	bestScore := 0

	for keyword, name := range semanticMap {
		kwords := strings.Fields(keyword)
		score := 0
		for _, w := range words {
			for _, kw := range kwords {
				if strings.Contains(kw, w) || strings.Contains(w, kw) {
					score++
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestName = name
		}
	}

	return bestName
}

// RecolorSVG replaces stroke/fill colours in an SVG with the target hex.
// It handles both Lucide (stroke-based) and Phosphor (fill-based) formats.
func RecolorSVG(svgBytes []byte, colorHex string) []byte {
	if !strings.HasPrefix(colorHex, "#") {
		colorHex = "#" + colorHex
	}

	svg := string(svgBytes)

	// Replace currentColor references.
	svg = strings.ReplaceAll(svg, `currentColor`, colorHex)

	// Replace explicit black/dark strokes.
	strokeRe := regexp.MustCompile(`stroke="#[0-9a-fA-F]{3,6}"`)
	svg = strokeRe.ReplaceAllString(svg, `stroke="`+colorHex+`"`)

	// Replace explicit fills (but not fill="none").
	fillRe := regexp.MustCompile(`fill="#[0-9a-fA-F]{3,6}"`)
	svg = fillRe.ReplaceAllString(svg, `fill="`+colorHex+`"`)

	return []byte(svg)
}

// placeholderSVG returns a simple circle SVG as a fallback.
func placeholderSVG(colorHex string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" `+
		`stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">`+
		`<circle cx="12" cy="12" r="10"/></svg>`, colorHex)
	return buf.Bytes()
}

// toPascalCase converts "arrow-right" → "ArrowRight".
func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
