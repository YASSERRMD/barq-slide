"use client";

import { useMemo, FocusEvent, KeyboardEvent } from "react";
import type { SlideNode, ContentBlock, DesignTokens } from "@/lib/store/deck-store";
import { useDeckStore } from "@/lib/store/deck-store";

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
  const layoutId = slide.layout?.layoutId ?? 1;

  const primary   = tokens?.primaryHex      || "#4F46E5";
  const secondary = tokens?.secondaryHex    || "#10B981";
  const bg        = slide.bgColorHex || tokens?.backgroundHex || "#FFFFFF";
  const onBg      = tokens?.onBackgroundHex || "#111827";
  const muted     = tokens?.mutedHex        || "#6B7280";
  const surface   = tokens?.surfaceHex      || "#F9FAFB";

  const isDark    = hexBrightness(bg) < 128;
  const textColor = isDark ? "#FFFFFF" : onBg;
  const mutedColor = isDark ? "rgba(255,255,255,0.55)" : muted;
  const cardBg    = isDark ? "rgba(255,255,255,0.08)" : surface;

  const headingFont = tokens?.heading?.fontFamily || "Inter, sans-serif";
  const bodyFont    = tokens?.body?.fontFamily    || "Inter, sans-serif";
  const headingSize = tokens?.heading?.sizePt     || 36;
  const subSize     = tokens?.subheading?.sizePt  || 20;
  const bodySize    = tokens?.body?.sizePt        || 15;

  // Group blocks by role
  const title    = slide.blocks.find(b => b.role === "title"    || b.role === "heading");
  const subtitle = slide.blocks.find(b => b.role === "subtitle" || b.role === "subheading");
  const bodies   = slide.blocks.filter(b =>
    b.role !== "title" && b.role !== "heading" &&
    b.role !== "subtitle" && b.role !== "subheading"
  );

  // Wrapper: always 960×540; caller clips with overflow:hidden
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

  const updateBlockText = useDeckStore.getState().updateBlockText;

  // ── TITLE / COVER (1, 9) ──────────────────────────────────────────────────
  if (layoutId === 1 || layoutId === 9) {
    return (
      <div className={className} style={wrap}>
        {/* Gradient top bar */}
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "6px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        {/* Bottom-right decorative circles */}
        <div style={{ position: "absolute", bottom: -80, right: -80, width: 320, height: 320, borderRadius: "50%", background: primary, opacity: 0.07 }} />
        <div style={{ position: "absolute", bottom: -40, right: -40, width: 200, height: 200, borderRadius: "50%", background: secondary, opacity: 0.09 }} />

        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: "72px 88px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize + 8}px`, fontWeight: 800, lineHeight: 1.1, letterSpacing: "-0.025em", color: textColor, marginBottom: "16px" }} />
          )}
          <div style={{ width: 56, height: 4, background: primary, borderRadius: 2, marginBottom: 20 }} />
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${subSize}px`, fontWeight: 400, color: mutedColor, lineHeight: 1.5, maxWidth: 560 }} />
          )}
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── TWO-COLUMN (3, 12) ────────────────────────────────────────────────────
  if (layoutId === 3 || layoutId === 12) {
    const half  = Math.ceil(bodies.length / 2);
    const left  = bodies.slice(0, half);
    const right = bodies.slice(half);
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "5px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: "48px 56px 40px" }}>
          {title && (
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 4}px`, fontWeight: 700, color: textColor, marginBottom: 6, lineHeight: 1.2 }} />
          )}
          {subtitle && (
            <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${bodySize}px`, color: mutedColor, marginBottom: 24 }} />
          )}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, flex: 1 }}>
            <Column blocks={left} slide={slide} tokens={tokens} primary={primary} cardBg={cardBg} textColor={textColor} accentColor={primary} isEditable={isEditable} update={updateBlockText} bodySize={bodySize} />
            <Column blocks={right} slide={slide} tokens={tokens} primary={secondary} cardBg={cardBg} textColor={textColor} accentColor={secondary} isEditable={isEditable} update={updateBlockText} bodySize={bodySize} />
          </div>
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── QUOTE (11) ────────────────────────────────────────────────────────────
  if (layoutId === 11) {
    const quoteBlock = title || bodies[0];
    const attrBlock  = subtitle || bodies[1];
    return (
      <div className={className} style={wrap}>
        <div style={{ position: "absolute", left: 0, top: 0, bottom: 0, width: 8, background: `linear-gradient(180deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: "64px 80px" }}>
          <div style={{ fontSize: 120, lineHeight: 0.7, color: primary, opacity: 0.15, fontFamily: "Georgia, serif", marginBottom: 16, userSelect: "none" }}>"</div>
          {quoteBlock && (
            <Editable slideId={slide.id} block={quoteBlock} index={slide.blocks.indexOf(quoteBlock)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 6}px`, fontStyle: "italic", fontWeight: 500, color: textColor, lineHeight: 1.4, marginBottom: 28 }} />
          )}
          {attrBlock && (
            <Editable slideId={slide.id} block={attrBlock} index={slide.blocks.indexOf(attrBlock)} isEditable={isEditable} update={updateBlockText}
              style={{ fontSize: `${bodySize - 1}px`, color: primary, fontWeight: 700, letterSpacing: "0.08em", textTransform: "uppercase" }} />
          )}
        </div>
        <SlideNum n={slide.index} color={mutedColor} />
      </div>
    );
  }

  // ── DEFAULT: title + body list ────────────────────────────────────────────
  return (
    <div className={className} style={wrap}>
      {/* Top gradient bar */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: "5px", background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
      {/* Subtle decoration */}
      <div style={{ position: "absolute", top: -60, right: -60, width: 280, height: 280, borderRadius: "50%", background: primary, opacity: 0.04 }} />

      <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: "52px 64px 44px" }}>
        {/* Title + accent */}
        {title && (
          <div style={{ marginBottom: 8 }}>
            <Editable slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} update={updateBlockText}
              style={{ fontFamily: headingFont, fontSize: `${headingSize - 4}px`, fontWeight: 700, color: textColor, lineHeight: 1.2, letterSpacing: "-0.01em" }} />
            <div style={{ width: 40, height: 3, background: primary, borderRadius: 2, marginTop: 10 }} />
          </div>
        )}
        {subtitle && (
          <Editable slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} update={updateBlockText}
            style={{ fontSize: `${bodySize}px`, color: mutedColor, marginBottom: 20, marginTop: 4 }} />
        )}

        {/* Body content */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 10, overflow: "hidden" }}>
          {bodies.map((block, i) => (
            <BodyItem
              key={block.id || i}
              slideId={slide.id}
              block={block}
              index={slide.blocks.indexOf(block)}
              primary={primary}
              cardBg={cardBg}
              textColor={textColor}
              mutedColor={mutedColor}
              bodySize={bodySize}
              isEditable={isEditable}
              update={updateBlockText}
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

interface BodyItemProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  primary: string;
  cardBg: string;
  textColor: string;
  mutedColor: string;
  bodySize: number;
  isEditable: boolean;
  update: (sid: string, bid: string, bi: number, t: string) => void;
}

function BodyItem({ slideId, block, index, primary, cardBg, textColor, mutedColor, bodySize, isEditable, update }: BodyItemProps) {
  const role = block.role || "body";

  if (role === "bullet") {
    const lines = block.text.split("\n").filter(Boolean);
    return (
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {lines.map((line, li) => (
          <div key={li} style={{ display: "flex", alignItems: "flex-start", gap: 12 }}>
            <div style={{ width: 6, height: 6, borderRadius: "50%", background: primary, marginTop: 6, flexShrink: 0 }} />
            <div
              contentEditable={isEditable ? "plaintext-only" : false}
              suppressContentEditableWarning
              style={{ fontSize: `${bodySize}px`, color: textColor, lineHeight: 1.65, flex: 1, outline: "none" }}
              onBlur={(e) => {
                if (!isEditable) return;
                const lines2 = block.text.split("\n");
                lines2[li] = e.currentTarget.innerText.trim();
                update(slideId, block.id, index, lines2.join("\n"));
              }}
            >
              {line}
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (block.emphasis > 1) {
    return (
      <div style={{ background: cardBg, borderLeft: `3px solid ${primary}`, padding: "12px 16px", borderRadius: "0 6px 6px 0" }}>
        <Editable slideId={slideId} block={block} index={index} isEditable={isEditable} update={update}
          style={{ fontSize: `${bodySize}px`, color: primary, fontWeight: 600, lineHeight: 1.5 }} />
      </div>
    );
  }

  return (
    <Editable slideId={slideId} block={block} index={index} isEditable={isEditable} update={update}
      style={{ fontSize: `${bodySize}px`, color: mutedColor, lineHeight: 1.7 }} />
  );
}

interface ColumnProps {
  blocks: ContentBlock[];
  slide: SlideNode;
  tokens?: DesignTokens | null;
  primary: string;
  cardBg: string;
  textColor: string;
  accentColor: string;
  bodySize: number;
  isEditable: boolean;
  update: (sid: string, bid: string, bi: number, t: string) => void;
}

function Column({ blocks, slide, cardBg, textColor, accentColor, bodySize, isEditable, update }: ColumnProps) {
  return (
    <div style={{ background: cardBg, borderRadius: 10, padding: "20px 24px", borderTop: `3px solid ${accentColor}` }}>
      {blocks.map((b, i) => (
        <BodyItem key={b.id || i} slideId={slide.id} block={b} index={slide.blocks.indexOf(b)}
          primary={accentColor} cardBg={cardBg} textColor={textColor} mutedColor={textColor}
          bodySize={bodySize} isEditable={isEditable} update={update} />
      ))}
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
