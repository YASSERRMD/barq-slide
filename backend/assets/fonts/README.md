# Bundled OSS Fonts

These fonts are used by the barq-slides dual-render engine:

| Font | License | Usage |
|------|---------|-------|
| Inter | OFL-1.1 | Corporate, Startup headings/body |
| Plus Jakarta Sans | OFL-1.1 | Creative, Startup headings |
| DM Sans | OFL-1.1 | Creative, Sustainable body |
| Space Grotesk | OFL-1.1 | Technology headings |
| Montserrat | OFL-1.1 | Marketing headings |
| IBM Plex Sans | OFL-1.1 | Finance headings/body |
| Source Sans 3 | OFL-1.1 | Academic, Government body |

## Download Instructions

Run `make fonts` from the backend directory to download these fonts from Google Fonts CDN into this directory. The `.ttf` files are excluded from git (see root .gitignore) and must be downloaded before building a PPTX.

In CI/CD, fonts are cached by the Docker image layer.

## Font Embedding

The PPTX assembler (Phase 17) embeds the TTF files directly into the `.pptx` archive so the resulting file is portable across systems without the fonts installed.

The HTML composer (Phase 15) loads fonts via CSS `@font-face` rules pointing to the `/fonts/` static asset path.
