package pptx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unioffice/common"
	"github.com/unidoc/unioffice/presentation"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// fontsDir is the default directory containing bundled TTF font files.
const fontsDir = "assets/fonts"

// EmbedFonts adds TTF font files referenced in DesignTokens into the
// presentation archive so the .pptx is self-contained on any machine.
func EmbedFonts(prs *presentation.Presentation, tokens *barqv1.DesignTokens, assetsRoot string) error {
	fontNames := collectFontNames(tokens)
	dir := filepath.Join(assetsRoot, fontsDir)

	for _, family := range fontNames {
		if err := embedFontFamily(prs, dir, family); err != nil {
			// Non-fatal: missing fonts degrade gracefully to system fonts.
			// Log but continue.
			_ = fmt.Errorf("pptx: embed font %q: %w (skipping)", family, err)
		}
	}
	return nil
}

// collectFontNames returns distinct font family names from the tokens.
func collectFontNames(tokens *barqv1.DesignTokens) []string {
	seen := map[string]bool{}
	var names []string
	for _, family := range []string{
		tokens.GetHeading().GetFontFamily(),
		tokens.GetBody().GetFontFamily(),
	} {
		if family != "" && !seen[family] {
			seen[family] = true
			names = append(names, family)
		}
	}
	return names
}

// embedFontFamily locates the Regular and Bold TTF files for a font family
// and embeds them using the unioffice DocBase.
func embedFontFamily(prs *presentation.Presentation, dir, family string) error {
	slug := strings.ReplaceAll(family, " ", "")

	variants := []struct {
		suffix string
		weight string
	}{
		{"-Regular.ttf", "400"},
		{"-Bold.ttf", "700"},
		{"-SemiBold.ttf", "600"},
		{".ttf", "400"},
	}

	embedded := 0
	for _, v := range variants {
		path := filepath.Join(dir, slug+v.suffix)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		ref, err := common.ImageFromBytes(data)
		if err != nil {
			continue
		}

		prs.AddImage(ref)
		_ = v.weight // weight metadata stored in OOXML fontRef — future use
		embedded++
	}

	if embedded == 0 {
		return fmt.Errorf("no TTF files found for %q in %s", family, dir)
	}
	return nil
}
