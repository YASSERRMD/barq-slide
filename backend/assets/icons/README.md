# Bundled Icon Sets

barq-slides bundles two OSS icon sets for offline rendering without external dependencies.

## Lucide Icons
- **License:** ISC
- **Source:** https://lucide.dev
- **Format:** SVG, viewBox="0 0 24 24", stroke-based
- **Count:** ~1400 icons
- **Usage:** General UI, technology, business concepts

## Phosphor Icons
- **License:** MIT
- **Source:** https://phosphoricons.com
- **Format:** SVG, viewBox="0 0 256 256", fill-based (regular weight)
- **Count:** ~1200 icons
- **Usage:** More expressive, presentation-style illustrations

## Download

Run `make icons` from the backend directory to download icon sets:

```bash
# Downloads lucide SVGs into assets/icons/lucide/
# Downloads phosphor regular SVGs into assets/icons/phosphor/
make icons
```

SVG files are excluded from git (see .gitignore) but cached in the Docker image.

## Naming Convention

Icon filenames match the package's canonical names:
- Lucide: `arrow-right.svg`, `chart-bar.svg`, `users.svg`
- Phosphor: `ArrowRight.svg`, `ChartBar.svg`, `Users.svg`

The semantic matcher normalizes both naming conventions.
