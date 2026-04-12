"use client";

import { useMemo, FocusEvent, KeyboardEvent } from "react";
import type { SlideNode, ContentBlock, DesignTokens } from "@/lib/store/deck-store";
import { useDeckStore } from "@/lib/store/deck-store";

/**
 * Maps SlideLayout layoutId to CSS grid definitions.
 * The enum values mirror the Go layout package.
 */
const GRID_LAYOUTS: Record<number, { columns: string; areas: string }> = {
  // TITLE_CENTER
  1: { columns: "1fr", areas: '"title" "subtitle"' },
  // TITLE_LEFT
  2: { columns: "1fr 1fr", areas: '"title media"' },
  // TWO_COLUMN
  3: { columns: "1fr 1fr", areas: '"left right"' },
  // THREE_COLUMN
  4: { columns: "1fr 1fr 1fr", areas: '"col1 col2 col3"' },
  // MEDIA_LEFT
  5: { columns: "2fr 3fr", areas: '"media content"' },
  // MEDIA_RIGHT
  6: { columns: "3fr 2fr", areas: '"content media"' },
  // FULL_BLEED
  7: { columns: "1fr", areas: '"content"' },
  // BLANK
  8: { columns: "1fr", areas: '"content"' },
  // SECTION_HEADER
  9: { columns: "1fr", areas: '"title" "divider" "subtitle"' },
  // BIG_NUMBER
  10: { columns: "1fr", areas: '"number" "caption"' },
  // QUOTE
  11: { columns: "1fr", areas: '"quote" "attribution"' },
  // COMPARISON
  12: { columns: "1fr 1fr", areas: '"left right"' },
  // TIMELINE
  13: { columns: "repeat(4, 1fr)", areas: '"t1 t2 t3 t4"' },
  // ICON_GRID
  14: { columns: "repeat(3, 1fr)", areas: '"i1 i2 i3" "i4 i5 i6"' },
  // KPI_ROW
  15: { columns: "repeat(4, 1fr)", areas: '"k1 k2 k3 k4"' },
};

const DEFAULT_GRID = { columns: "1fr", areas: '"content"' };

interface SlideCanvasProps {
  slide: SlideNode;
  tokens?: DesignTokens | null;
  scale?: number;
  className?: string;
  /** If true, blocks can be edited inline */
  isEditable?: boolean;
}

/**
 * SlideCanvas renders a single slide as a responsive CSS Grid layout
 * using the DesignTokens as CSS custom properties.
 */
export function SlideCanvas({ slide, tokens, scale = 1, className = "", isEditable = false }: SlideCanvasProps) {
  const layoutId = slide.layout?.layoutId ?? 1;
  const grid = GRID_LAYOUTS[layoutId] ?? DEFAULT_GRID;

  const cssVars = useMemo(() => {
    if (!tokens) return {};
    return {
      "--slide-primary": tokens.primaryHex || "#4F46E5",
      "--slide-secondary": tokens.secondaryHex || "#10B981",
      "--slide-bg": tokens.backgroundHex || "#FFFFFF",
      "--slide-surface": tokens.surfaceHex || "#F9FAFB",
      "--slide-on-bg": tokens.onBackgroundHex || "#111827",
      "--slide-muted": tokens.mutedHex || "#6B7280",
      "--slide-heading-font": tokens.heading?.fontFamily || "Inter",
      "--slide-heading-size": `${tokens.heading?.sizePt || 36}px`,
      "--slide-body-font": tokens.body?.fontFamily || "Inter",
      "--slide-body-size": `${tokens.body?.sizePt || 14}px`,
    } as React.CSSProperties;
  }, [tokens]);

  const bgColor = slide.bgColorHex || tokens?.backgroundHex || "#FFFFFF";
  const isDark = slide.layout?.isDarkBackground ?? false;
  const textColor = isDark ? "#FFFFFF" : (tokens?.onBackgroundHex || "#111827");

  return (
    <div
      className={`relative overflow-hidden ${className}`}
      style={{
        ...cssVars,
        width: `${960 * scale}px`,
        height: `${540 * scale}px`,
        backgroundColor: bgColor,
        color: textColor,
        borderRadius: `${8 * scale}px`,
        boxShadow: "0 1px 3px rgba(0,0,0,0.08), 0 4px 12px rgba(0,0,0,0.04)",
        transform: `scale(${scale})`,
        transformOrigin: "top left",
      }}
      id={`slide-canvas-${slide.id}`}
    >
      <div
        style={{
          display: "grid",
          gridTemplateColumns: grid.columns,
          gridTemplateAreas: grid.areas,
          gap: `${16 * scale}px`,
          padding: `${40 * scale}px`,
          height: "100%",
          alignItems: "center",
        }}
      >
        {slide.blocks.map((block, i) => (
          <SlideBlock
            key={block.id || i}
            slideId={slide.id}
            block={block}
            index={i}
            tokens={tokens}
            scale={scale}
            isDark={isDark}
            isEditable={isEditable}
          />
        ))}
      </div>
    </div>
  );
}

interface SlideBlockProps {
  slideId: string;
  block: ContentBlock;
  index: number;
  tokens?: DesignTokens | null;
  scale: number;
  isDark: boolean;
  isEditable: boolean;
}

function SlideBlock({ slideId, block, index, tokens, scale, isDark, isEditable }: SlideBlockProps) {
  const role = block.role || "body";
  const updateBlockText = useDeckStore((s) => s.updateBlockText);

  const style: React.CSSProperties = {};

  switch (role) {
    case "heading":
    case "title":
      style.fontFamily = tokens?.heading?.fontFamily || "Inter";
      style.fontSize = `${(tokens?.heading?.sizePt || 36) * scale}px`;
      style.fontWeight = 700;
      style.lineHeight = 1.1;
      style.letterSpacing = "-0.02em";
      break;
    case "subheading":
    case "subtitle":
      style.fontFamily = tokens?.subheading?.fontFamily || "Inter";
      style.fontSize = `${(tokens?.subheading?.sizePt || 20) * scale}px`;
      style.fontWeight = 500;
      style.opacity = 0.8;
      break;
    case "body":
    case "bullet":
    default:
      style.fontFamily = tokens?.body?.fontFamily || "Inter";
      style.fontSize = `${(tokens?.body?.sizePt || 14) * scale}px`;
      style.fontWeight = 400;
      style.lineHeight = 1.6;
      break;
  }

  // Emphasis styling.
  if (block.emphasis > 1) {
    const primary = tokens?.primaryHex || "#4F46E5";
    style.color = primary;
  }

  const handleBlur = (e: FocusEvent<HTMLDivElement>) => {
    if (!isEditable) return;
    const newText = e.currentTarget.innerText;
    if (newText !== block.text) {
      updateBlockText(slideId, block.id, index, newText);
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (!isEditable) return;
    // Prevent adding complex formatting or bolding via shortcuts if desired
    // Currently relying on plaintext processing when converting to PPTX
    if (e.key === 'Enter' && !e.shiftKey) {
      if (role !== "bullet") {
        e.preventDefault();
        e.currentTarget.blur();
      }
    }
  };

  return (
    <div
      className={`min-w-0 break-words ${isEditable ? "focus:outline-none focus:ring-2 focus:ring-primary/20 rounded-sm -m-1 p-1 transition-shadow cursor-text" : ""}`}
      style={style}
      id={`slide-block-${block.id || index}`}
    >
      {role === "bullet" ? (
        <ul style={{ paddingLeft: `${16 * scale}px`, margin: 0 }}>
          {block.text.split("\n").map((line, li) => (
            <li key={li} style={{ marginBottom: `${4 * scale}px` }}>
              <div
                contentEditable={isEditable ? "plaintext-only" : false}
                suppressContentEditableWarning
                onBlur={(e) => {
                  if (!isEditable) return;
                  const newLines = [...block.text.split("\n")];
                  newLines[li] = e.currentTarget.innerText.trim();
                  updateBlockText(slideId, block.id, index, newLines.join("\n"));
                }}
              >
                {line}
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <div
          contentEditable={isEditable ? "plaintext-only" : false}
          suppressContentEditableWarning
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
          style={{ whiteSpace: "pre-wrap" }}
        >
          {block.text}
        </div>
      )}
    </div>
  );
}
