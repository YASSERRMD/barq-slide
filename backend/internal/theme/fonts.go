package theme

import (
	"fmt"
	"os"
	"path/filepath"
)

// FontEntry describes a bundled OSS font.
type FontEntry struct {
	// Family is the CSS/PPTX font-family name.
	Family string

	// FilenameRegular is the TTF filename for weight 400.
	FilenameRegular string

	// FilenameBold is the TTF filename for weight 700.
	FilenameBold string

	// License is the SPDX identifier.
	License string
}

// BundledFonts lists all 7 OSS fonts bundled with barq-slides.
var BundledFonts = []FontEntry{
	{
		Family:          "Inter",
		FilenameRegular: "Inter-Regular.ttf",
		FilenameBold:    "Inter-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "Plus Jakarta Sans",
		FilenameRegular: "PlusJakartaSans-Regular.ttf",
		FilenameBold:    "PlusJakartaSans-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "DM Sans",
		FilenameRegular: "DMSans-Regular.ttf",
		FilenameBold:    "DMSans-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "Space Grotesk",
		FilenameRegular: "SpaceGrotesk-Regular.ttf",
		FilenameBold:    "SpaceGrotesk-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "Montserrat",
		FilenameRegular: "Montserrat-Regular.ttf",
		FilenameBold:    "Montserrat-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "IBM Plex Sans",
		FilenameRegular: "IBMPlexSans-Regular.ttf",
		FilenameBold:    "IBMPlexSans-Bold.ttf",
		License:         "OFL-1.1",
	},
	{
		Family:          "Source Sans 3",
		FilenameRegular: "SourceSans3-Regular.ttf",
		FilenameBold:    "SourceSans3-Bold.ttf",
		License:         "OFL-1.1",
	},
}

// FontResolver resolves font file paths for a given font family and weight.
type FontResolver struct {
	fontsDir string
}

// NewFontResolver creates a resolver pointing at fontsDir.
func NewFontResolver(fontsDir string) *FontResolver {
	return &FontResolver{fontsDir: fontsDir}
}

// Resolve returns the absolute path to a TTF file for the given family and weight.
// weight >= 600 returns the bold variant; otherwise regular.
func (r *FontResolver) Resolve(family string, weight int) (string, error) {
	for _, f := range BundledFonts {
		if f.Family == family {
			filename := f.FilenameRegular
			if weight >= 600 {
				filename = f.FilenameBold
			}
			path := filepath.Join(r.fontsDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return "", fmt.Errorf("font file not found: %s (run `make fonts` to download)", path)
			}
			return path, nil
		}
	}
	// Fallback: try Inter.
	fallback := filepath.Join(r.fontsDir, "Inter-Regular.ttf")
	if _, err := os.Stat(fallback); err == nil {
		return fallback, nil
	}
	return "", fmt.Errorf("font family %q not found in bundled fonts", family)
}

// MoodFont returns the FontEntry for a mood's heading font.
func MoodFont(mood Mood) (FontEntry, bool) {
	pairing, ok := fontPairings[mood]
	if !ok {
		return FontEntry{}, false
	}
	for _, f := range BundledFonts {
		if f.Family == pairing.heading {
			return f, true
		}
	}
	return FontEntry{}, false
}
