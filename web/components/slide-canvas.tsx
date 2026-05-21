"use client";

import { FocusEvent, KeyboardEvent } from "react";
import type { SlideNode, ContentBlock, DesignTokens } from "@/lib/store/deck-store";
import { useDeckStore } from "@/lib/store/deck-store";

// ─── Proto LayoutID constants (must match slide.proto enum) ──────────────────
const LAYOUT_TITLE_CENTER   = 1;
const LAYOUT_BULLET_LIST    = 6;
const LAYOUT_QUOTE_FULL     = 8;
const LAYOUT_KPI_CARDS      = 9;
const LAYOUT_TWO_COLUMN     = 4;
const LAYOUT_SECTION_DIV    = 15;
const LAYOUT_CLOSING        = 20;

// ─── types ───────────────────────────────────────────────────────────────────

interface SlideCanvasProps {
  slide: SlideNode;
  tokens?: DesignTokens | null;
  /** CSS scale factor — 1 = full 960×540, 0.72 = main viewer, 0.17 = thumbnail */
  scale?: number;
  className?: string;
  isEditable?: boolean;
}

// ─── helpers ─────────────────────────────────────────────────────────────────

function hexBrightness(hex: string) {
  const m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!m) return 128;
  return (parseInt(m[1], 16) * 299 + parseInt(m[2], 16) * 587 + parseInt(m[3], 16) * 114) / 1000;
}

// ─── main component ──────────────────────────────────────────────────────────

/**
 * Renders a slide at a fixed 960×540 canvas.
 * The caller must wrap this in a clip container sized to 960*scale × 540*scale
 * with overflow:hidden so the CSS-transformed canvas is properly cropped.
 */
export function SlideCanvas({ slide, tokens, scale = 1, className = "", isEditable = false }: SlideCanvasProps) {
  const layoutId = slide.layout?.layoutId ?? LAYOUT_BULLET_LIST;

  const primary    = tokens?.primaryHex      || "#4F46E5";
  const secondary  = tokens?.secondaryHex    || "#10B981";
  const bg         = slide.bgColorHex || tokens?.backgroundHex || "#FFFFFF";
  const onBg       = tokens?.onBackgroundHex || "#111827";
  const muted      = tokens?.mutedHex        || "#6B7280";
  const surface    = tokens?.surfaceHex      || "#F9FAFB";

  const isDark     = hexBrightness(bg) < 128;
  const textColor  = isDark ? "#FFFFFF" : onBg;
  const mutedColor = isDark ? "rgba(255,255,255,0.55)" : muted;
  const cardBg     = isDark ? "rgba(255,255,255,0.10)" : surface;

  const headingFont = tokens?.heading?.fontFamily || "Inter, sans-serif";
  const bodyFont    = tokens?.body?.fontFamily    || "Inter, sans-serif";
  const headingSize = tokens?.heading?.sizePt     || 36;
  const subSize     = tokens?.subheading?.sizePt  || 20;
  const bodySize    = tokens?.body?.sizePt        || 15;

  // Block helpers
  const title      = slide.blocks.find(b => b.role === "heading"   || b.role === "title");
  const subtitle   = slide.blocks.find(b => b.role === "subheading"|| b.role === "subtitle");
  const bullets    = slide.blocks.filter(b => b.role === "bullet");
  const kpiBlocks  = slide.blocks.filter(b => b.role === "kpi_value");
  const quoteBlock = slide.blocks.find(b => b.role === "quote");
  const attrBlock  = slide.blocks.find(b => b.role === "attribution");

  const updateBlockText = useDeckStore.getState().updateBlockText;

  const wrap: React.CSSProperties = {
    position: "relative",
    width: "960px",
    height: "540px",
    backgroundColor: bg,
    color: textColor,
    fontFamily: bodyFont,
    overflow: "hidden",
    borderRadius: "8px",
    flexShrink: 0,
    transform: `scale(${scale})`,
    transformOrigin: "top left",
  };

  // ── TITLE / COVER ─────────────────────────────────────────────────────────
  if (layoutId === LAYOUT_TITLE_CENTER) {
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "6px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        {/* Decorative circles */}
        <div style={{ position: "absolute", bottom: -100, right: -100, width: 380, height: 380, borderRadius: "50%", background: primary, opacity: 0.07 }} />
        <div style={{ position: "absolute", bottom: -50, right: -50, width: 240, height: 240, borderRadius: "50%", background: secondary, opacity: 0.09 }} />
        <div style={{ position: "absolute", top: -80, left: -80, width: 300, height: 300, borderRadius: "50%", background: secondary, opacity: 0.05 }} />

        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: "72px 96px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize + 10}px`, fontWeight: 800, lineHeight: 1.1, letterSpacing: "-0.03em", color: textColor, marginBottom: "14px" }} />
          )}
          <div style={{ width: 64, height: 5, background: `linear-gradient(90deg, ${primary}, ${secondary})`, borderRadius: 3, marginBottom: 22 }} />
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${subSize + 2}px`, fontWeight: 400, color: mutedColor, lineHeight: 1.5, maxWidth: 580 }} />
          )}
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── SECTION DIVIDER ───────────────────────────────────────────────────────
  if (layoutId === LAYOUT_SECTION_DIV) {
    return (
      <div className={className} style={{ ...wrap, backgroundColor: primary }}>
        {/* Full-bleed gradient overlay */}
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(135deg, ${primary} 0%, ${secondary} 100%)`, opacity: 0.95 }} />
        {/* Large decorative number */}
        <div style={{ position: "absolute", right: 60, top: "50%", transform: "translateY(-50%)", fontSize: "200px", fontWeight: 900, color: "rgba(255,255,255,0.06)", lineHeight: 1, userSelect: "none", fontFamily: headingFont }}>
          {slide.index}
        </div>
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: "72px 88px", zIndex: 1 }}>
          <div style={{ fontSize: "13px", fontWeight: 700, letterSpacing: "0.15em", color: "rgba(255,255,255,0.6)", textTransform: "uppercase", marginBottom: 20 }}>
            SECTION
          </div>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize + 6}px`, fontWeight: 800, lineHeight: 1.15, color: "#FFFFFF", marginBottom: 16 }} />
          )}
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${subSize - 2}px`, color: "rgba(255,255,255,0.7)", lineHeight: 1.5 }} />
          )}
        </div>
      </div>
    );
  }

  // ── TWO COLUMN ────────────────────────────────────────────────────────────
  if (layoutId === LAYOUT_TWO_COLUMN) {
    // Split bullet blocks: first half left, second half right
    const half  = Math.ceil(bullets.length / 2);
    const leftB = bullets.slice(0, half);
    const rightB = bullets.slice(half);
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "5px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: "44px 52px 36px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 4}px`, fontWeight: 700, color: textColor, marginBottom: 4, lineHeight: 1.2 }} />
          )}
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${bodySize}px`, color: mutedColor, marginBottom: 20 }} />
          )}
          {!subtitle && <div style={{ height: 16 }} />}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, flex: 1, minHeight: 0 }}>
            <TwoColPanel blocks={leftB} slide={slide} accentColor={primary} cardBg={cardBg} textColor={textColor} bodySize={bodySize} isEditable={isEditable} update={updateBlockText} />
            <TwoColPanel blocks={rightB} slide={slide} accentColor={secondary} cardBg={cardBg} textColor={textColor} bodySize={bodySize} isEditable={isEditable} update={updateBlockText} />
          </div>
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── QUOTE ─────────────────────────────────────────────────────────────────
  if (layoutId === LAYOUT_QUOTE_FULL) {
    const qt = quoteBlock || bullets[0];
    const at = attrBlock || subtitle;
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", left: 0, top: 0, bottom: 0, width: 8, background: `linear-gradient(180deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", top: -20, right: -20, width: 320, height: 320, borderRadius: "50%", background: primary, opacity: 0.04 }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: "64px 88px" }}>
          <div style={{ fontSize: 140, lineHeight: 0.6, color: primary, opacity: 0.12, fontFamily: "Georgia, serif", marginBottom: 24, userSelect: "none" }}>"</div>
          {qt && (
            <Editable slideId={slide.id} block={qt} index={slide.blocks.indexOf(qt)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 4}px`, fontStyle: "italic", fontWeight: 500, color: textColor, lineHeight: 1.45, marginBottom: 32 }} />
          )}
          {at && (
            <Editable slideId={slide.id} block={at} index={slide.blocks.indexOf(at)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${bodySize}px`, color: primary, fontWeight: 700, letterSpacing: "0.08em", textTransform: "uppercase" }} />
          )}
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── KPI CARDS ─────────────────────────────────────────────────────────────
  if (layoutId === LAYOUT_KPI_CARDS) {
    const kpis = kpiBlocks.length > 0 ? kpiBlocks : bullets;
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "5px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", bottom: -80, right: -80, width: 300, height: 300, borderRadius: "50%", background: secondary, opacity: 0.06 }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: "44px 52px 40px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 6}px`, fontWeight: 700, color: textColor, marginBottom: 8, lineHeight: 1.2 }} />
          )}
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${bodySize}px`, color: mutedColor, marginBottom: 24 }} />
          )}
          {!subtitle && <div style={{ height: 16 }} />}
          <div style={{ display: "grid", gridTemplateColumns: `repeat(${Math.min(kpis.length, 4)}, 1fr)`, gap: 16, flex: 1, alignItems: "stretch" }}>
            {kpis.map((block, i) => {
              const parts = block.text.split("|");
              const label = parts[0] || block.text;
              const value = parts[1] || "";
              const delta = parts[2] || "";
              const accent = i % 2 === 0 ? primary : secondary;
              return (
                <div key={block.id || i} style={{ background: cardBg, borderRadius: 12, padding: "20px 18px", borderTop: `4px solid ${accent}`, display: "flex", flexDirection: "column", justifyContent: "space-between" }}>
                  <div style={{ fontSize: `${bodySize - 1}px`, color: mutedColor, fontWeight: 600, letterSpacing: "0.04em", textTransform: "uppercase", marginBottom: 8 }}>{label}</div>
                  <div style={{ fontSize: "38px", fontWeight: 800, color: accent, fontFamily: headingFont, lineHeight: 1.1 }}>{value}</div>
                  {delta && <div style={{ fontSize: `${bodySize - 1}px`, color: delta.startsWith("+") || delta.startsWith("↑") ? "#10B981" : delta.startsWith("-") || delta.startsWith("↓") ? "#EF4444" : mutedColor, fontWeight: 700, marginTop: 6 }}>{delta}</div>}
                </div>
              );
            })}
          </div>
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── CLOSING ───────────────────────────────────────────────────────────────
  if (layoutId === LAYOUT_CLOSING) {
    return (
      <div className={className} style={{ ...wrap, background: `linear-gradient(135deg, ${primary} 0%, ${secondary} 100%)` }}>
        <div style={{ position: "absolute", top: -120, left: -120, width: 400, height: 400, borderRadius: "50%", background: "rgba(255,255,255,0.06)" }} />
        <div style={{ position: "absolute", bottom: -80, right: -80, width: 320, height: 320, borderRadius: "50%", background: "rgba(255,255,255,0.08)" }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", alignItems: "center", textAlign: "center", padding: "72px 96px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize + 8}px`, fontWeight: 800, color: "#FFFFFF", marginBottom: 16, lineHeight: 1.1 }} />
          )}
          <div style={{ width: 80, height: 4, background: "rgba(255,255,255,0.5)", borderRadius: 2, marginBottom: 20 }} />
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${subSize}px`, color: "rgba(255,255,255,0.75)", lineHeight: 1.5 }} />
          )}
        </div>
      </div>
    );
  }

  // ── DEFAULT: BULLET LIST ──────────────────────────────────────────────────
  return (
    <div className={className} style={wrap}>
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "5px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
      <div style={{ position: "absolute", top: -60, right: -60, width: 280, height: 280, borderRadius: "50%", background: primary, opacity: 0.04 }} />

      <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: "48px 60px 40px" }}>
        {title && (
          <div style={{ marginBottom: 10 }}>
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 4}px`, fontWeight: 700, color: textColor, lineHeight: 1.2, letterSpacing: "-0.01em" }} />
            <div style={{ width: 44, height: 3, background: primary, borderRadius: 2, marginTop: 10 }} />
          </div>
        )}
        {subtitle && (
          <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
            style={{ fontSize: `${bodySize}px`, color: mutedColor, marginBottom: 16, marginTop: 4 }} />
        )}

        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 12, overflow: "hidden", marginTop: subtitle ? 0 : 8 }}>
          {bullets.map((block, i) => (
            <BulletItem
              key={block.id || i}
              slideId={slide.id}
              block={block}
              index={slide.blocks.indexOf(block)}
              primary={primary}
              secondary={secondary}
              textColor={textColor}
              mutedColor={mutedColor}
              bodySize={bodySize}
              isEditable={isEditable}
              update={updateBlockText}
              bulletIndex={i}
            />
          ))}
        </div>
      </div>

      <SlideNum n={slide.index} color={mutedColor} />
    </div>
  );
}

// ─── sub-components ───────────────────────────────────────────────────────────

interface EditableProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  isEditable: boolean;
  style: React.CSSProperties;
  update: (sid: string, bid: string, bi: number, t: string) => void;
}

function Editable({ slideId, block, index, isEditable, style, update }: EditableProps) {
  const onBlur = (e: FocusEvent<HTMLDivElement>) => {
    if (!isEditable) return;
    const t = e.currentTarget.innerText;
    if (t !== block.text) update(slideId, block.id, index, t);
  };
  const onKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); e.currentTarget.blur(); }
  };
  return (
    <div
      contentEditable={isEditable ? "plaintext-only" : false}
      suppressContentEditableWarning
      onBlur={onBlur}
      onKeyDown={onKeyDown}
      style={{ ...style, outline: "none", whiteSpace: "pre-wrap" }}
    >
      {block.text}
    </div>
  );
}

interface BulletItemProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  primary: string;
  secondary: string;
  textColor: string;
  mutedColor: string;
  bodySize: number;
  isEditable: boolean;
  bulletIndex: number;
  update: (sid: string, bid: string, bi: number, t: string) => void;
}

function BulletItem({ slideId, block, index, primary, secondary, textColor, mutedColor, bodySize, isEditable, bulletIndex, update }: BulletItemProps) {
  const accent = bulletIndex % 3 === 0 ? primary : bulletIndex % 3 === 1 ? secondary : primary;
  return (
    <div style={{ display: "flex", alignItems: "flex-start", gap: 14 }}>
      <div style={{ width: 8, height: 8, borderRadius: "50%", background: accent, marginTop: (bodySize * 1.65 - 8) / 2, flexShrink: 0 }} />
      <Editable
        slideId={slideId}
        block={block}
        index={index}
        isEditable={isEditable}
        update={update}
        style={{ fontSize: `${bodySize + 1}px`, color: textColor, lineHeight: 1.65, flex: 1 }}
      />
    </div>
  );
}

interface TwoColPanelProps {
  blocks: ContentBlock[];
  slide: SlideNode;
  accentColor: string;
  cardBg: string;
  textColor: string;
  bodySize: number;
  isEditable: boolean;
  update: (sid: string, bid: string, bi: number, t: string) => void;
}

function TwoColPanel({ blocks, slide, accentColor, cardBg, textColor, bodySize, isEditable, update }: TwoColPanelProps) {
  return (
    <div style={{ background: cardBg, borderRadius: 12, padding: "20px 24px", borderTop: `4px solid ${accentColor}`, overflow: "hidden" }}>
      {blocks.map((b, i) => {
        const lines = b.text.split("\n").filter(Boolean);
        if (lines.length > 1) {
          return (
            <div key={b.id || i} style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {lines.map((line, li) => (
                <div key={li} style={{ display: "flex", alignItems: "flex-start", gap: 10 }}>
                  <div style={{ width: 6, height: 6, borderRadius: "50%", background: accentColor, marginTop: 7, flexShrink: 0 }} />
                  <div style={{ fontSize: `${bodySize}px`, color: textColor, lineHeight: 1.6 }}>{line}</div>
                </div>
              ))}
            </div>
          );
        }
        return (
          <div key={b.id || i} style={{ display: "flex", alignItems: "flex-start", gap: 10, marginBottom: i < blocks.length - 1 ? 10 : 0 }}>
            <div style={{ width: 6, height: 6, borderRadius: "50%", background: accentColor, marginTop: 7, flexShrink: 0 }} />
            <Editable slideId={slide.id} block={b} index={slide.blocks.indexOf(b)} isEditable={isEditable} update={update}
              style={{ fontSize: `${bodySize}px`, color: textColor, lineHeight: 1.6, flex: 1 }} />
          </div>
        );
      })}
    </div>
  );
}

function SlideNum({ n, color }: { n: number; color: string }) {
  if (!n) return null;
  return (
    <div style={{ position: "absolute", bottom: 16, right: 22, fontSize: "12px", color, fontWeight: 500, letterSpacing: "0.04em" }}>
      {n}
    </div>
  );
}
