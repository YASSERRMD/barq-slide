"use client";

import { useMemo, FocusEvent, KeyboardEvent } from "react";
import type { SlideNode, ContentBlock, DesignTokens } from "@/lib/store/deck-store";
import { useDeckStore } from "@/lib/store/deck-store";

interface SlideCanvasProps {
  slide: SlideNode;
  tokens?: DesignTokens | null;
  scale?: number;
  className?: string;
  isEditable?: boolean;
}

// Hex → {r,g,b}
function hexToRgb(hex: string) {
  const m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return m ? { r: parseInt(m[1], 16), g: parseInt(m[2], 16), b: parseInt(m[3], 16) } : null;
}

// Perceived brightness 0–255
function brightness(hex: string) {
  const c = hexToRgb(hex);
  return c ? (c.r * 299 + c.g * 587 + c.b * 114) / 1000 : 128;
}

export function SlideCanvas({ slide, tokens, scale = 1, className = "", isEditable = false }: SlideCanvasProps) {
  const layoutId = slide.layout?.layoutId ?? 1;
  const primary   = tokens?.primaryHex      || "#4F46E5";
  const secondary = tokens?.secondaryHex    || "#10B981";
  const bg        = slide.bgColorHex || tokens?.backgroundHex || "#FFFFFF";
  const onBg      = tokens?.onBackgroundHex || "#111827";
  const muted     = tokens?.mutedHex        || "#6B7280";
  const surface   = tokens?.surfaceHex      || "#F9FAFB";

  const isDark = brightness(bg) < 128;
  const textColor   = isDark ? "#FFFFFF" : onBg;
  const mutedColor  = isDark ? "rgba(255,255,255,0.6)" : muted;
  const surfaceBg   = isDark ? "rgba(255,255,255,0.07)" : surface;

  // Group blocks by role
  const title    = slide.blocks.find(b => b.role === "title"    || b.role === "heading");
  const subtitle = slide.blocks.find(b => b.role === "subtitle" || b.role === "subheading");
  const bodies   = slide.blocks.filter(b =>
    b.role !== "title" && b.role !== "heading" &&
    b.role !== "subtitle" && b.role !== "subheading"
  );

  const s = scale;

  // Always render at 960×540. The parent wrapper is responsible for
  // providing a clipping container sized to 960*scale × 540*scale.
  const wrapperStyle: React.CSSProperties = {
    position: "relative",
    width: "960px",
    height: "540px",
    backgroundColor: bg,
    color: textColor,
    borderRadius: `${8 * s}px`,
    overflow: "hidden",
    fontFamily: tokens?.body?.fontFamily || "Inter, sans-serif",
    boxShadow: "0 2px 8px rgba(0,0,0,0.10), 0 8px 32px rgba(0,0,0,0.07)",
    transform: `scale(${s})`,
    transformOrigin: "top left",
    flexShrink: 0,
  };

  // Title / cover layout (1, 9)
  if (layoutId === 1 || layoutId === 9) {
    return (
      <div className={className} style={wrapperStyle}>
        {/* Decorative accent bar top */}
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: `${6 * s}px`, background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        {/* Accent shape bottom-right */}
        <div style={{ position: "absolute", bottom: 0, right: 0, width: `${220 * s}px`, height: `${220 * s}px`, borderRadius: "50%", background: primary, opacity: 0.08, transform: "translate(30%, 30%)" }} />
        <div style={{ position: "absolute", bottom: 0, right: 0, width: `${140 * s}px`, height: `${140 * s}px`, borderRadius: "50%", background: secondary, opacity: 0.10, transform: "translate(10%, 10%)" }} />

        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: `${72 * s}px ${80 * s}px` }}>
          {title && (
            <EditableBlock slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{
                fontSize: `${(tokens?.heading?.sizePt || 44) * s}px`,
                fontWeight: 800,
                lineHeight: 1.1,
                letterSpacing: "-0.025em",
                color: textColor,
                marginBottom: `${16 * s}px`,
                maxWidth: `${640 * s}px`,
              }}
            />
          )}
          {/* Accent divider */}
          <div style={{ width: `${56 * s}px`, height: `${4 * s}px`, background: primary, borderRadius: `${2 * s}px`, marginBottom: `${20 * s}px` }} />
          {subtitle && (
            <EditableBlock slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{
                fontSize: `${(tokens?.subheading?.sizePt || 20) * s}px`,
                fontWeight: 400,
                color: mutedColor,
                lineHeight: 1.5,
                maxWidth: `${520 * s}px`,
              }}
            />
          )}
          {bodies.map((b, i) => (
            <EditableBlock key={b.id || i} slideId={slide.id} block={b} index={slide.blocks.indexOf(b)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.body?.sizePt || 14) * s}px`, color: mutedColor, marginTop: `${8 * s}px` }}
            />
          ))}
        </div>

        {/* Slide number */}
        <SlideNumber index={slide.index} s={s} color={mutedColor} />
      </div>
    );
  }

  // Two-column layout (3, 12, comparison)
  if (layoutId === 3 || layoutId === 12) {
    const half = Math.floor(bodies.length / 2);
    const left  = bodies.slice(0, Math.max(1, half));
    const right = bodies.slice(Math.max(1, half));
    return (
      <div className={className} style={wrapperStyle}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: `${5 * s}px`, background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: `${44 * s}px ${52 * s}px ${36 * s}px` }}>
          {/* Header */}
          {title && (
            <EditableBlock slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.heading?.sizePt || 28) * s}px`, fontWeight: 700, color: textColor, marginBottom: `${6 * s}px`, lineHeight: 1.2 }}
            />
          )}
          {subtitle && (
            <EditableBlock slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.body?.sizePt || 13) * s}px`, color: mutedColor, marginBottom: `${20 * s}px` }}
            />
          )}
          {/* Two columns */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: `${24 * s}px`, flex: 1 }}>
            <div style={{ background: surfaceBg, borderRadius: `${10 * s}px`, padding: `${20 * s}px`, borderTop: `${3 * s}px solid ${primary}` }}>
              {left.map((b, i) => <EditableBlock key={b.id || i} slideId={slide.id} block={b} index={slide.blocks.indexOf(b)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
                style={{ fontSize: `${(tokens?.body?.sizePt || 13) * s}px`, color: textColor, lineHeight: 1.6 }} />)}
            </div>
            <div style={{ background: surfaceBg, borderRadius: `${10 * s}px`, padding: `${20 * s}px`, borderTop: `${3 * s}px solid ${secondary}` }}>
              {right.map((b, i) => <EditableBlock key={b.id || i} slideId={slide.id} block={b} index={slide.blocks.indexOf(b)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
                style={{ fontSize: `${(tokens?.body?.sizePt || 13) * s}px`, color: textColor, lineHeight: 1.6 }} />)}
            </div>
          </div>
        </div>
        <SlideNumber index={slide.index} s={s} color={mutedColor} />
      </div>
    );
  }

  // Quote layout (11)
  if (layoutId === 11) {
    const quote = title || bodies[0];
    const attr  = subtitle || bodies[1];
    return (
      <div className={className} style={wrapperStyle}>
        {/* Left accent stripe */}
        <div style={{ position: "absolute", left: 0, top: 0, bottom: 0, width: `${8 * s}px`, background: `linear-gradient(180deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", justifyContent: "center", padding: `${60 * s}px ${80 * s}px` }}>
          {/* Quote mark */}
          <div style={{ fontSize: `${120 * s}px`, lineHeight: 0.8, color: primary, opacity: 0.2, fontFamily: "Georgia, serif", marginBottom: `${16 * s}px`, userSelect: "none" }}>"</div>
          {quote && (
            <EditableBlock slideId={slide.id} block={quote} index={slide.blocks.indexOf(quote)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.heading?.sizePt || 28) * s}px`, fontWeight: 600, fontStyle: "italic", color: textColor, lineHeight: 1.4, marginBottom: `${24 * s}px` }}
            />
          )}
          {attr && (
            <EditableBlock slideId={slide.id} block={attr} index={slide.blocks.indexOf(attr)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.body?.sizePt || 14) * s}px`, color: primary, fontWeight: 600, letterSpacing: "0.05em", textTransform: "uppercase" }}
            />
          )}
        </div>
        <SlideNumber index={slide.index} s={s} color={mutedColor} />
      </div>
    );
  }

  // Big number / KPI (10, 15)
  if (layoutId === 10 || layoutId === 15) {
    const kpis = bodies.length > 0 ? bodies : (title ? [title] : []);
    return (
      <div className={className} style={wrapperStyle}>
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: `${5 * s}px`, background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
        <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: `${44 * s}px ${52 * s}px` }}>
          {title && layoutId === 15 && (
            <EditableBlock slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.heading?.sizePt || 28) * s}px`, fontWeight: 700, color: textColor, marginBottom: `${24 * s}px` }}
            />
          )}
          <div style={{ display: "flex", gap: `${20 * s}px`, flex: 1, alignItems: "center" }}>
            {kpis.map((b, i) => (
              <div key={b.id || i} style={{ flex: 1, background: surfaceBg, borderRadius: `${12 * s}px`, padding: `${24 * s}px`, textAlign: "center", borderBottom: `${3 * s}px solid ${i % 2 === 0 ? primary : secondary}` }}>
                <div style={{ fontSize: `${48 * s}px`, fontWeight: 800, color: i % 2 === 0 ? primary : secondary, lineHeight: 1 }}>
                  <EditableBlock slideId={slide.id} block={b} index={slide.blocks.indexOf(b)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText} style={{}} />
                </div>
              </div>
            ))}
          </div>
          {subtitle && (
            <EditableBlock slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{ fontSize: `${(tokens?.body?.sizePt || 13) * s}px`, color: mutedColor, marginTop: `${12 * s}px` }}
            />
          )}
        </div>
        <SlideNumber index={slide.index} s={s} color={mutedColor} />
      </div>
    );
  }

  // Default: content slide (title + bullet list or body)
  return (
    <div className={className} style={wrapperStyle}>
      {/* Top accent bar */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: `${5 * s}px`, background: `linear-gradient(90deg, ${primary}, ${secondary})` }} />
      {/* Subtle background shape */}
      <div style={{ position: "absolute", top: 0, right: 0, width: `${300 * s}px`, height: `${300 * s}px`, borderRadius: "50%", background: primary, opacity: 0.04, transform: "translate(40%, -40%)" }} />

      <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", padding: `${48 * s}px ${56 * s}px ${40 * s}px` }}>
        {/* Title area */}
        {title && (
          <div style={{ marginBottom: `${4 * s}px` }}>
            <EditableBlock slideId={slide.id} block={title} index={slide.blocks.indexOf(title)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
              style={{
                fontSize: `${(tokens?.heading?.sizePt || 30) * s}px`,
                fontWeight: 700,
                color: textColor,
                lineHeight: 1.2,
                letterSpacing: "-0.01em",
              }}
            />
            {/* Underline accent */}
            <div style={{ width: `${40 * s}px`, height: `${3 * s}px`, background: primary, borderRadius: `${2 * s}px`, marginTop: `${8 * s}px` }} />
          </div>
        )}
        {subtitle && (
          <EditableBlock slideId={slide.id} block={subtitle} index={slide.blocks.indexOf(subtitle)} isEditable={isEditable} updateBlockText={useDeckStore.getState().updateBlockText}
            style={{ fontSize: `${(tokens?.body?.sizePt || 14) * s}px`, color: mutedColor, marginBottom: `${20 * s}px`, marginTop: `${4 * s}px` }}
          />
        )}

        {/* Body blocks */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: `${10 * s}px`, overflow: "hidden" }}>
          {bodies.map((block, i) => (
            <BodyBlock
              key={block.id || i}
              slideId={slide.id}
              block={block}
              index={slide.blocks.indexOf(block)}
              tokens={tokens}
              scale={s}
              textColor={textColor}
              mutedColor={mutedColor}
              primary={primary}
              surfaceBg={surfaceBg}
              isEditable={isEditable}
            />
          ))}
        </div>
      </div>

      <SlideNumber index={slide.index} s={s} color={mutedColor} />
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Sub-components
// ─────────────────────────────────────────────────────────────────────────────

interface EditableBlockProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  isEditable: boolean;
  style: React.CSSProperties;
  updateBlockText: (slideId: string, blockId: string, blockIndex: number, text: string) => void;
}

function EditableBlock({ slideId, block, index, isEditable, style, updateBlockText }: EditableBlockProps) {
  const handleBlur = (e: FocusEvent<HTMLDivElement>) => {
    if (!isEditable) return;
    const t = e.currentTarget.innerText;
    if (t !== block.text) updateBlockText(slideId, block.id, index, t);
  };
  const handleKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); e.currentTarget.blur(); }
  };
  return (
    <div
      contentEditable={isEditable ? "plaintext-only" : false}
      suppressContentEditableWarning
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      style={{ ...style, outline: "none", whiteSpace: "pre-wrap" }}
      className={isEditable ? "focus:ring-2 focus:ring-primary/20 rounded-sm cursor-text" : ""}
    >
      {block.text}
    </div>
  );
}

interface BodyBlockProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  tokens?: DesignTokens | null;
  scale: number;
  textColor: string;
  mutedColor: string;
  primary: string;
  surfaceBg: string;
  isEditable: boolean;
}

function BodyBlock({ slideId, block, index, tokens, scale: s, textColor, mutedColor, primary, surfaceBg, isEditable }: BodyBlockProps) {
  const updateBlockText = useDeckStore((state) => state.updateBlockText);
  const role = block.role || "body";
  const fontSize = `${(tokens?.body?.sizePt || 14) * s}px`;

  if (role === "bullet") {
    const lines = block.text.split("\n").filter(Boolean);
    return (
      <div style={{ display: "flex", flexDirection: "column", gap: `${6 * s}px` }}>
        {lines.map((line, li) => (
          <div key={li} style={{ display: "flex", alignItems: "flex-start", gap: `${10 * s}px` }}>
            <div style={{ width: `${6 * s}px`, height: `${6 * s}px`, borderRadius: "50%", background: primary, marginTop: `${6 * s}px`, flexShrink: 0 }} />
            <div
              contentEditable={isEditable ? "plaintext-only" : false}
              suppressContentEditableWarning
              style={{ fontSize, color: textColor, lineHeight: 1.6, flex: 1, outline: "none" }}
              onBlur={(e) => {
                if (!isEditable) return;
                const newLines = [...block.text.split("\n")];
                newLines[li] = e.currentTarget.innerText.trim();
                updateBlockText(slideId, block.id, index, newLines.join("\n"));
              }}
            >
              {line}
            </div>
          </div>
        ))}
      </div>
    );
  }

  // Callout / emphasis block
  if (block.emphasis > 1) {
    return (
      <div style={{ background: surfaceBg, borderLeft: `${3 * s}px solid ${primary}`, padding: `${12 * s}px ${16 * s}px`, borderRadius: `${0}px ${6 * s}px ${6 * s}px ${0}px` }}>
        <EditableBlock slideId={slideId} block={block} index={index} isEditable={isEditable} updateBlockText={updateBlockText}
          style={{ fontSize, color: primary, fontWeight: 600, lineHeight: 1.5 }} />
      </div>
    );
  }

  return (
    <EditableBlock slideId={slideId} block={block} index={index} isEditable={isEditable} updateBlockText={updateBlockText}
      style={{ fontSize, color: mutedColor, lineHeight: 1.7 }} />
  );
}

function SlideNumber({ index, s, color }: { index: number; s: number; color: string }) {
  if (!index) return null;
  return (
    <div style={{
      position: "absolute", bottom: `${16 * s}px`, right: `${20 * s}px`,
      fontSize: `${11 * s}px`, color, fontWeight: 500, letterSpacing: "0.05em",
      fontVariantNumeric: "tabular-nums",
    }}>
      {index}
    </div>
  );
}
