import { marked } from 'marked';
import katex from 'katex';
import 'katex/dist/katex.min.css';

export interface RenderedGuideBlock {
  startLine: number;
  endLine: number;
  rawText: string;
  html: string;
}

marked.setOptions({
  gfm: true,
  breaks: true
});

/**
 * Renders Markdown and LaTeX ($...$, $$...$$, \(...\), \[...\]) into safe HTML
 * while protecting fenced and inline code spans from LaTeX delimiter collisions.
 */
export function renderMarkdownWithLatex(source: string, inlineOnly = false): string {
  if (!source || !source.trim()) {
    return '';
  }

  const codePlaceholders: string[] = [];
  const mathPlaceholders: string[] = [];

  // 1. Protect fenced code blocks (```...```) and inline code (`...`) first
  let text = source.replace(/```[\s\S]*?```|`[^`\n]+`/g, (match) => {
    const idx = codePlaceholders.length;
    codePlaceholders.push(match);
    return `@@QST_CODE_${idx}@@`;
  });

  // 2. Extract and render display math: $$...$$ and \[...\]
  text = text.replace(/\$\$([\s\S]+?)\$\$|\\\[([\s\S]+?)\\\]/g, (_, d1, d2) => {
    const expr = (d1 ?? d2 ?? '').trim();
    const idx = mathPlaceholders.length;
    try {
      const html = katex.renderToString(expr, {
        displayMode: true,
        throwOnError: false,
        output: 'html'
      });
      mathPlaceholders.push(`<div class="qst-math-block">${html}</div>`);
    } catch {
      mathPlaceholders.push(`<code>$$${expr}$$</code>`);
    }
    return `@@QST_MATH_${idx}@@`;
  });

  // 3. Extract and render inline math: $...$ and \(...\)
  text = text.replace(/\\\(([\s\S]+?)\\\)|(?<!\$)\$(?!\$)([^\$\n]+?)(?<!\$)\$(?!\$)/g, (_, p1, p2) => {
    const expr = (p1 ?? p2 ?? '').trim();
    if (!expr) return _;
    const idx = mathPlaceholders.length;
    try {
      const html = katex.renderToString(expr, {
        displayMode: false,
        throwOnError: false,
        output: 'html'
      });
      mathPlaceholders.push(`<span class="qst-math-inline">${html}</span>`);
    } catch {
      mathPlaceholders.push(`<code>$${expr}$</code>`);
    }
    return `@@QST_MATH_${idx}@@`;
  });

  // 4. Restore code spans before Markdown parsing so Marked formats them as <pre><code> / <code>
  text = text.replace(/@@QST_CODE_(\d+)@@/g, (_, idxStr) => {
    return codePlaceholders[Number(idxStr)] ?? '';
  });

  // 5. Parse Markdown with Marked
  let rendered = inlineOnly
    ? (marked.parseInline(text) as string)
    : (marked.parse(text) as string);

  // 6. Restore rendered KaTeX HTML
  rendered = rendered.replace(/@@QST_MATH_(\d+)@@/g, (_, idxStr) => {
    return mathPlaceholders[Number(idxStr)] ?? '';
  });

  return rendered;
}

/**
 * Splits a Markdown guide body into logical source-line-mapped blocks
 * (headings, fenced code blocks, math blocks, lists, tables, paragraphs)
 * so each rendered block preserves its exact 1-based [startLine, endLine] range.
 */
export function parseMarkdownBlocksWithLines(body: string): RenderedGuideBlock[] {
  if (!body) return [];
  const normalized = body.replace(/\r\n/g, '\n');
  const lines = normalized.split('\n');
  const blocks: RenderedGuideBlock[] = [];

  let i = 0;
  while (i < lines.length) {
    // Skip blank lines between blocks
    if (lines[i].trim() === '') {
      i++;
      continue;
    }

    const startLine = i + 1;
    const trimmed = lines[i].trim();

    // Case 1: Fenced code block (``` or ~~~)
    if (trimmed.startsWith('```') || trimmed.startsWith('~~~')) {
      const fence = trimmed.slice(0, 3);
      const buf = [lines[i]];
      i++;
      while (i < lines.length) {
        buf.push(lines[i]);
        if (lines[i].trim().startsWith(fence)) {
          i++;
          break;
        }
        i++;
      }
      const rawText = buf.join('\n');
      blocks.push({
        startLine,
        endLine: i,
        rawText,
        html: renderMarkdownWithLatex(rawText, false)
      });
      continue;
    }

    // Case 2: Multi-line display math block ($$ or \[)
    if (
      (trimmed.startsWith('$$') && (trimmed === '$$' || !trimmed.endsWith('$$'))) ||
      (trimmed.startsWith('\\[') && !trimmed.endsWith('\\]'))
    ) {
      const closeToken = trimmed.startsWith('$$') ? '$$' : '\\]';
      const buf = [lines[i]];
      i++;
      while (i < lines.length) {
        buf.push(lines[i]);
        if (lines[i].trim().endsWith(closeToken)) {
          i++;
          break;
        }
        i++;
      }
      const rawText = buf.join('\n');
      blocks.push({
        startLine,
        endLine: i,
        rawText,
        html: renderMarkdownWithLatex(rawText, false)
      });
      continue;
    }

    // Case 3: Single-line ATX Heading (# ... )
    if (/^#{1,6}\s+/.test(trimmed)) {
      const rawText = lines[i];
      i++;
      blocks.push({
        startLine,
        endLine: startLine,
        rawText,
        html: renderMarkdownWithLatex(rawText, false)
      });
      continue;
    }

    // Case 4: Contiguous paragraph, list, blockquote, or table until next blank line or fence/heading
    const buf: string[] = [];
    while (i < lines.length && lines[i].trim() !== '') {
      const t = lines[i].trim();
      if (buf.length > 0 && (t.startsWith('```') || t.startsWith('~~~') || /^#{1,6}\s+/.test(t))) {
        break;
      }
      buf.push(lines[i]);
      i++;
    }
    const rawText = buf.join('\n');
    blocks.push({
      startLine,
      endLine: i,
      rawText,
      html: renderMarkdownWithLatex(rawText, false)
    });
  }

  return blocks;
}
